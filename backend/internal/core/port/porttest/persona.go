package porttest

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// PersonaRepositoryFactory returns an empty repository for the given test.
// Implementations that enforce referential integrity must make sure the given collectors exist.
type PersonaRepositoryFactory func(t *testing.T, collectors ...domain.CollectorID) port.PersonaRepository

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
	emptyFields := notice // explicit empty slice, which every repository reports as nil
	emptyFields.ID = uuid.New()
	emptyFields.DocumentSpec = &domain.DocumentSpec{MetadataFields: []domain.MetadataField{}}
	withEmptyFields := personaWithItems(aliceID, "Empty fields", "", emptyFields)
	nilFields := emptyFields
	nilFields.DocumentSpec = &domain.DocumentSpec{}
	withNilFields := withEmptyFields
	withNilFields.Items = withPersonaID(withEmptyFields.ID, nilFields)
	orphan := personaWithItems("zzzz", "Orphan", "")                    // no such collector
	stolenItem := personaWithItems(aliceID, "Stolen item", "", payslip) // payslip belongs to employed
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
			name:         "A document spec without metadata fields is returned with nil fields",
			saves:        []domain.Persona{withEmptyFields},
			getCollector: aliceID,
			getID:        withEmptyFields.ID,
			exp:          &withNilFields,
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
			name:         "Saving an item id used by another persona is rejected",
			saves:        []domain.Persona{employed, stolenItem},
			saveErr:      domain.ErrItemIDTaken,
			getCollector: aliceID,
			getID:        stolenItem.ID,
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Saving a persona for an unknown collector is rejected",
			saves:        []domain.Persona{orphan},
			saveErr:      domain.ErrUnknownCollector,
			getCollector: "zzzz",
			getID:        orphan.ID,
			expErr:       domain.ErrPersonaNotFound,
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

// TestPersonaRepositoryList runs the contract tests of ListPersonas against repositories created by newRepo.
func TestPersonaRepositoryList(t *testing.T, newRepo PersonaRepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	const (
		aliceID domain.CollectorID = "aaaa"
		bobID   domain.CollectorID = "bbbb"
		noneID  domain.CollectorID = "cccc"
	)

	payslip := domain.ItemDefinition{
		ID: uuid.New(), Title: "Payslip", Kind: domain.KindDocument, Recurrence: domain.RecurrenceMonthly,
		DocumentSpec: &domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
		}},
	}
	days := domain.ItemDefinition{
		ID: uuid.New(), Title: "Commuting days", Kind: domain.KindData, Recurrence: domain.RecurrenceQuarterly,
		ValueSpec: &domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true},
	}
	bobsDays := days
	bobsDays.ID = uuid.New() // item ids are unique across all personas

	// Names are compared bytewise, so "Zeta" sorts before "alpha". Equal names fall back to the id.
	zeta := personaWithItems(aliceID, "Zeta", "Last", payslip, days)
	alpha1 := personaWithItems(aliceID, "alpha", "")
	alpha2 := personaWithItems(aliceID, "alpha", "")
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	alphas := []domain.Persona{alpha1, alpha2}
	slices.SortFunc(alphas, func(a, b domain.Persona) int { return bytes.Compare(a.ID[:], b.ID[:]) })

	for _, tc := range []struct {
		name        string
		saves       []domain.Persona
		collectorID domain.CollectorID
		exp         []*domain.Persona
		expErr      error
	}{
		{
			name:        "Lists personas with their items ordered by name and id",
			saves:       []domain.Persona{alpha1, zeta, alpha2},
			collectorID: aliceID,
			exp:         []*domain.Persona{&zeta, &alphas[0], &alphas[1]},
		},
		{
			name:        "Personas of other collectors are not listed",
			saves:       []domain.Persona{zeta, bobs},
			collectorID: bobID,
			exp:         []*domain.Persona{&bobs},
		},
		{
			name:        "A collector without personas yields an empty list",
			saves:       []domain.Persona{zeta},
			collectorID: noneID,
			exp:         []*domain.Persona{},
		},
		{
			name:        "An unknown collector is not found",
			saves:       []domain.Persona{zeta},
			collectorID: "zzzz",
			expErr:      domain.ErrCollectorNotFound,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t, aliceID, bobID, noneID)

			for _, p := range tc.saves {
				if err := repo.SavePersona(ctx, &p); err != nil {
					t.Fatalf("SavePersona() error = %v", err)
				}
			}

			got, err := repo.ListPersonas(ctx, tc.collectorID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("ListPersonas() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got); diff != "" {
				t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
