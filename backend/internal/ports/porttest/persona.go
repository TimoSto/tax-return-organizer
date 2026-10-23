package porttest

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

// PersonaRepositoryFactory returns an empty repository for the given test.
// Implementations that enforce referential integrity must make sure the given collectors exist.
type PersonaRepositoryFactory func(t *testing.T, collectors ...domain.CollectorID) ports.PersonaRepository

// TestPersonaRepository runs the contract tests against repositories created by newRepo.
func TestPersonaRepository(t *testing.T, newRepo PersonaRepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	const (
		aliceID domain.CollectorID = "aaaa"
		bobID   domain.CollectorID = "bbbb"
	)

	payslip := domain.ItemDefinition{
		ID: uuid.New(), Title: "Payslip", Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly,
		DocumentSpec: &domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
			{Name: "Gross", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR"}},
		}},
	}
	notice := domain.ItemDefinition{
		ID: uuid.New(), Title: "Tax notice", Kind: domain.KindDocument, Recurrence: domain.RecurrenceAnnually,
		DocumentSpec: &domain.DocumentSpec{},
	}
	days := domain.ItemDefinition{
		ID: uuid.New(), Title: "Commuting days", Kind: domain.KindData, Recurrence: domain.RecurrenceQuarterly,
		ValueSpec: &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true},
	}

	employed := personaWithItems(aliceID, "Employed", "Has a job", payslip, notice, days)
	employedRearranged := employed
	employedRearranged.Name = "Employed (renamed)"
	employedRearranged.Items = withPersonaID(employed.ID, days, payslip)
	employedReduced := employed
	employedReduced.Items = withPersonaID(employed.ID, notice)
	employedCleared := employed
	employedCleared.Items = nil
	hijack := employed
	hijack.CollectorID = bobID
	hijack.Name = "Hijacked"
	letting := personaWithItems(aliceID, "Letting", "")
	bobsDays := days
	bobsDays.ID = uuid.New() // item ids are unique across all personas
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	for _, tc := range []struct {
		name         string
		saves        []domain.Persona
		getCollector domain.CollectorID
		getID        uuid.UUID
		exp          *domain.Persona
		saveErr      error
		expErr       error
	}{
		{
			name:         "Get saved persona with its items in order",
			saves:        []domain.Persona{employed},
			getCollector: aliceID,
			getID:        employed.ID,
			exp:          &employed,
		},
		{
			name:         "Get persona with empty description and no items",
			saves:        []domain.Persona{letting},
			getCollector: aliceID,
			getID:        letting.ID,
			exp:          &letting,
		},
		{
			name:         "Saving again replaces the persona, reorders items and keeps their ids",
			saves:        []domain.Persona{employed, employedRearranged},
			getCollector: aliceID,
			getID:        employed.ID,
			exp:          &employedRearranged,
		},
		{
			name:         "Saving again deletes removed items",
			saves:        []domain.Persona{employed, employedReduced},
			getCollector: aliceID,
			getID:        employed.ID,
			exp:          &employedReduced,
		},
		{
			name:         "Saving again without items deletes all items",
			saves:        []domain.Persona{employed, employedCleared},
			getCollector: aliceID,
			getID:        employed.ID,
			exp:          &employedCleared,
		},
		{
			name:         "Saving a persona id of another collector is rejected and keeps the original",
			saves:        []domain.Persona{employed, hijack},
			saveErr:      domain.ErrPersonaNotFound,
			getCollector: aliceID,
			getID:        employed.ID,
			exp:          &employed,
		},
		{
			name:         "Get unknown persona",
			saves:        []domain.Persona{employed},
			getCollector: aliceID,
			getID:        uuid.New(),
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Get from empty repository",
			getCollector: aliceID,
			getID:        employed.ID,
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Persona of another collector is not found",
			saves:        []domain.Persona{employed, bobs},
			getCollector: bobID,
			getID:        employed.ID,
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Other personas are left untouched",
			saves:        []domain.Persona{employed, letting, bobs},
			getCollector: bobID,
			getID:        bobs.ID,
			exp:          &bobs,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t, aliceID, bobID)

			for _, p := range tc.saves {
				if err := repo.SavePersona(ctx, &p); err != nil && !errors.Is(err, tc.saveErr) {
					t.Fatalf("SavePersona() error = %v, want nil or %v", err, tc.saveErr)
				}
			}

			got, err := repo.GetPersona(ctx, tc.getCollector, tc.getID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("GetPersona() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got); diff != "" {
				t.Errorf("GetPersona() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// personaWithItems returns a persona whose items belong to it.
func personaWithItems(collectorID domain.CollectorID, name, description string, items ...domain.ItemDefinition) domain.Persona {
	p := domain.Persona{ID: uuid.New(), CollectorID: collectorID, Name: name, Description: description}
	p.Items = withPersonaID(p.ID, items...)

	return p
}

// withPersonaID returns copies of items that belong to the given persona, or nil for no items.
func withPersonaID(personaID uuid.UUID, items ...domain.ItemDefinition) []domain.ItemDefinition {
	if len(items) == 0 {
		return nil
	}

	out := slices.Clone(items)
	for i := range out {
		out[i].PersonaID = personaID
	}

	return out
}
