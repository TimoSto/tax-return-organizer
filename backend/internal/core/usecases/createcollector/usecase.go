// Package createcollector implements the use case of creating a new collector.
package createcollector

import (
	"context"
	"fmt"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/inbound"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// Repositories are the repositories the use case writes through, inside one transaction.
type Repositories struct {
	Collectors outbound.CollectorRepository
	Personas   outbound.PersonaStore
}

// Usecase is the createcollector use case.
type Usecase struct {
	tx        outbound.Transactor[Repositories]
	templates []personatemplate.PersonaTemplate
	ids       outbound.IDGenerator
}

var _ inbound.CreateCollector = (*Usecase)(nil)

// New returns a use case that stores collectors, together with a persona per template, through tx.
// ids creates the new ids.
func New(tx outbound.Transactor[Repositories], templates []personatemplate.PersonaTemplate, ids outbound.IDGenerator) *Usecase {
	return &Usecase{tx: tx, templates: templates, ids: ids}
}

// Execute creates and persists a collector with the predefined personas in the requested locale and returns its id.
// The collector and its personas are saved in one transaction: on an error nothing is stored, so a retry is safe.
func (u *Usecase) Execute(ctx context.Context, cmd inbound.CreateCollectorCommand) (domain.CollectorID, error) {
	l, err := locale.ParseLocale(cmd.Locale)
	if err != nil {
		return domain.CollectorID{}, err
	}

	id, err := u.ids.NewCollectorID()
	if err != nil {
		return domain.CollectorID{}, err
	}

	ids, err := u.newPersonaIDs()
	if err != nil {
		return domain.CollectorID{}, err
	}

	c, err := domain.NewCollector(id, cmd.Name, cmd.Email)
	if err != nil {
		return domain.CollectorID{}, err
	}

	personas, err := u.personas(id, l, ids)
	if err != nil {
		return domain.CollectorID{}, err
	}

	err = u.tx.InTx(ctx, func(ctx context.Context, repos Repositories) error {
		return save(ctx, repos, c, personas)
	})
	if err != nil {
		return domain.CollectorID{}, err
	}

	return id, nil
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
func (u *Usecase) personas(id domain.CollectorID, l locale.Locale, ids []personaIDs) ([]*domain.Persona, error) {
	personas := make([]*domain.Persona, 0, len(u.templates))

	for i, tmpl := range u.templates {
		p, err := tmpl.Instantiate(id, l, ids[i].persona, ids[i].items)
		if err != nil {
			return nil, fmt.Errorf("instantiate persona %q: %w", tmpl.Name.For(locale.DefaultLocale), err)
		}

		personas = append(personas, p)
	}

	return personas, nil
}

func save(ctx context.Context, repos Repositories, c *domain.Collector, personas []*domain.Persona) error {
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
