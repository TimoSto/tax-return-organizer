package httpx

import (
	"slices"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

// problemCodes maps every domain error to the stable code clients see in Problem.Code. The codes are part of the
// API contract: clients switch on them and translate them, so never change one. A new domain error needs an entry
// here (TestProblemCodes fails otherwise) and the code in the Problem.code enum of api/openapi.yaml
// (TestOpenAPI_ProblemCodes fails otherwise).
var problemCodes = map[*domain.Error]string{
	domain.ErrCollectorIDTaken:          "collector_id_taken",
	domain.ErrCollectorNotFound:         "collector_not_found",
	domain.ErrDuplicateMetadataField:    "duplicate_metadata_field",
	domain.ErrForbidden:                 "forbidden",
	domain.ErrInvalidCollectorID:        "invalid_collector_id",
	domain.ErrInvalidEmail:              "invalid_email",
	domain.ErrInvalidItemID:             "invalid_item_id",
	domain.ErrInvalidPersonaID:          "invalid_persona_id",
	domain.ErrInvalidRecurrence:         "invalid_recurrence",
	domain.ErrInvalidValueType:          "invalid_value_type",
	domain.ErrItemIDTaken:               "item_id_taken",
	domain.ErrItemNotFound:              "item_not_found",
	domain.ErrItemTitleRequired:         "item_title_required",
	domain.ErrItemTitleTooLong:          "item_title_too_long",
	domain.ErrMetadataFieldNameRequired: "metadata_field_name_required",
	domain.ErrMetadataFieldNameTooLong:  "metadata_field_name_too_long",
	domain.ErrNameRequired:              "collector_name_required",
	domain.ErrNameTooLong:               "collector_name_too_long",
	domain.ErrPersonaDescriptionTooLong: "persona_description_too_long",
	domain.ErrPersonaNameRequired:       "persona_name_required",
	domain.ErrPersonaNameTooLong:        "persona_name_too_long",
	domain.ErrPersonaNotFound:           "persona_not_found",
	domain.ErrSpecMismatch:              "item_spec_mismatch",
	domain.ErrSpecRequired:              "item_spec_required",
	domain.ErrTooManyItems:              "too_many_items",
	domain.ErrTooManyMetadataFields:     "too_many_metadata_fields",
	domain.ErrUnitTooLong:               "unit_too_long",
	personatemplate.ErrInvalidLocale:    "invalid_locale",
}

// codeFor returns the code of a domain error.
func codeFor(err *domain.Error) (string, bool) {
	code, ok := problemCodes[err]

	return code, ok
}

// ProblemCodes returns all codes a Problem can carry, sorted: the codes of the domain errors and the ones the
// adapter produces itself.
func ProblemCodes() []string {
	codes := []string{CodeInternal, CodeInvalidBody, CodeBodyTooLarge, CodeUnsupportedMedia}

	for _, code := range problemCodes {
		codes = append(codes, code)
	}

	slices.Sort(codes)

	return codes
}
