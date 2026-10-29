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
		name      string
		actor     domain.Actor
		role      domain.Role
		collector domain.CollectorID
		user      domain.UserID
		staffErr  error
	}{
		{"staff", domain.NewStaffActor(testCollectorID, user), domain.RoleStaff, testCollectorID, user, nil},
		{"client", domain.NewClientActor(testCollectorID, user), domain.RoleClient, testCollectorID, user, domain.ErrForbidden},
		{"zero", domain.Actor{}, 0, domain.CollectorID{}, domain.UserID{}, domain.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.actor
			if a.Role() != tc.role || a.CollectorID() != tc.collector || a.UserID() != tc.user {
				t.Errorf("actor = (%v, %v, %v), want (%v, %v, %v)",
					a.Role(), a.CollectorID(), a.UserID(), tc.role, tc.collector, tc.user)
			}

			if err := a.RequireStaff(); !errors.Is(err, tc.staffErr) {
				t.Errorf("RequireStaff() = %v, want %v", err, tc.staffErr)
			}
		})
	}
}
