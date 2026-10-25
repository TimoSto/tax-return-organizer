// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"errors"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// maxIDAttempts bounds the retries when a randomly drawn collector id is already taken.
const maxIDAttempts = 5

// Usecase is the createcollector use case.
type Usecase struct {
	tx        outbound.Transactor
	templates []personatemplate.PersonaTemplate
	ids       outbound.IDGenerator
}

var _ inbound.CreateCollector = (*Usecase)(nil)

// New returns a use case that stores collectors, together with a persona per template, through tx.
// ids creates the new ids.
func New(tx outbound.Transactor, templates []personatemplate.PersonaTemplate, ids outbound.IDGenerator) *Usecase {
	return &Usecase{tx: tx, templates: templates, ids: ids}
}

// Execute creates and persists a collector with the predefined personas in the requested locale and returns its id.
// The collector and its personas are saved in one transaction: on an error nothing is stored, so a retry is safe.
func (u *Usecase) Execute(ctx context.Context, cmd inbound.CreateCollectorCommand) (domain.CollectorID, error) {
	locale, err := personatemplate.ParseLocale(cmd.Locale)
	if err != nil {
		return "", err
	}

	// The persona and item ids do not depend on the collector id, so they are drawn once and survive a retry.
	ids, err := u.newPersonaIDs()
	if err != nil {
		return "", err
	}

	return u.withFreeCollectorID(func(id domain.CollectorID) error {
		c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
		if err != nil {
			return err
		}

		personas, err := u.personas(id, locale, ids)
		if err != nil {
			return err
		}

		return u.tx.InTx(ctx, func(ctx context.Context, repos outbound.Repositories) error {
			return save(ctx, repos, c, personas)
		})
	})
}

// withFreeCollectorID calls create with a freshly drawn collector id until the id is not taken, and returns it.
// Collector ids are short, so a draw can hit an existing one. create must be safe to repeat: one transaction per
// attempt, since a taken id is a unique violation, which aborts the transaction in Postgres.
func (u *Usecase) withFreeCollectorID(create func(domain.CollectorID) error) (domain.CollectorID, error) {
	for range maxIDAttempts {
		id, err := u.ids.NewCollectorID()
		if err != nil {
			return "", err
		}

		err = create(id)
		if errors.Is(err, domain.ErrCollectorIDTaken) {
			continue
		}

		if err != nil {
			return "", err
		}

		return id, nil
	}

	return "", fmt.Errorf("save collector: no free id after %d attempts", maxIDAttempts)
}

// personaIDs are the ids of one persona from a template and of its items, in the order of the template's items.
type personaIDs struct {
	persona domain.PersonaID
	items   []domain.ItemID
}

func (u *Usecase) newPersonaIDs() ([]personaIDs, error) {
	all := make([]personaIDs, len(u.templates))

	for i, tmpl := range u.templates {
		persona, err := u.ids.NewPersonaID()
		if err != nil {
			return nil, err
		}

		items := make([]domain.ItemID, len(tmpl.Items))
		for j := range items {
			if items[j], err = u.ids.NewItemID(); err != nil {
				return nil, err
			}
		}

		all[i] = personaIDs{persona: persona, items: items}
	}

	return all, nil
}

// personas instantiates a persona per template for the collector.
func (u *Usecase) personas(id domain.CollectorID, locale personatemplate.Locale, ids []personaIDs) ([]*domain.Persona, error) {
	personas := make([]*domain.Persona, 0, len(u.templates))

	for i, tmpl := range u.templates {
		p, err := tmpl.Instantiate(id, locale, ids[i].persona, ids[i].items)
		if err != nil {
			return nil, fmt.Errorf("instantiate persona %q: %w", tmpl.Name.For(personatemplate.DefaultLocale), err)
		}

		personas = append(personas, p)
	}

	return personas, nil
}

func save(ctx context.Context, repos outbound.Repositories, c *domain.Collector, personas []*domain.Persona) error {
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
