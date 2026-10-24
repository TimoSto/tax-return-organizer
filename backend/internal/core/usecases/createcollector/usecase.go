// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

var _ ports.CreateCollectorPort = (*Usecase)(nil).Execute

// maxIDAttempts bounds the retries when a randomly drawn collector id is already taken.
const maxIDAttempts = 5

// Usecase implements ports.CreateCollector.
type Usecase struct {
	repo     ports.CollectorRepository
	personas ports.PersonaRepository
	newID    func() (domain.CollectorID, error)
}

// Option customizes a Usecase.
type Option func(*Usecase)

// WithIDGenerator replaces the random collector id generation, e.g. to seed a collector with a well-known id.
func WithIDGenerator(newID func() (domain.CollectorID, error)) Option {
	return func(u *Usecase) { u.newID = newID }
}

// New returns a use case that stores collectors in repo and gives each the predefined personas in personas.
func New(repo ports.CollectorRepository, personas ports.PersonaRepository, opts ...Option) *Usecase {
	u := &Usecase{repo: repo, personas: personas, newID: domain.NewCollectorID}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Execute creates and persists a collector with the predefined personas in the requested locale and returns its id.
func (u *Usecase) Execute(ctx context.Context, cmd ports.CreateCollectorCmd) (domain.CollectorID, error) {
	locale, err := domain.ParseLocale(cmd.Locale)
	if err != nil {
		return "", err
	}

	for range maxIDAttempts {
		id, err := u.newID()
		if err != nil {
			return "", err
		}

		c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
		if err != nil {
			return "", err
		}

		err = u.repo.SaveCollector(ctx, c)
		if errors.Is(err, domain.ErrCollectorIDTaken) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("save collector: %w", err)
		}

		u.addPredefinedPersonas(ctx, c.ID, locale)

		return c.ID, nil
	}

	return "", fmt.Errorf("save collector: no free id after %d attempts", maxIDAttempts)
}

// addPredefinedPersonas gives the collector a persona per template. The collector and its personas are saved
// separately, so a failure here is logged and not returned: the collector exists and can add personas itself,
// while failing the request would make a retry create a second collector.
func (u *Usecase) addPredefinedPersonas(ctx context.Context, id domain.CollectorID, locale domain.Locale) {
	for _, tmpl := range domain.PersonaTemplates() {
		p, err := tmpl.Instantiate(id, locale)
		if err == nil {
			err = u.personas.SavePersona(ctx, p)
		}

		if err != nil {
			slog.ErrorContext(ctx, "add predefined persona", "collector_id", id, "persona", tmpl.Name.For(domain.DefaultLocale), "error", err)
		}
	}
}
