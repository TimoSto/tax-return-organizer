package getcollector_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/usecases/getcollector"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/domaintest"
)

var (
	// testCollectorID is the collector the tests act for.
	testCollectorID = domaintest.CollectorID()
)

// collectorStub returns the collectors it holds and domain.ErrCollectorNotFound for any other id.
type collectorStub map[domain.CollectorID]*domain.Collector

func (s collectorStub) GetCollector(_ context.Context, id domain.CollectorID) (*domain.Collector, error) {
	c, ok := s[id]
	if !ok {
		return nil, domain.ErrCollectorNotFound
	}

	return c, nil
}

func (collectorStub) SaveCollector(context.Context, *domain.Collector) error {
	return errors.New("collectorStub: unexpected SaveCollector")
}

type failingRepo struct {
	outbound.CollectorRepository
	err error
}

func (r failingRepo) GetCollector(context.Context, domain.CollectorID) (*domain.Collector, error) {
	return nil, r.err
}

func TestExecute_Success(t *testing.T) {
	want := domaintest.Collector(testCollectorID, "Jane", "jane@example.com")
	repo := collectorStub{want.ID(): want}

	got, err := getcollector.New(repo).Execute(context.Background(), domaintest.Staff(want.ID()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *got != *want {
		t.Errorf("collector = %+v, want %+v", got, want)
	}
}

// Clients may see the collector they belong to.
func TestExecute_Client(t *testing.T) {
	want := domaintest.Collector(testCollectorID, "Jane", "jane@example.com")
	repo := collectorStub{want.ID(): want}

	got, err := getcollector.New(repo).Execute(context.Background(), domaintest.Client(want.ID()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *got != *want {
		t.Errorf("collector = %+v, want %+v", got, want)
	}
}

// The zero actor belongs to no collector.
func TestExecute_ZeroActor(t *testing.T) {
	_, err := getcollector.New(collectorStub{}).Execute(context.Background(), domain.Actor{})
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}
}

func TestExecute_NotFound(t *testing.T) {
	got, err := getcollector.New(collectorStub{}).Execute(context.Background(), domaintest.Staff(testCollectorID))
	if !errors.Is(err, domain.ErrCollectorNotFound) {
		t.Errorf("err = %v, want %v", err, domain.ErrCollectorNotFound)
	}

	if got != nil {
		t.Errorf("collector = %+v, want nil", got)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	boom := errors.New("boom")

	_, err := getcollector.New(failingRepo{err: boom}).Execute(context.Background(), domaintest.Staff(testCollectorID))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}

	if n := strings.Count(err.Error(), "get collector"); n != 1 {
		t.Errorf("err = %q, want the operation name exactly once, got %d", err, n)
	}
}
