package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// Every sentinel must be a classified *domain.Error, otherwise adapters would answer it with a 500.
func TestSentinelsAreClassified(t *testing.T) {
	for _, err := range []error{
		domain.ErrCollectorNotFound, domain.ErrCollectorIDTaken, domain.ErrNameRequired, domain.ErrInvalidCollectorID, domain.ErrInvalidPersonaID, domain.ErrInvalidItemID,
		domain.ErrInvalidEmail, domain.ErrPersonaNameRequired, domain.ErrPersonaNotFound,
		domain.ErrItemNotFound, domain.ErrItemIDTaken, domain.ErrItemTitleRequired,
		domain.ErrInvalidRecurrence, domain.ErrSpecRequired, domain.ErrSpecMismatch, domain.ErrInvalidValueType,
		domain.ErrMetadataFieldNameRequired, domain.ErrDuplicateMetadataField,
		domain.ErrNameTooLong, domain.ErrPersonaNameTooLong, domain.ErrPersonaDescriptionTooLong, domain.ErrTooManyItems,
		domain.ErrItemTitleTooLong, domain.ErrMetadataFieldNameTooLong, domain.ErrTooManyMetadataFields, domain.ErrUnitTooLong,
	} {
		var de *domain.Error
		if !errors.As(err, &de) || de.Kind == 0 {
			t.Errorf("%v is not classified", err)
		}
	}
}
