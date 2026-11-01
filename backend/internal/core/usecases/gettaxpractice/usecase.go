// Package gettaxpractice implements the use case to retrieve the tax practice of the caller.
package gettaxpractice

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Usecase is the gettaxpractice use case.
type Usecase struct {
	repo outbound.TaxPracticeRepository
}

var _ inbound.GetTaxPractice = (*Usecase)(nil)

// New returns a use case that gets tax practices from repo.
func New(repo outbound.TaxPracticeRepository) *Usecase {
	return &Usecase{repo: repo}
}

// Execute retrieves the tax practice the actor belongs to. Staff and clients may, the zero actor has no tax practice
// and gets domain.ErrTaxPracticeNotFound.
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor) (*domain.TaxPractice, error) {
	c, err := u.repo.GetTaxPractice(ctx, actor.TaxPracticeID())
	if err != nil {
		return nil, fmt.Errorf("get tax practice: %w", err)
	}
	return c, nil
}
