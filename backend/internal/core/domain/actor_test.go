package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestActor(t *testing.T) {
	user := domain.UserID(uuid.New())

	for _, tc := range []struct {
		name        string
		actor       domain.Actor
		role        domain.Role
		taxPractice domain.TaxPracticeID
		user        domain.UserID
		staffErr    error
	}{
		{"staff", domain.NewStaffActor(testTaxPracticeID, user), domain.RoleStaff, testTaxPracticeID, user, nil},
		{"client", domain.NewClientActor(testTaxPracticeID, user), domain.RoleClient, testTaxPracticeID, user, domain.ErrForbidden},
		{"zero", domain.Actor{}, 0, domain.TaxPracticeID{}, domain.UserID{}, domain.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.actor
			if a.Role() != tc.role || a.TaxPracticeID() != tc.taxPractice || a.UserID() != tc.user {
				t.Errorf("actor = (%v, %v, %v), want (%v, %v, %v)",
					a.Role(), a.TaxPracticeID(), a.UserID(), tc.role, tc.taxPractice, tc.user)
			}

			if err := a.RequireStaff(); !errors.Is(err, tc.staffErr) {
				t.Errorf("RequireStaff() = %v, want %v", err, tc.staffErr)
			}
		})
	}
}
