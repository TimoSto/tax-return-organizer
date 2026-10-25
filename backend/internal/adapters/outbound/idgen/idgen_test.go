package idgen_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/idgen"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestRandom_NewCollectorID(t *testing.T) {
	seen := map[domain.CollectorID]bool{}

	for range 200 {
		id, err := idgen.Random{}.NewCollectorID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := domain.ParseCollectorID(id.String()); err != nil {
			t.Fatalf("generated id %q is not parseable: %v", id, err)
		}

		seen[id] = true
	}

	if len(seen) < 150 {
		t.Errorf("only %d distinct ids in 200 draws, generator looks broken", len(seen))
	}
}

func TestRandom_AreUnique(t *testing.T) {
	ids := idgen.Random{}

	p1, _ := ids.NewPersonaID()
	p2, _ := ids.NewPersonaID()
	i1, _ := ids.NewItemID()
	i2, _ := ids.NewItemID()

	if p1 == p2 || i1 == i2 || p1 == (domain.PersonaID{}) || i1 == (domain.ItemID{}) {
		t.Errorf("ids are not unique and non-zero: %v %v %v %v", p1, p2, i1, i2)
	}
}
