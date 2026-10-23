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
		expErr                 error
	}{
		{"unknown kind", "documnet", "monthly", errUnknownValue},
		{"unknown recurrence", "document", "weekly", errUnknownValue},
		{"valid", "document", "monthly", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := itemRow{ID: uuid.New(), Title: "Payslip", Kind: tc.kind, Recurrence: tc.recurrence, DocumentSpec: []byte(`{}`)}

			_, err := row.toDomain(uuid.New())
			if !errors.Is(err, tc.expErr) {
				t.Errorf("toDomain() error = %v, want %v", err, tc.expErr)
			}
		})
	}
}
