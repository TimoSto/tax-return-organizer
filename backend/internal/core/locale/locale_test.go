package locale_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/locale"
)

func TestParseLocale(t *testing.T) {
	for in, want := range map[string]locale.Locale{"": locale.LocaleEN, "en": locale.LocaleEN, "de": locale.LocaleDE} {
		if got, err := locale.ParseLocale(in); err != nil || got != want {
			t.Errorf("ParseLocale(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	if _, err := locale.ParseLocale("fr"); err == nil {
		t.Error("expected an error for fr")
	}
}

func TestLocalized_FallsBackToDefault(t *testing.T) {
	if got := (locale.Localized{locale.LocaleEN: "x"}).For(locale.LocaleDE); got != "x" {
		t.Errorf("got %q, want x", got)
	}
}
