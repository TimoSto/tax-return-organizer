package domain_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// Every sentinel must be a classified *domain.Error with a message, otherwise adapters would answer it with a 500.
func TestSentinelsAreClassified(t *testing.T) {
	errs := domain.Errors()
	if len(errs) == 0 {
		t.Fatal("no domain errors registered")
	}

	for _, de := range errs {
		if de.Kind == 0 || de.Error() == "" {
			t.Errorf("%v is not classified or has no message", de)
		}
	}
}
