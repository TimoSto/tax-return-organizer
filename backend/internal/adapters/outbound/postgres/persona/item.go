package persona

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// itemRow mirrors the item_definitions table. The specs are stored as JSONB.
type itemRow struct {
	ID           uuid.UUID
	Title        string
	Kind         string
	Recurrence   string
	DocumentSpec []byte
	ValueSpec    []byte
}

// nullableItemRow scans the item columns of a LEFT JOIN, which are all NULL for a persona without items.
// Only ID is checked for NULL; the other columns are NOT NULL (or nullable specs) once an item exists.
type nullableItemRow struct {
	ID           *uuid.UUID
	Title        *string
	Kind         *string
	Recurrence   *string
	DocumentSpec []byte
	ValueSpec    []byte
}

// toRow must only be called when ID is not nil.
func (r nullableItemRow) toRow() itemRow {
	return itemRow{
		ID: *r.ID, Title: *r.Title, Kind: *r.Kind, Recurrence: *r.Recurrence,
		DocumentSpec: r.DocumentSpec, ValueSpec: r.ValueSpec,
	}
}

type valueSpecJSON struct {
	Type     string `json:"type"`
	Unit     string `json:"unit"`
	Required bool   `json:"required"`
}

type metadataFieldJSON struct {
	Name string        `json:"name"`
	Spec valueSpecJSON `json:"spec"`
}

type documentSpecJSON struct {
	MetadataFields []metadataFieldJSON `json:"metadataFields"`
}

func newItemRow(item domain.ItemDefinition) (itemRow, error) {
	row := itemRow{ID: item.ID, Title: item.Title, Kind: string(item.Kind), Recurrence: string(item.Recurrence)}

	var err error

	if item.DocumentSpec != nil {
		if row.DocumentSpec, err = json.Marshal(newDocumentSpecJSON(*item.DocumentSpec)); err != nil {
			return itemRow{}, fmt.Errorf("encode document spec: %w", err)
		}
	}

	if item.ValueSpec != nil {
		if row.ValueSpec, err = json.Marshal(newValueSpecJSON(*item.ValueSpec)); err != nil {
			return itemRow{}, fmt.Errorf("encode value spec: %w", err)
		}
	}

	return row, nil
}

func (r itemRow) toDomain(personaID uuid.UUID) (domain.ItemDefinition, error) {
	item := domain.ItemDefinition{
		ID:         r.ID,
		PersonaID:  personaID,
		Title:      r.Title,
		Kind:       domain.Kind(r.Kind),
		Recurrence: domain.Recurrence(r.Recurrence),
	}

	if r.DocumentSpec != nil {
		var spec documentSpecJSON
		if err := json.Unmarshal(r.DocumentSpec, &spec); err != nil {
			return domain.ItemDefinition{}, fmt.Errorf("decode document spec of item %s: %w", r.ID, err)
		}

		item.DocumentSpec = spec.toDomain()
	}

	if r.ValueSpec != nil {
		var spec valueSpecJSON
		if err := json.Unmarshal(r.ValueSpec, &spec); err != nil {
			return domain.ItemDefinition{}, fmt.Errorf("decode value spec of item %s: %w", r.ID, err)
		}

		item.ValueSpec = spec.toDomain()
	}

	return item, nil
}

func newValueSpecJSON(s domain.ValueSpec) valueSpecJSON {
	return valueSpecJSON{Type: string(s.Type), Unit: s.Unit, Required: s.Required}
}

func (s valueSpecJSON) toDomain() *domain.ValueSpec {
	return &domain.ValueSpec{Type: domain.ValueType(s.Type), Unit: s.Unit, Required: s.Required}
}

func newDocumentSpecJSON(s domain.DocumentSpec) documentSpecJSON {
	// Always an array, never null, so the stored JSON has one shape.
	fields := make([]metadataFieldJSON, 0, len(s.MetadataFields))
	for _, f := range s.MetadataFields {
		fields = append(fields, metadataFieldJSON{Name: f.Name, Spec: newValueSpecJSON(f.ValueSpec)})
	}

	return documentSpecJSON{MetadataFields: fields}
}

func (s documentSpecJSON) toDomain() *domain.DocumentSpec {
	spec := &domain.DocumentSpec{}

	for _, f := range s.MetadataFields {
		spec.MetadataFields = append(spec.MetadataFields, domain.MetadataField{Name: f.Name, ValueSpec: *f.Spec.toDomain()})
	}

	return spec
}
