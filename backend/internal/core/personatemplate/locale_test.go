package personatemplate_test

import (
	"testing"

	pt "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"
)

func TestParseLocale(t *testing.T) {
	for in, want := range map[string]pt.Locale{"": pt.LocaleEN, "en": pt.LocaleEN, "de": pt.LocaleDE} {
		if got, err := pt.ParseLocale(in); err != nil || got != want {
			t.Errorf("ParseLocale(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	if _, err := pt.ParseLocale("fr"); err == nil {
		t.Error("expected an error for fr")
	}
}

func TestLocalized_FallsBackToDefault(t *testing.T) {
	if got := (pt.Localized{pt.LocaleEN: "x"}).For(pt.LocaleDE); got != "x" {
		t.Errorf("got %q, want x", got)
	}
}
