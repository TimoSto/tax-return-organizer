package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrPersonaNameRequired is returned for a blank persona name.
	ErrPersonaNameRequired = invalid("persona name is required")
	// ErrPersonaNotFound is returned when a collector has no persona with the requested id.
	ErrPersonaNotFound = notFound("persona not found")
	// ErrPersonaNameTooLong is returned for a persona name longer than MaxPersonaNameLength.
	ErrPersonaNameTooLong = invalid("persona name is too long")
	// ErrPersonaDescriptionTooLong is returned for a persona description longer than MaxPersonaDescriptionLength.
	ErrPersonaDescriptionTooLong = invalid("persona description is too long")
	// ErrTooManyItems is returned when a persona would get more than MaxItemsPerPersona item definitions.
	ErrTooManyItems = conflict("persona has too many item definitions")
	// ErrItemNotFound is returned when a persona has no item definition with the requested id.
	ErrItemNotFound = notFound("item definition not found")
)

// ErrDuplicateItemID is returned when an item definition id is used twice within one persona. Item ids are created by
// the server, so this is deliberately not a *Error: it signals a bug or corrupt data (a 500), not a client mistake.
var ErrDuplicateItemID = errors.New("duplicate item definition id")

// ErrItemIDTaken is returned when storing an item definition whose id its persona already uses. Item ids are random
// and unique per persona only, so a clash is a bug, not a client error: it is a plain error, not a *Error.
var ErrItemIDTaken = errors.New("item definition id already taken")

// Persona groups the items a client has to provide if they do something tax relevant, e.g. "Employed".
// It is the aggregate root of its item definitions: they are only read and changed through the persona.
// Create personas with NewPersona or RehydratePersona, so they are always valid.
type Persona struct {
	id          PersonaID
	collectorID CollectorID
	name        string
	description string
	// items are in the order the collector arranged them.
	items []ItemDefinition
}

// NewPersona validates the input and returns a persona with the given id and no items.
func NewPersona(id PersonaID, collectorID CollectorID, name, description string) (*Persona, error) {
	return RehydratePersona(id, collectorID, name, description, nil)
}

// RehydratePersona rebuilds a persona that was stored earlier, e.g. when a repository loads it. It applies the same
// rules as NewPersona, so corrupt stored data is reported instead of becoming an invalid persona. The items are
// copied.
func RehydratePersona(
	id PersonaID, collectorID CollectorID, name, description string, items []ItemDefinition,
) (*Persona, error) {
	if collectorID == (CollectorID{}) {
		return nil, ErrInvalidCollectorID
	}

	name, description = strings.TrimSpace(name), strings.TrimSpace(description)

	if name == "" {
		return nil, ErrPersonaNameRequired
	}

	if tooLong(name, MaxPersonaNameLength) {
		return nil, ErrPersonaNameTooLong
	}

	if tooLong(description, MaxPersonaDescriptionLength) {
		return nil, ErrPersonaDescriptionTooLong
	}

	if len(items) > MaxItemsPerPersona {
		return nil, ErrTooManyItems
	}

	p := &Persona{id: id, collectorID: collectorID, name: name, description: description}
	for _, item := range items {
		if p.itemIndex(item.id) >= 0 {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateItemID, item.id)
		}

		p.items = append(p.items, cloneItem(item))
	}

	return p, nil
}

// ID returns the id of the persona.
func (p *Persona) ID() PersonaID { return p.id }

// CollectorID returns the collector the persona belongs to.
func (p *Persona) CollectorID() CollectorID { return p.collectorID }

// Name returns the name of the persona.
func (p *Persona) Name() string { return p.name }

// Description returns the description of the persona, which may be empty.
func (p *Persona) Description() string { return p.description }

// Items returns the item definitions in the order the collector arranged them. The result is a copy:
// use AddItem, UpdateItem and RemoveItem to change the persona.
func (p *Persona) Items() []ItemDefinition {
	if p.items == nil {
		return nil
	}

	items := make([]ItemDefinition, len(p.items))
	for i, item := range p.items {
		items[i] = cloneItem(item)
	}

	return items
}

// AddItem validates the input, appends a new item definition and returns it.
func (p *Persona) AddItem(id ItemID, title string, recurrence Recurrence, spec ItemSpec) (ItemDefinition, error) {
	if len(p.items) >= MaxItemsPerPersona {
		return ItemDefinition{}, ErrTooManyItems
	}

	if p.itemIndex(id) >= 0 {
		return ItemDefinition{}, fmt.Errorf("%w: %s", ErrDuplicateItemID, id)
	}

	item, err := NewItemDefinition(id, title, recurrence, spec)
	if err != nil {
		return ItemDefinition{}, err
	}

	p.items = append(p.items, item)

	return cloneItem(item), nil
}

// UpdateItem replaces title, recurrence and spec of an item. The id and the kind stay the same,
// so the spec has to be of the item's kind.
func (p *Persona) UpdateItem(id ItemID, title string, recurrence Recurrence, spec ItemSpec) error {
	i := p.itemIndex(id)
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrItemNotFound, id)
	}

	if spec != nil && spec.Kind() != p.items[i].Kind() {
		return fmt.Errorf("%w: a %s item cannot get a %s spec", ErrSpecMismatch, p.items[i].Kind(), spec.Kind())
	}

	updated, err := NewItemDefinition(id, title, recurrence, spec)
	if err != nil {
		return err
	}

	p.items[i] = updated

	return nil
}

// RemoveItem deletes an item definition.
func (p *Persona) RemoveItem(id ItemID) error {
	i := p.itemIndex(id)
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrItemNotFound, id)
	}

	p.items = slices.Delete(p.items, i, i+1)

	return nil
}

func (p *Persona) itemIndex(id ItemID) int {
	return slices.IndexFunc(p.items, func(item ItemDefinition) bool { return item.id == id })
}

func cloneItem(item ItemDefinition) ItemDefinition {
	item.spec = cloneSpec(item.spec)

	return item
}
