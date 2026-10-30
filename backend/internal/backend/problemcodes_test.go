package backend_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"

	// Links the whole production graph, so that domain.Errors() holds every sentinel the server can return.
	_ "github.com/TimoSto/tax-return-organizer/backend/internal/backend"
)

// Every domain error must reach the client with its own, unique snake_case code and a 4xx status. The test lives
// here and not in httpx, because domain.Errors() only knows the sentinels of the packages linked into the test
// binary: httpx does not import every core package, the composition root does.
func TestEveryDomainErrorHasAProblemCode(t *testing.T) {
	snake := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	seen := map[string]*domain.Error{}

	for _, de := range domain.Errors() {
		rec := httptest.NewRecorder()

		httpx.WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), de)

		var p httpx.Problem
		if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}

		if rec.Code >= 500 || p.Code == httpx.CodeInternal {
			t.Errorf("%v: status %d, code %q: no problem code in httpx/problemcodes.go?", de, rec.Code, p.Code)

			continue
		}

		if !snake.MatchString(p.Code) {
			t.Errorf("%v: code %q is not snake_case", de, p.Code)
		}

		if prev, dup := seen[p.Code]; dup {
			t.Errorf("code %q used by %v and %v", p.Code, prev, de)
		}

		seen[p.Code] = de
	}
}
