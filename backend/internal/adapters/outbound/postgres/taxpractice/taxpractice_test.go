package taxpractice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/persona"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/pgdb"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/taxpractice"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/porttest"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/postgrestest"
)

var (
	// testTaxPracticeID is the tax practice the tests act for.
	testTaxPracticeID = domaintest.TaxPracticeID()
	// otherTaxPracticeID is another tax practice, which owns data the tests must not see.
	otherTaxPracticeID = domaintest.TaxPracticeID()
)

func TestTaxPracticeRepository(t *testing.T) {
	pool := postgrestest.RequirePool(t)

	porttest.TestTaxPracticeRepository(t, func(t *testing.T) outbound.TaxPracticeRepository {
		if _, err := pool.Exec(context.Background(), `TRUNCATE tax_practices CASCADE`); err != nil {
			t.Fatalf("truncate tax practices: %v", err)
		}

		return taxpractice.NewRepository(pool)
	})
}

// If saving a persona fails inside the transaction, the tax practice saved before it must not stay behind.
// The failure is a real one: the persona id already belongs to another tax practice, so Postgres updates no row.
func TestSaveTaxPractice_RolledBackWhenPersonaSaveFails(t *testing.T) {
	ctx := context.Background()
	pool := postgrestest.RequirePool(t)

	if _, err := pool.Exec(ctx, `TRUNCATE personas, tax_practices CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// otherTaxPracticeID owns a persona, whose id testTaxPracticeID's persona then tries to reuse.
	if err := taxpractice.NewRepository(pool).SaveTaxPractice(ctx, domaintest.TaxPractice(otherTaxPracticeID, "B", "b@example.com")); err != nil {
		t.Fatalf("seed tax practice: %v", err)
	}

	existing := domaintest.Persona(domaintest.PersonaID(), otherTaxPracticeID, "Employed", "")

	if err := persona.NewStore(pool).For(otherTaxPracticeID).SavePersona(ctx, existing); err != nil {
		t.Fatalf("seed persona: %v", err)
	}

	clashing := domaintest.Persona(existing.ID(), testTaxPracticeID, "Employed", "")

	tx := pgdb.NewTransactor(pool, func(db pgdb.DBTX) pgdb.DBTX { return db })

	err := tx.InTx(ctx, func(ctx context.Context, db pgdb.DBTX) error {
		if err := taxpractice.NewRepository(db).SaveTaxPractice(ctx, domaintest.TaxPractice(testTaxPracticeID, "A", "a@example.com")); err != nil {
			return err
		}

		return persona.NewStore(db).For(testTaxPracticeID).SavePersona(ctx, clashing)
	})
	if !errors.Is(err, domain.ErrPersonaNotFound) {
		t.Fatalf("err = %v, want %v", err, domain.ErrPersonaNotFound)
	}

	if _, err := taxpractice.NewRepository(pool).GetTaxPractice(ctx, testTaxPracticeID); !errors.Is(err, domain.ErrTaxPracticeNotFound) {
		t.Errorf("GetTaxPractice(testTaxPracticeID) err = %v, want %v: the tax practice was not rolled back", err, domain.ErrTaxPracticeNotFound)
	}

	// With the tax practice rolled back, the store reports it missing before looking for the persona.
	if _, err := persona.NewStore(pool).For(testTaxPracticeID).GetPersona(ctx, clashing.ID()); !errors.Is(err, domain.ErrTaxPracticeNotFound) {
		t.Errorf("GetPersona err = %v, want %v", err, domain.ErrTaxPracticeNotFound)
	}

	// The seeded data of the other tax practice is untouched.
	got, err := persona.NewStore(pool).For(otherTaxPracticeID).GetPersona(ctx, existing.ID())
	if err != nil {
		t.Fatalf("seeded persona lost: %v", err)
	}

	if got.TaxPracticeID() != otherTaxPracticeID || got.Name() != existing.Name() {
		t.Errorf("seeded persona = %q of %v, want %q of %v", got.Name(), got.TaxPracticeID(), existing.Name(), otherTaxPracticeID)
	}
}
