package porttest

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// PersonaRepositoryFactory returns an empty repository for the given test.
// Implementations that enforce referential integrity (the foreign key to collectors) must make sure the given
// collectors exist. The repository itself does not check collectors, callers do.
type PersonaRepositoryFactory func(t *testing.T, collectors ...domain.CollectorID) port.PersonaRepository

// TestPersonaRepository runs the contract tests against repositories created by newRepo.
func TestPersonaRepository(t *testing.T, newRepo PersonaRepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	const (
		aliceID domain.CollectorID = "aaaa"
		bobID   domain.CollectorID = "bbbb"
	)

	payslip := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
			{Name: "Gross", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR"}},
		}})
	notice := domaintest.Item(domaintest.ItemID(), "Tax notice", domain.RecurrenceAnnually, domain.DocumentSpec{})
	days := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly,
		domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true})

	employed := personaWithItems(aliceID, "Employed", "Has a job", payslip, notice, days)
	employedRearranged := variant(employed, "Employed (renamed)", aliceID, days, payslip)
	employedReduced := variant(employed, "Employed", aliceID, notice)
	employedCleared := variant(employed, "Employed", aliceID)
	hijack := variant(employed, "Hijacked", bobID, payslip, notice, days)
	letting := personaWithItems(aliceID, "Letting", "")
	// explicit empty slice, which repositories may report as nil
	emptyFields := domaintest.Item(domaintest.ItemID(), "Tax notice", domain.RecurrenceAnnually,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{}})
	withEmptyFields := personaWithItems(aliceID, "Empty fields", "", emptyFields)
	stolenItem := personaWithItems(aliceID, "Stolen item", "", payslip) // payslip belongs to employed
	// item ids are unique across all personas
	bobsDays := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly, days.Spec())
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	for _, tc := range []struct {
		name         string
		saves        []*domain.Persona
		getCollector domain.CollectorID
		getID        domain.PersonaID
		exp          *domain.Persona
		saveErr      error
		expErr       error
	}{
		{
			name:         "Get saved persona with its items in order",
			saves:        []*domain.Persona{employed},
			getCollector: aliceID,
			getID:        employed.ID(),
			exp:          employed,
		},
		{
			name:         "A document spec with an empty list of metadata fields round-trips",
			saves:        []*domain.Persona{withEmptyFields},
			getCollector: aliceID,
			getID:        withEmptyFields.ID(),
			exp:          withEmptyFields,
		},
		{
			name:         "Get persona with empty description and no items",
			saves:        []*domain.Persona{letting},
			getCollector: aliceID,
			getID:        letting.ID(),
			exp:          letting,
		},
		{
			name:         "Saving again replaces the persona, reorders items and keeps their ids",
			saves:        []*domain.Persona{employed, employedRearranged},
			getCollector: aliceID,
			getID:        employed.ID(),
			exp:          employedRearranged,
		},
		{
			name:         "Saving again deletes removed items",
			saves:        []*domain.Persona{employed, employedReduced},
			getCollector: aliceID,
			getID:        employed.ID(),
			exp:          employedReduced,
		},
		{
			name:         "Saving again without items deletes all items",
			saves:        []*domain.Persona{employed, employedCleared},
			getCollector: aliceID,
			getID:        employed.ID(),
			exp:          employedCleared,
		},
		{
			name:         "Saving a persona id of another collector is rejected and keeps the original",
			saves:        []*domain.Persona{employed, hijack},
			saveErr:      domain.ErrPersonaNotFound,
			getCollector: aliceID,
			getID:        employed.ID(),
			exp:          employed,
		},
		{
			name:         "Saving an item id used by another persona is rejected",
			saves:        []*domain.Persona{employed, stolenItem},
			saveErr:      domain.ErrItemIDTaken,
			getCollector: aliceID,
			getID:        stolenItem.ID(),
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Get unknown persona",
			saves:        []*domain.Persona{employed},
			getCollector: aliceID,
			getID:        newPersonaID(),
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Get from empty repository",
			getCollector: aliceID,
			getID:        employed.ID(),
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Persona of another collector is not found",
			saves:        []*domain.Persona{employed, bobs},
			getCollector: bobID,
			getID:        employed.ID(),
			expErr:       domain.ErrPersonaNotFound,
		},
		{
			name:         "Other personas are left untouched",
			saves:        []*domain.Persona{employed, letting, bobs},
			getCollector: bobID,
			getID:        bobs.ID(),
			exp:          bobs,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t, aliceID, bobID)

			for _, p := range tc.saves {
				if err := repo.SavePersona(ctx, p); err != nil && !errors.Is(err, tc.saveErr) {
					t.Fatalf("SavePersona() error = %v, want nil or %v", err, tc.saveErr)
				}
			}

			got, err := repo.GetPersona(ctx, tc.getCollector, tc.getID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("GetPersona() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got, domaintest.CmpOptions()); diff != "" {
				t.Errorf("GetPersona() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// personaWithItems returns a persona with the given items, or none.
func personaWithItems(
	collectorID domain.CollectorID, name, description string, items ...domain.ItemDefinition,
) *domain.Persona {
	return domaintest.Persona(domaintest.PersonaID(), collectorID, name, description, items...)
}

// variant returns a persona with the id of p but another name, collector and items.
func variant(p *domain.Persona, name string, collectorID domain.CollectorID, items ...domain.ItemDefinition) *domain.Persona {
	return domaintest.Persona(p.ID(), collectorID, name, p.Description(), items...)
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

	payslip := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
		}})
	days := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly,
		domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true})
	// item ids are unique across all personas
	bobsDays := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly, days.Spec())

	// Names are compared bytewise, so "Zeta" sorts before "alpha". Equal names fall back to the id.
	zeta := personaWithItems(aliceID, "Zeta", "Last", payslip, days)
	alpha1 := personaWithItems(aliceID, "alpha", "")
	alpha2 := personaWithItems(aliceID, "alpha", "")
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	alphas := []*domain.Persona{alpha1, alpha2}
	slices.SortFunc(alphas, func(a, b *domain.Persona) int { return compareIDs(a.ID(), b.ID()) })

	for _, tc := range []struct {
		name        string
		saves       []*domain.Persona
		collectorID domain.CollectorID
		exp         []*domain.Persona
		expErr      error
	}{
		{
			name:        "Lists personas with their items ordered by name and id",
			saves:       []*domain.Persona{alpha1, zeta, alpha2},
			collectorID: aliceID,
			exp:         []*domain.Persona{zeta, alphas[0], alphas[1]},
		},
		{
			name:        "Personas of other collectors are not listed",
			saves:       []*domain.Persona{zeta, bobs},
			collectorID: bobID,
			exp:         []*domain.Persona{bobs},
		},
		{
			name:        "A collector without personas yields an empty list",
			saves:       []*domain.Persona{zeta},
			collectorID: noneID,
			exp:         []*domain.Persona{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t, aliceID, bobID, noneID)

			for _, p := range tc.saves {
				if err := repo.SavePersona(ctx, p); err != nil {
					t.Fatalf("SavePersona() error = %v", err)
				}
			}

			got, err := repo.ListPersonas(ctx, tc.collectorID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("ListPersonas() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got, domaintest.CmpOptions()); diff != "" {
				t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func newPersonaID() domain.PersonaID { return domaintest.PersonaID() }

func compareIDs(a, b domain.PersonaID) int { return bytes.Compare(a[:], b[:]) }
