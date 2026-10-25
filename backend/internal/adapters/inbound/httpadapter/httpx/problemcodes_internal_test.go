package httpx

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// The table must list exactly the domain errors that exist: no sentinel without a code, no code for a sentinel
// that is gone.
func TestProblemCodes_MatchDomainErrors(t *testing.T) {
	registered := map[*domain.Error]bool{}

	for _, de := range domain.Errors() {
		registered[de] = true

		if _, ok := problemCodes[de]; !ok {
			t.Errorf("%v has no entry in problemCodes", de)
		}
	}

	for de, code := range problemCodes {
		if !registered[de] {
			t.Errorf("problemCodes has %q for %v, which is not a domain error", code, de)
		}
	}
}
