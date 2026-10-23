package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var (
	// ErrPersonaNameRequired is returned for a blank persona name.
	ErrPersonaNameRequired = errors.New("persona name is required")
	// ErrPersonaNotFound is returned when a collector has no persona with the requested id.
	ErrPersonaNotFound = errors.New("persona not found")
	// ErrItemNotFound is returned when a persona has no item definition with the requested id.
	ErrItemNotFound = errors.New("item definition not found")
	// ErrItemIDTaken is returned when an item definition id is already used by another persona.
	ErrItemIDTaken = errors.New("item definition id already taken")
	// ErrUnknownCollector is returned when a persona refers to a collector that does not exist.
	ErrUnknownCollector = errors.New("collector does not exist")
)

// Persona groups the items a client has to provide if they do something tax relevant, e.g. "Employed".
// It is the aggregate root of its item definitions: they are only read and changed through the persona.
type Persona struct {
	ID          uuid.UUID
	CollectorID CollectorID
	Name        string
	Description string
	// Items are in the order the collector arranged them.
	Items []ItemDefinition
}

// NewPersona validates the input and returns a persona with a fresh id.
func NewPersona(collectorID CollectorID, name, description string) (*Persona, error) {
	name, description = strings.TrimSpace(name), strings.TrimSpace(description)

	if name == "" {
		return nil, ErrPersonaNameRequired
	}

	return &Persona{ID: uuid.New(), CollectorID: collectorID, Name: name, Description: description}, nil
}

// AddItem validates the input, appends a new item definition and returns a copy of it.
// Changes to the copy do not affect the persona, use UpdateItem for that.
func (p *Persona) AddItem(
	title string,
	kind Kind,
	recurrence Recurrence,
	documentSpec *DocumentSpec,
	valueSpec *ValueSpec,
) (ItemDefinition, error) {
	item, err := NewItemDefinition(p.ID, title, kind, recurrence, documentSpec, valueSpec)
	if err != nil {
		return ItemDefinition{}, err
	}

	p.Items = append(p.Items, *item)

	return *item, nil
}

// UpdateItem replaces title, recurrence and spec of an item. The id and the kind stay the same,
// so the spec has to match the item's kind.
func (p *Persona) UpdateItem(
	id uuid.UUID,
	title string,
	recurrence Recurrence,
	documentSpec *DocumentSpec,
	valueSpec *ValueSpec,
) error {
	i := p.itemIndex(id)
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrItemNotFound, id)
	}

	updated, err := buildItemDefinition(id, p.ID, title, p.Items[i].Kind, recurrence, documentSpec, valueSpec)
	if err != nil {
		return err
	}

	p.Items[i] = *updated

	return nil
}

// RemoveItem deletes an item definition.
func (p *Persona) RemoveItem(id uuid.UUID) error {
	i := p.itemIndex(id)
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrItemNotFound, id)
	}

	p.Items = slices.Delete(p.Items, i, i+1)

	return nil
}

func (p *Persona) itemIndex(id uuid.UUID) int {
	return slices.IndexFunc(p.Items, func(item ItemDefinition) bool { return item.ID == id })
}
