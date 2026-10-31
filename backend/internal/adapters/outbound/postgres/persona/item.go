package persona

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/port/outbound"
)

// errUnknownValue marks a stored row with a kind or recurrence the domain does not know, or whose kind and specs
// disagree. It is deliberately not a domain error
// (those are bad requests): bad stored data is a server fault.
var errUnknownValue = errors.New("unknown value in stored item")

// itemRow mirrors the item_definitions table. The specs and instance fields are stored as JSONB.
type itemRow struct {
	ID             uuid.UUID
	Title          string
	Kind           string
	Recurrence     string
	DocumentSpec   []byte
	ValueSpec      []byte
	InstanceFields []byte
}

// nullableItemRow scans the item columns of a LEFT JOIN, which are all NULL for a persona without items.
// Only ID is checked for NULL; the other columns are NOT NULL (or nullable specs) once an item exists.
type nullableItemRow struct {
	ID             *uuid.UUID
	Title          *string
	Kind           *string
	Recurrence     *string
	DocumentSpec   []byte
	ValueSpec      []byte
	InstanceFields []byte
}

// toRow must only be called when ID is not nil.
func (r nullableItemRow) toRow() itemRow {
	return itemRow{
		ID: *r.ID, Title: *r.Title, Kind: *r.Kind, Recurrence: *r.Recurrence,
		DocumentSpec: r.DocumentSpec, ValueSpec: r.ValueSpec, InstanceFields: r.InstanceFields,
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
	row := itemRow{
		ID: uuid.UUID(item.ID()), Title: item.Title(), Kind: string(item.Kind()), Recurrence: string(item.Recurrence()),
	}

	var err error

	if row.InstanceFields, err = json.Marshal(newFieldsJSON(item.InstanceFields())); err != nil {
		return itemRow{}, fmt.Errorf("encode instance fields: %w", err)
	}

	switch spec := item.Spec().(type) {
	case domain.DocumentSpec:
		if row.DocumentSpec, err = json.Marshal(newDocumentSpecJSON(spec)); err != nil {
			return itemRow{}, fmt.Errorf("encode document spec: %w", err)
		}
	case domain.ValueSpec:
		if row.ValueSpec, err = json.Marshal(newValueSpecJSON(spec)); err != nil {
			return itemRow{}, fmt.Errorf("encode value spec: %w", err)
		}
	default:
		return itemRow{}, fmt.Errorf("item %s: %w: spec %T", item.ID(), errUnknownValue, item.Spec())
	}

	return row, nil
}

// toDomain maps the row to the domain type. The table does not restrict kind and recurrence, nor does it tie
// the kind to the spec columns, so unknown values and inconsistent rows are rejected here instead of being
// passed on to the domain.
func (r itemRow) toDomain() (domain.ItemDefinition, error) {
	spec, err := r.spec()
	if err != nil {
		return domain.ItemDefinition{}, err
	}

	recurrence := domain.Recurrence(r.Recurrence)
	if !recurrence.Valid() {
		return domain.ItemDefinition{}, fmt.Errorf("item %s: %w: recurrence %q", r.ID, errUnknownValue, r.Recurrence)
	}

	var instanceFields []metadataFieldJSON
	if err := json.Unmarshal(r.InstanceFields, &instanceFields); err != nil {
		return domain.ItemDefinition{}, fmt.Errorf("decode instance fields of item %s: %w", r.ID, err)
	}

	item, err := domain.RehydrateItemDefinition(domain.ItemID(r.ID), r.Title, recurrence, spec, fieldsToDomain(instanceFields))
	if err != nil {
		return domain.ItemDefinition{}, outbound.CorruptData("item "+r.ID.String(), err)
	}

	return item, nil
}

// spec decodes the spec column that belongs to the row's kind. The other column must be empty.
func (r itemRow) spec() (domain.ItemSpec, error) {
	switch kind := domain.Kind(r.Kind); kind {
	case domain.KindDocument:
		if r.DocumentSpec == nil || r.ValueSpec != nil {
			return nil, fmt.Errorf("item %s: %w: a document item needs exactly a document spec", r.ID, errUnknownValue)
		}

		var spec documentSpecJSON
		if err := json.Unmarshal(r.DocumentSpec, &spec); err != nil {
			return nil, fmt.Errorf("decode document spec of item %s: %w", r.ID, err)
		}

		return spec.toDomain(), nil
	case domain.KindData:
		if r.ValueSpec == nil || r.DocumentSpec != nil {
			return nil, fmt.Errorf("item %s: %w: a data item needs exactly a value spec", r.ID, errUnknownValue)
		}

		var spec valueSpecJSON
		if err := json.Unmarshal(r.ValueSpec, &spec); err != nil {
			return nil, fmt.Errorf("decode value spec of item %s: %w", r.ID, err)
		}

		return spec.toDomain(), nil
	default:
		return nil, fmt.Errorf("item %s: %w: kind %q", r.ID, errUnknownValue, r.Kind)
	}
}

func newValueSpecJSON(s domain.ValueSpec) valueSpecJSON {
	return valueSpecJSON{Type: string(s.Type), Unit: s.Unit, Required: s.Required}
}

func (s valueSpecJSON) toDomain() domain.ValueSpec {
	return domain.ValueSpec{Type: domain.ValueType(s.Type), Unit: s.Unit, Required: s.Required}
}

func newDocumentSpecJSON(s domain.DocumentSpec) documentSpecJSON {
	return documentSpecJSON{MetadataFields: newFieldsJSON(s.MetadataFields)}
}

func (s documentSpecJSON) toDomain() domain.DocumentSpec {
	return domain.DocumentSpec{MetadataFields: fieldsToDomain(s.MetadataFields)}
}

func newFieldsJSON(fs []domain.MetadataField) []metadataFieldJSON {
	// Always an array, never null, so the stored JSON has one shape.
	fields := make([]metadataFieldJSON, 0, len(fs))
	for _, f := range fs {
		fields = append(fields, metadataFieldJSON{Name: f.Name, Spec: newValueSpecJSON(f.ValueSpec)})
	}

	return fields
}

func fieldsToDomain(fs []metadataFieldJSON) []domain.MetadataField {
	var fields []domain.MetadataField
	for _, f := range fs {
		fields = append(fields, domain.MetadataField{Name: f.Name, ValueSpec: f.Spec.toDomain()})
	}

	return fields
}
