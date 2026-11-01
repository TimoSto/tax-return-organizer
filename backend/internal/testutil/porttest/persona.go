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

// PersonaStoreFactory returns an empty store for the given test whose only existing tax practices are the given ones.
// Repositories must report every other tax practice as domain.ErrTaxPracticeNotFound.
type PersonaStoreFactory func(t *testing.T, taxPractices ...domain.TaxPracticeID) outbound.PersonaStore

// TestPersonaRepository runs the contract tests against stores created by newStore. Each persona is saved through
// the repository of its own tax practice, as a use case would.
func TestPersonaRepository(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	var (
		aliceID = domaintest.TaxPracticeID()
		bobID   = domaintest.TaxPracticeID()
	)

	payslip := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
			{Name: "Gross", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "EUR"}},
		}})
	notice := domaintest.Item(domaintest.ItemID(), "Tax notice", domain.RecurrenceAnnually, domain.DocumentSpec{},
		domain.MetadataField{Name: "Tax office", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
		domain.MetadataField{Name: "Since", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeDate}})
	days := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly,
		domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true})

	employed := personaWithItems(aliceID, "Employed", "Has a job", payslip, notice, days)
	employedRearranged := variant(employed, "Employed (renamed)", aliceID, days, payslip)
	employedReduced := variant(employed, "Employed", aliceID, notice)
	employedCleared := variant(employed, "Employed", aliceID)
	singleNotice := domaintest.Item(notice.ID(), notice.Title(), notice.Recurrence(), notice.Spec())
	employedNoticeSingle := variant(employed, "Employed", aliceID, payslip, singleNotice, days)
	hijack := variant(employed, "Hijacked", bobID, payslip, notice, days)
	letting := personaWithItems(aliceID, "Letting", "")
	// explicit empty slice, which repositories may report as nil
	emptyFields := domaintest.Item(domaintest.ItemID(), "Tax notice", domain.RecurrenceAnnually,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{}})
	withEmptyFields := personaWithItems(aliceID, "Empty fields", "", emptyFields)
	// Item ids are unique per persona only: these reuse the id of employed's payslip.
	sharedItem := personaWithItems(aliceID, "Shared item", "", payslip)
	bobsPayslip := personaWithItems(bobID, "Bob's payslip", "", payslip)
	bobsDays := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly, days.Spec())
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	for _, tc := range []struct {
		name           string
		saves          []*domain.Persona
		getTaxPractice domain.TaxPracticeID
		getID          domain.PersonaID
		exp            *domain.Persona
		saveErr        error
		expErr         error
	}{
		{
			name:           "Get saved persona with its items in order",
			saves:          []*domain.Persona{employed},
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employed,
		},
		{
			name:           "A document spec with an empty list of metadata fields round-trips",
			saves:          []*domain.Persona{withEmptyFields},
			getTaxPractice: aliceID,
			getID:          withEmptyFields.ID(),
			exp:            withEmptyFields,
		},
		{
			name:           "Get persona with empty description and no items",
			saves:          []*domain.Persona{letting},
			getTaxPractice: aliceID,
			getID:          letting.ID(),
			exp:            letting,
		},
		{
			name:           "Saving again replaces the persona, reorders items and keeps their ids",
			saves:          []*domain.Persona{employed, employedRearranged},
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employedRearranged,
		},
		{
			name:           "Saving again updates the instance fields of an item",
			saves:          []*domain.Persona{employed, employedNoticeSingle},
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employedNoticeSingle,
		},
		{
			name:           "Saving again deletes removed items",
			saves:          []*domain.Persona{employed, employedReduced},
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employedReduced,
		},
		{
			name:           "Saving again without items deletes all items",
			saves:          []*domain.Persona{employed, employedCleared},
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employedCleared,
		},
		{
			name:           "Saving a persona id of another tax practice is rejected and keeps the original",
			saves:          []*domain.Persona{employed, hijack},
			saveErr:        domain.ErrPersonaNotFound,
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			exp:            employed,
		},
		{
			name:           "Another persona may use the same item id",
			saves:          []*domain.Persona{employed, sharedItem},
			getTaxPractice: aliceID,
			getID:          sharedItem.ID(),
			exp:            sharedItem,
		},
		{
			name:           "Removing an item from one persona keeps it in another persona with the same item id",
			saves:          []*domain.Persona{employed, sharedItem, employedReduced},
			getTaxPractice: aliceID,
			getID:          sharedItem.ID(),
			exp:            sharedItem,
		},
		{
			name:           "A persona of another tax practice may use the same item id",
			saves:          []*domain.Persona{employed, bobsPayslip},
			getTaxPractice: bobID,
			getID:          bobsPayslip.ID(),
			exp:            bobsPayslip,
		},
		{
			name:           "Get unknown persona",
			saves:          []*domain.Persona{employed},
			getTaxPractice: aliceID,
			getID:          newPersonaID(),
			expErr:         domain.ErrPersonaNotFound,
		},
		{
			name:           "Get from empty repository",
			getTaxPractice: aliceID,
			getID:          employed.ID(),
			expErr:         domain.ErrPersonaNotFound,
		},
		{
			name:           "Persona of another tax practice is not found",
			saves:          []*domain.Persona{employed, bobs},
			getTaxPractice: bobID,
			getID:          employed.ID(),
			expErr:         domain.ErrPersonaNotFound,
		},
		{
			name:           "Other personas are left untouched",
			saves:          []*domain.Persona{employed, letting, bobs},
			getTaxPractice: bobID,
			getID:          bobs.ID(),
			exp:            bobs,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore(t, aliceID, bobID)

			for _, p := range tc.saves {
				if err := store.For(p.TaxPracticeID()).SavePersona(ctx, p); err != nil && !errors.Is(err, tc.saveErr) {
					t.Fatalf("SavePersona() error = %v, want nil or %v", err, tc.saveErr)
				}
			}

			got, err := store.For(tc.getTaxPractice).GetPersona(ctx, tc.getID)

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
	taxPracticeID domain.TaxPracticeID, name, description string, items ...domain.ItemDefinition,
) *domain.Persona {
	return domaintest.Persona(domaintest.PersonaID(), taxPracticeID, name, description, items...)
}

// variant returns a persona with the id of p but another name, tax practice and items.
func variant(p *domain.Persona, name string, taxPracticeID domain.TaxPracticeID, items ...domain.ItemDefinition) *domain.Persona {
	return domaintest.Persona(p.ID(), taxPracticeID, name, p.Description(), items...)
}

// TestPersonaRepositoryList runs the contract tests of ListPersonas against stores created by newStore.
func TestPersonaRepositoryList(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	var (
		aliceID = domaintest.TaxPracticeID()
		bobID   = domaintest.TaxPracticeID()
		noneID  = domaintest.TaxPracticeID()
	)

	payslip := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly,
		domain.DocumentSpec{MetadataFields: []domain.MetadataField{
			{Name: "Employer", ValueSpec: domain.ValueSpec{Type: domain.ValueTypeText, Required: true}},
		}})
	days := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly,
		domain.ValueSpec{Type: domain.ValueTypeNumber, Unit: "days", Required: true})
	bobsDays := domaintest.Item(domaintest.ItemID(), "Commuting days", domain.RecurrenceQuarterly, days.Spec())

	zeta := personaWithItems(aliceID, "Zeta", "Last", payslip, days)
	alpha1 := personaWithItems(aliceID, "alpha", "")
	alpha2 := personaWithItems(aliceID, "alpha", "")
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer", bobsDays)

	for _, tc := range []struct {
		name          string
		saves         []*domain.Persona
		taxPracticeID domain.TaxPracticeID
		exp           []*domain.Persona
		expErr        error
	}{
		{
			name:          "Lists all personas with their items in order",
			saves:         []*domain.Persona{alpha1, zeta, alpha2},
			taxPracticeID: aliceID,
			exp:           []*domain.Persona{alpha1, zeta, alpha2},
		},
		{
			name:          "Personas of other tax practices are not listed",
			saves:         []*domain.Persona{zeta, bobs},
			taxPracticeID: bobID,
			exp:           []*domain.Persona{bobs},
		},
		{
			name:          "A tax practice without personas yields an empty list",
			saves:         []*domain.Persona{zeta},
			taxPracticeID: noneID,
			exp:           []*domain.Persona{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore(t, aliceID, bobID, noneID)

			for _, p := range tc.saves {
				if err := store.For(p.TaxPracticeID()).SavePersona(ctx, p); err != nil {
					t.Fatalf("SavePersona() error = %v", err)
				}
			}

			got, err := store.For(tc.taxPracticeID).ListPersonas(ctx)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("ListPersonas() error = %v, wantErr %v", err, tc.expErr)
			}

			// The order of the personas is not part of the contract, only that of their items.
			byID := func(a, b *domain.Persona) int { return compareIDs(a.ID(), b.ID()) }
			slices.SortFunc(tc.exp, byID)
			slices.SortFunc(got, byID)

			if diff := cmp.Diff(tc.exp, got, domaintest.CmpOptions()); diff != "" {
				t.Errorf("ListPersonas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPersonaRepositoryIsolation runs the contract tests that a repository bound to one tax practice cannot reach the
// personas of another one, whatever the caller passes.
func TestPersonaRepositoryIsolation(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	var (
		aliceID = domaintest.TaxPracticeID()
		bobID   = domaintest.TaxPracticeID()
	)

	alices := personaWithItems(aliceID, "Employed", "Has a job")
	bobs := personaWithItems(bobID, "Self-employed", "Freelancer")

	t.Run("A repository bound to another tax practice cannot get a persona by its id", func(t *testing.T) {
		store := newStore(t, aliceID, bobID)
		mustSave(ctx, t, store.For(aliceID), alices)

		if _, err := store.For(bobID).GetPersona(ctx, alices.ID()); !errors.Is(err, domain.ErrPersonaNotFound) {
			t.Errorf("GetPersona() error = %v, want %v", err, domain.ErrPersonaNotFound)
		}
	})

	t.Run("A repository bound to another tax practice does not list the persona", func(t *testing.T) {
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

	t.Run("Saving a persona of another tax practice is rejected and changes nothing", func(t *testing.T) {
		store := newStore(t, aliceID, bobID)

		if err := store.For(aliceID).SavePersona(ctx, bobs); !errors.Is(err, outbound.ErrTenantMismatch) {
			t.Fatalf("SavePersona() error = %v, want %v", err, outbound.ErrTenantMismatch)
		}

		for _, id := range []domain.TaxPracticeID{aliceID, bobID} {
			got, err := store.For(id).ListPersonas(ctx)
			if err != nil || len(got) != 0 {
				t.Errorf("ListPersonas(%s) = %v, %v, want an empty list", id, got, err)
			}
		}
	})
}

// TestPersonaRepositoryUnknownTaxPractice runs the contract tests that a repository bound to a tax practice that does
// not exist reports domain.ErrTaxPracticeNotFound on every method and saves nothing.
func TestPersonaRepositoryUnknownTaxPractice(t *testing.T, newStore PersonaStoreFactory) {
	t.Helper()

	ctx := context.Background()

	var (
		aliceID   = domaintest.TaxPracticeID()
		unknownID = domaintest.TaxPracticeID()
	)

	t.Run("Get reports the unknown tax practice, not the missing persona", func(t *testing.T) {
		store := newStore(t, aliceID)

		if _, err := store.For(unknownID).GetPersona(ctx, newPersonaID()); !errors.Is(err, domain.ErrTaxPracticeNotFound) {
			t.Errorf("GetPersona() error = %v, want %v", err, domain.ErrTaxPracticeNotFound)
		}
	})

	t.Run("List reports the unknown tax practice instead of an empty list", func(t *testing.T) {
		store := newStore(t, aliceID)

		got, err := store.For(unknownID).ListPersonas(ctx)
		if !errors.Is(err, domain.ErrTaxPracticeNotFound) {
			t.Errorf("ListPersonas() = %v, %v, want error %v", got, err, domain.ErrTaxPracticeNotFound)
		}
	})

	t.Run("Saving for an unknown tax practice is rejected and changes nothing", func(t *testing.T) {
		store := newStore(t, aliceID)
		item := domaintest.Item(domaintest.ItemID(), "Payslip", domain.RecurrenceMonthly, domain.DocumentSpec{})
		orphan := personaWithItems(unknownID, "Orphan", "", item)

		if err := store.For(unknownID).SavePersona(ctx, orphan); !errors.Is(err, domain.ErrTaxPracticeNotFound) {
			t.Fatalf("SavePersona() error = %v, want %v", err, domain.ErrTaxPracticeNotFound)
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

	t.Run("Saving a persona id of another tax practice for an unknown tax practice reports the tax practice", func(t *testing.T) {
		store := newStore(t, aliceID)
		alices := personaWithItems(aliceID, "Employed", "Has a job")
		mustSave(ctx, t, store.For(aliceID), alices)

		// Both causes apply. The tax practice takes precedence, like in GetPersona.
		hijack := variant(alices, "Hijacked", unknownID)
		if err := store.For(unknownID).SavePersona(ctx, hijack); !errors.Is(err, domain.ErrTaxPracticeNotFound) {
			t.Fatalf("SavePersona() error = %v, want %v", err, domain.ErrTaxPracticeNotFound)
		}

		got, err := store.For(aliceID).GetPersona(ctx, alices.ID())
		if err != nil {
			t.Fatalf("GetPersona() error = %v", err)
		}

		if diff := cmp.Diff(alices, got, domaintest.CmpOptions()); diff != "" {
			t.Errorf("GetPersona() mismatch (-want +got):\n%s", diff)
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
