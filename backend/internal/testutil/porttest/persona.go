package porttest

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// PersonaStoreFactory returns an empty store for the given test whose only existing collectors are the given ones.
// Repositories must report every other collector as domain.ErrCollectorNotFound.
type PersonaStoreFactory func(t *testing.T, collectors ...domain.CollectorID) outbound.PersonaStore

// TestPersonaRepository runs the contract tests against stores created by newStore. Each persona is saved through
// the repository of its own collector, as a use case would.
func TestPersonaRepository(t *testing.T, newStore PersonaStoreFactory) {
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
			store := newStore(t, aliceID, bobID)

			for _, p := range tc.saves {
				if err := store.For(p.CollectorID()).SavePersona(ctx, p); err != nil && !errors.Is(err, tc.saveErr) {
					t.Fatalf("SavePersona() error = %v, want nil or %v", err, tc.saveErr)
				}
			}

			got, err := store.For(tc.getCollector).GetPersona(ctx, tc.getID)

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

// TestPersonaRepositoryList runs the contract tests of ListPersonas against stores created by newStore.
func TestPersonaRepositoryList(t *testing.T, newStore PersonaStoreFactory) {
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
			store := newStore(t, aliceID, bobID, noneID)

			for _, p := range tc.saves {
				if err := store.For(p.CollectorID()).SavePersona(ctx, p); err != nil {
					t.Fatalf("SavePersona() error = %v", err)
				}
			}

			got, err := store.For(tc.collectorID).ListPersonas(ctx)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("ListPersonas() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got, domaintest.CmpOptions()); diff != "" {
				t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPersonaRepositoryIsolation runs the contract tests that a repository bound to one collector cannot reach the
// personas of another one, whatever the caller passes.
func TestPersonaRepositoryIsolation(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	const (
		aliceID domain.CollectorID = "aaaa"
		bobID   domain.CollectorID = "bbbb"
	)

	alices := personaWithItems(aliceID, "Employed", "Has a job")
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer")

	t.Run("A repository bound to another collector cannot get a persona by its id", func(t *testing.T) {
		store := newStore(t, aliceID, bobID)
		mustSave(ctx, t, store.For(aliceID), alices)

		if _, err := store.For(bobID).GetPersona(ctx, alices.ID()); !errors.Is(err, domain.ErrPersonaNotFound) {
			t.Errorf("GetPersona() error = %v, want %v", err, domain.ErrPersonaNotFound)
		}
	})

	t.Run("A repository bound to another collector does not list the persona", func(t *testing.T) {
		store := newStore(t, aliceID, bobID)
		mustSave(ctx, t, store.For(aliceID), alices)
		mustSave(ctx, t, store.For(bobID), bobs)

		got, err := store.For(bobID).ListPersonas(ctx)
		if err != nil {
			t.Fatalf("ListPersonas() error = %v", err)
		}

		if diff := cmp.Diff([]*domain.Persona{bobs}, got, domaintest.CmpOptions()); diff != "" {
			t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Saving a persona of another collector is rejected and changes nothing", func(t *testing.T) {
		store := newStore(t, aliceID, bobID)

		if err := store.For(aliceID).SavePersona(ctx, bobs); !errors.Is(err, outbound.ErrTenantMismatch) {
			t.Fatalf("SavePersona() error = %v, want %v", err, outbound.ErrTenantMismatch)
		}

		for _, id := range []domain.CollectorID{aliceID, bobID} {
			got, err := store.For(id).ListPersonas(ctx)
			if err != nil || len(got) != 0 {
				t.Errorf("ListPersonas(%s) = %v, %v, want an empty list", id, got, err)
			}
		}
	})
}

// TestPersonaRepositoryUnknownCollector runs the contract tests that a repository bound to a collector that does
// not exist reports domain.ErrCollectorNotFound on every method and saves nothing.
func TestPersonaRepositoryUnknownCollector(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	const (
		aliceID   domain.CollectorID = "aaaa"
		unknownID domain.CollectorID = "zzzz"
	)

	t.Run("Get reports the unknown collector, not the missing persona", func(t *testing.T) {
		store := newStore(t, aliceID)

		if _, err := store.For(unknownID).GetPersona(ctx, newPersonaID()); !errors.Is(err, domain.ErrCollectorNotFound) {
			t.Errorf("GetPersona() error = %v, want %v", err, domain.ErrCollectorNotFound)
		}
	})

	t.Run("List reports the unknown collector instead of an empty list", func(t *testing.T) {
		store := newStore(t, aliceID)

		got, err := store.For(unknownID).ListPersonas(ctx)
		if !errors.Is(err, domain.ErrCollectorNotFound) {
			t.Errorf("ListPersonas() = %v, %v, want error %v", got, err, domain.ErrCollectorNotFound)
		}
	})

	t.Run("Saving for an unknown collector is rejected and changes nothing", func(t *testing.T) {
		store := newStore(t, aliceID)
		item := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{})
		orphan := personaWithItems(unknownID, "Orphan", "", item)

		if err := store.For(unknownID).SavePersona(ctx, orphan); !errors.Is(err, domain.ErrCollectorNotFound) {
			t.Fatalf("SavePersona() error = %v, want %v", err, domain.ErrCollectorNotFound)
		}

		// The item id was not taken by the rejected save.
		reuse := personaWithItems(aliceID, "Employed", "", item)
		mustSave(ctx, t, store.For(aliceID), reuse)

		got, err := store.For(aliceID).ListPersonas(ctx)
		if err != nil {
			t.Fatalf("ListPersonas() error = %v", err)
		}

		if diff := cmp.Diff([]*domain.Persona{reuse}, got, domaintest.CmpOptions()); diff != "" {
			t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
		}
	})
}

func mustSave(ctx context.Context, t *testing.T, repo outbound.PersonaRepository, p *domain.Persona) {
	t.Helper()

	if err := repo.SavePersona(ctx, p); err != nil {
		t.Fatalf("SavePersona() error = %v", err)
	}
}

func newPersonaID() domain.PersonaID { return domaintest.PersonaID() }

func compareIDs(a, b domain.PersonaID) int { return bytes.Compare(a[:], b[:]) }
