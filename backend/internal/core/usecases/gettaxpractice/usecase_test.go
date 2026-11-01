package gettaxpractice_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/gettaxpractice"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var (
	// testTaxPracticeID is the tax practice the tests act for.
	testTaxPracticeID = domaintest.TaxPracticeID()
)

// taxPracticeStub returns the tax practices it holds and domain.ErrTaxPracticeNotFound for any other id.
type taxPracticeStub map[domain.TaxPracticeID]*domain.TaxPractice

func (s taxPracticeStub) GetTaxPractice(_ context.Context, id domain.TaxPracticeID) (*domain.TaxPractice, error) {
	c, ok := s[id]
	if !ok {
		return nil, domain.ErrTaxPracticeNotFound
	}

	return c, nil
}

func (taxPracticeStub) SaveTaxPractice(context.Context, *domain.TaxPractice) error {
	return errors.New("taxPracticeStub: unexpected SaveTaxPractice")
}

type failingRepo struct {
	outbound.TaxPracticeRepository
	err error
}

func (r failingRepo) GetTaxPractice(context.Context, domain.TaxPracticeID) (*domain.TaxPractice, error) {
	return nil, r.err
}

func TestExecute_Success(t *testing.T) {
	want := domaintest.TaxPractice(testTaxPracticeID, "Jane", "jane@example.com")
	repo := taxPracticeStub{want.ID(): want}

	got, err := gettaxpractice.New(repo).Execute(context.Background(), domaintest.Staff(want.ID()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *got != *want {
		t.Errorf("tax practice = %+v, want %+v", got, want)
	}
}

// Clients may see the tax practice they belong to.
func TestExecute_Client(t *testing.T) {
	want := domaintest.TaxPractice(testTaxPracticeID, "Jane", "jane@example.com")
	repo := taxPracticeStub{want.ID(): want}

	got, err := gettaxpractice.New(repo).Execute(context.Background(), domaintest.Client(want.ID()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *got != *want {
		t.Errorf("tax practice = %+v, want %+v", got, want)
	}
}

// The zero actor belongs to no tax practice.
func TestExecute_ZeroActor(t *testing.T) {
	_, err := gettaxpractice.New(taxPracticeStub{}).Execute(context.Background(), domain.Actor{})
	if !errors.Is(err, domain.ErrTaxPracticeNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrTaxPracticeNotFound)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := gettaxpractice.New(taxPracticeStub{}).Execute(context.Background(), domaintest.Staff(testTaxPracticeID))
	if !errors.Is(err, domain.ErrTaxPracticeNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrTaxPracticeNotFound)
	}

	if got != nil {
		t.Errorf("tax practice = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := gettaxpractice.New(failingRepo{err: boom}).Execute(context.Background(), domaintest.Staff(testTaxPracticeID))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}

	if n := strings.Count(err.Error(), "get tax practice"); n != 1 {
		t.Errorf("err = %q, want the operation name exactly once, got %d", err, n)
	}
}
