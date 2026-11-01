// Package porttest contains contract tests that every adapter of a port must pass.
package porttest

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

// TaxPracticeRepositoryFactory returns an empty repository for the given test.
type TaxPracticeRepositoryFactory func(t *testing.T) outbound.TaxPracticeRepository

// TestTaxPracticeRepository runs the contract tests against repositories created by newRepo.
func TestTaxPracticeRepository(t *testing.T, newRepo TaxPracticeRepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	alice := domaintest.TaxPractice(domaintest.TaxPracticeID(), "Alice", "alice@example.com")
	bob := domaintest.TaxPractice(domaintest.TaxPracticeID(), "Bob", "bob@example.com")
	aliceClone := domaintest.TaxPractice(alice.ID(), "Alice Clone", "clone@example.com")

	for _, tc := range []struct {
		name    string
		saves   []*domain.TaxPractice
		getID   domain.TaxPracticeID
		exp     *domain.TaxPractice
		saveErr error
		expErr  error
	}{
		{
			name:  "Get saved tax practice",
			saves: []*domain.TaxPractice{alice},
			getID: alice.ID(),
			exp:   alice,
		},
		{
			name:   "Get unknown tax practice",
			saves:  []*domain.TaxPractice{alice},
			getID:  domaintest.TaxPracticeID(),
			expErr: domain.ErrTaxPracticeNotFound,
		},
		{
			name:   "Get from empty repository",
			getID:  alice.ID(),
			expErr: domain.ErrTaxPracticeNotFound,
		},
		{
			name:    "Save with taken id is rejected and keeps the original",
			saves:   []*domain.TaxPractice{alice, aliceClone},
			saveErr: domain.ErrTaxPracticeIDTaken,
			getID:   alice.ID(),
			exp:     alice,
		},
		{
			name:  "Other tax practices are left untouched",
			saves: []*domain.TaxPractice{alice, bob},
			getID: bob.ID(),
			exp:   bob,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newRepo(t)

			for _, c := range tc.saves {
				if err := repo.SaveTaxPractice(ctx, c); err != nil && !errors.Is(err, tc.saveErr) {
					t.Fatalf("SaveTaxPractice() error = %v, want nil or %v", err, tc.saveErr)
				}
			}

			got, err := repo.GetTaxPractice(ctx, tc.getID)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("GetTaxPractice() error = %v, wantErr %v", err, tc.expErr)
			}

			if diff := cmp.Diff(tc.exp, got, domaintest.CmpOptions()); diff != "" {
				t.Errorf("GetTaxPractice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
