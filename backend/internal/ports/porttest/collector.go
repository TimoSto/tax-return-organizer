// Package porttest contains contract tests that every adapter of a port must pass.
package porttest

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

// CollectorRepositoryFactory returns an empty repository for the given test.
type CollectorRepositoryFactory func(t *testing.T) ports.CollectorRepository

// TestCollectorRepository runs the contract tests against repositories created by newRepo.
func TestCollectorRepository(t *testing.T, newRepo CollectorRepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	alice := domain.Collector{ID: uuid.New(), Name: "Alice", Email: "alice@example.com"}
	bob := domain.Collector{ID: uuid.New(), Name: "Bob", Email: "bob@example.com"}
	aliceUpdated := domain.Collector{ID: alice.ID, Name: "Alice Updated", Email: "alice.new@example.com"}

	for _, tc := range []struct {
		name   string
		saves  []domain.Collector
		getID  uuid.UUID
		exp    *domain.Collector
		expErr error
	}{
		{
			name:  "Get saved collector",
			saves: []domain.Collector{alice},
			getID: alice.ID,
			exp:   &alice,
		},
		{
			name:   "Get unknown collector",
			saves:  []domain.Collector{alice},
			getID:  uuid.New(),
			expErr: domain.ErrCollectorNotFound,
		},
		{
			name:   "Get from empty repository",
			getID:  alice.ID,
			expErr: domain.ErrCollectorNotFound,
		},
		{
			name:  "Save again overwrites existing collector",
			saves: []domain.Collector{alice, aliceUpdated},
			getID: alice.ID,
			exp:   &aliceUpdated,
		},
		{
			name:  "Other collectors are left untouched",
			saves: []domain.Collector{alice, bob},
			getID: bob.ID,
			exp:   &bob,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t)

			for _, c := range tc.saves {
				if err := repo.SaveCollector(ctx, &c); err != nil {
					t.Fatalf("SaveCollector() unexpected error: %v", err)
				}
			}

			got, err := repo.GetCollector(ctx, tc.getID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("GetCollector() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got); diff != "" {
				t.Errorf("GetCollector() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
