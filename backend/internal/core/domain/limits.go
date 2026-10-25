package domain

import "unicode/utf8"

// Upper bounds for user-provided texts and collections. Texts are counted in runes after trimming. They keep one
// request from storing unbounded data; the HTTP body cap is only a coarse outer limit.
const (
	MaxCollectorNameLength      = 200
	MaxEmailLength              = 254 // the longest address SMTP allows
	MaxPersonaNameLength        = 200
	MaxPersonaDescriptionLength = 2000
	MaxItemsPerPersona          = 100
	MaxItemTitleLength          = 200
	MaxMetadataFields           = 50
	MaxMetadataFieldNameLength  = 200
	MaxUnitLength               = 50
)

// tooLong reports whether s has more than limit runes.
func tooLong(s string, limit int) bool {
	return utf8.RuneCountInString(s) > limit
}
