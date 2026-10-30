package httpx

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// The table must not keep a code for a sentinel that is gone. The other direction, no sentinel without a code, is
// TestEveryDomainErrorHasAProblemCode in internal/backend: only there are all sentinels registered. This direction
// works here, since this package imports every package the table refers to.
func TestProblemCodes_OnlyDomainErrors(t *testing.T) {
	registered := map[*domain.Error]bool{}

	for _, de := range domain.Errors() {
		registered[de] = true
	}

	for de, code := range problemCodes {
		if !registered[de] {
			t.Errorf("problemCodes has %q for %v, which is not a domain error", code, de)
		}
	}
}
