package persona

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// The table does not restrict kind and recurrence, so reading must reject values the domain does not know.
func TestItemRow_ToDomain_RejectsUnknownValues(t *testing.T) {
	for _, tc := range []struct {
		name, kind, recurrence string
		doc, value             []byte
		expErr                 error
	}{
		{"unknown kind", "documnet", "monthly", []byte(`{}`), nil, errUnknownValue},
		{"unknown recurrence", "document", "weekly", []byte(`{}`), nil, errUnknownValue},
		{"document without document spec", "document", "monthly", nil, nil, errUnknownValue},
		{"document with value spec", "document", "monthly", []byte(`{}`), []byte(`{}`), errUnknownValue},
		{"data without value spec", "data", "monthly", nil, nil, errUnknownValue},
		{"data with document spec", "data", "monthly", []byte(`{}`), []byte(`{}`), errUnknownValue},
		{"valid document", "document", "monthly", []byte(`{}`), nil, nil},
		{"valid data", "data", "monthly", nil, []byte(`{"type":"number"}`), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := itemRow{
				ID: uuid.New(), Title: "Payslip", Kind: tc.kind, Recurrence: tc.recurrence,
				DocumentSpec: tc.doc, ValueSpec: tc.value,
			}

			_, err := row.toDomain()
			if !errors.Is(err, tc.expErr) {
				t.Errorf("toDomain() error = %v, want %v", err, tc.expErr)
			}
		})
	}
}
