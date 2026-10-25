// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
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
	tx      port.Transactor
	catalog port.PersonaTemplateCatalog
	ids     IDGenerator
	// collectorIDs draws the collector id, it is ids unless WithCollectorIDs replaces it.
	collectorIDs port.CollectorIDGenerator
}

// IDGenerator creates the ids of the new collector, its personas and their items.
type IDGenerator interface {
	port.CollectorIDGenerator
	port.PersonaIDGenerator
	port.ItemIDGenerator
}

// Option customizes a Usecase.
type Option func(*Usecase)

// WithCollectorIDs replaces how the collector id is drawn, e.g. to seed a collector with a well-known id.
func WithCollectorIDs(ids port.CollectorIDGenerator) Option {
	return func(u *Usecase) { u.collectorIDs = ids }
}

// New returns a use case that stores collectors, together with the predefined personas from catalog, through tx.
// ids creates the new ids.
func New(tx port.Transactor, catalog port.PersonaTemplateCatalog, ids IDGenerator, opts ...Option) *Usecase {
	u := &Usecase{tx: tx, catalog: catalog, ids: ids, collectorIDs: ids}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Execute creates and persists a collector with the predefined personas in the requested locale and returns its id.
// The collector and its personas are saved in one transaction: on an error nothing is stored, so a retry is safe.
func (u *Usecase) Execute(ctx context.Context, cmd Command) (domain.CollectorID, error) {
	locale, err := personatemplate.ParseLocale(cmd.Locale)
	if err != nil {
		return "", err
	}

	// One transaction per attempt: a taken id is a unique violation, which aborts the transaction in Postgres.
	for range maxIDAttempts {
		id, err := u.collectorIDs.NewCollectorID()
		if err != nil {
			return "", err
		}

		c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
		if err != nil {
			return "", err
		}

		personas, err := predefinedPersonas(u.catalog, u.ids, id, locale)
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
		if err := repos.Personas.For(c.ID()).SavePersona(ctx, p); err != nil {
			return fmt.Errorf("save persona %q: %w", p.Name(), err)
		}
	}

	return nil
}

// predefinedPersonas instantiates a persona per template for the collector.
func predefinedPersonas(catalog port.PersonaTemplateCatalog, ids personatemplate.IDs, id domain.CollectorID, locale personatemplate.Locale) ([]*domain.Persona, error) {
	templates := catalog.PersonaTemplates()
	personas := make([]*domain.Persona, 0, len(templates))

	for _, tmpl := range templates {
		p, err := tmpl.Instantiate(ids, id, locale)
		if err != nil {
			return nil, fmt.Errorf("instantiate persona %q: %w", tmpl.Name.For(personatemplate.DefaultLocale), err)
		}

		personas = append(personas, p)
	}

	return personas, nil
}
