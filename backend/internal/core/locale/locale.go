// Package locale holds the languages that predefined content is written in and that requests may ask for.
package locale

import (
	"fmt"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

// ErrInvalidLocale is returned for a locale that is not one of the known ones.
var ErrInvalidLocale = domain.NewError(domain.ErrorInvalid, "invalid locale")

// Locale is a language that predefined content is available in.
type Locale string

// The known locales.
const (
	LocaleEN Locale = "en"
	LocaleDE Locale = "de"
)

// DefaultLocale is used when none is requested and when a text has no translation for the requested one.
const DefaultLocale = LocaleEN

// Locales lists all known locales.
func Locales() []Locale { return []Locale{LocaleEN, LocaleDE} }

// ParseLocale returns the locale named by s. An empty s means DefaultLocale.
func ParseLocale(s string) (Locale, error) {
	if s == "" {
		return DefaultLocale, nil
	}

	for _, l := range Locales() {
		if Locale(s) == l {
			return l, nil
		}
	}

	return "", fmt.Errorf("%w: %q", ErrInvalidLocale, s)
}

// Localized is a text in several locales.
type Localized map[Locale]string

// For returns the text in locale l, or in DefaultLocale if there is none.
func (t Localized) For(l Locale) string {
	if s, ok := t[l]; ok {
		return s
	}

	return t[DefaultLocale]
}

// Comparer returns a function that compares two texts the way a reader of locale l sorts them: letters by their
// alphabet order, case and accents only breaking ties, so "alpha" < "Ärger" < "beta" < "Zeta". The function is not
// safe for concurrent use; get one per sort.
func (l Locale) Comparer() func(a, b string) int {
	return collate.New(language.Make(string(l))).CompareString
}
