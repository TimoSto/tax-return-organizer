// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port"
)

// Command is the input of the use case.
type Command struct {
	Name  string
	Email string
	// Locale is the language of the predefined personas, empty for the default.
	Locale string
}

// maxIDAttempts bounds the retries when a randomly drawn collector id is already taken.
const maxIDAttempts = 5

// Usecase is the createcollector use case.
type Usecase struct {
	tx  port.Transactor
	ids domain.IDGenerator
}

// Option customizes a Usecase.
type Option func(*Usecase)

// WithIDGenerator replaces the random id generation, e.g. to seed a collector with a well-known id.
func WithIDGenerator(ids domain.IDGenerator) Option {
	return func(u *Usecase) { u.ids = ids }
}

// New returns a use case that stores collectors, together with their predefined personas, through tx.
func New(tx port.Transactor, opts ...Option) *Usecase {
	u := &Usecase{tx: tx, ids: domain.RandomIDs{}}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Execute creates and persists a collector with the predefined personas in the requested locale and returns its id.
// The collector and its personas are saved in one transaction: on an error nothing is stored, so a retry is safe.
func (u *Usecase) Execute(ctx context.Context, cmd Command) (domain.CollectorID, error) {
	locale, err := domain.ParseLocale(cmd.Locale)
	if err != nil {
		return "", err
	}

	// One transaction per attempt: a taken id is a unique violation, which aborts the transaction in Postgres.
	for range maxIDAttempts {
		id, err := u.ids.NewCollectorID()
		if err != nil {
			return "", err
		}

		c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
		if err != nil {
			return "", err
		}

		personas, err := predefinedPersonas(u.ids, id, locale)
		if err != nil {
			return "", err
		}

		err = u.tx.InTx(ctx, func(ctx context.Context, repos port.Repositories) error {
			return save(ctx, repos, c, personas)
		})
		if errors.Is(err, domain.ErrCollectorIDTaken) {
			continue
		}

		if err != nil {
			return "", err
		}

		return c.ID(), nil
	}

	return "", fmt.Errorf("save collector: no free id after %d attempts", maxIDAttempts)
}

func save(ctx context.Context, repos port.Repositories, c *domain.Collector, personas []*domain.Persona) error {
	if err := repos.Collectors.SaveCollector(ctx, c); err != nil {
		return fmt.Errorf("save collector: %w", err)
	}

	for _, p := range personas {
		if err := repos.Personas.SavePersona(ctx, p); err != nil {
			return fmt.Errorf("save persona %q: %w", p.Name(), err)
		}
	}

	return nil
}

// predefinedPersonas instantiates a persona per template for the collector.
func predefinedPersonas(ids domain.IDGenerator, id domain.CollectorID, locale domain.Locale) ([]*domain.Persona, error) {
	templates := domain.PersonaTemplates()
	personas := make([]*domain.Persona, 0, len(templates))

	for _, tmpl := range templates {
		p, err := tmpl.Instantiate(ids, id, locale)
		if err != nil {
			return nil, fmt.Errorf("instantiate persona %q: %w", tmpl.Name.For(domain.DefaultLocale), err)
		}

		personas = append(personas, p)
	}

	return personas, nil
}
