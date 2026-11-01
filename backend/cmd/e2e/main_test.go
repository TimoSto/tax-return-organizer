//go:build e2e

package main

import (
	"context"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// TODO(auth): delete with staticSessions.
func TestStaticSessions(t *testing.T) {
	actor, ok, err := staticSessions{}.ResolveSession(context.Background(), sessionToken)
	if err != nil || !ok {
		t.Fatalf("sessionToken: ok = %v, err = %v, want resolved", ok, err)
	}

	if want := domain.NewStaffActor(defaultTaxPracticeID, employeeAID); actor != want {
		t.Errorf("actor = %+v, want employeeA %+v", actor, want)
	}

	for _, token := range []string{"", "e2e-employee-b", sessionToken + "x"} {
		if _, ok, err := (staticSessions{}).ResolveSession(context.Background(), token); ok || err != nil {
			t.Errorf("%q: ok = %v, err = %v, want unknown", token, ok, err)
		}
	}
}
