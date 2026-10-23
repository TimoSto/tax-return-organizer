package domain_test

import (
	"errors"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestNewPersona(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pName       string
		description string
		expName     string
		expDesc     string
		expErr      error
	}{
		{name: "Valid", pName: "Employed", description: "Has a job", expName: "Employed", expDesc: "Has a job"},
		{name: "Trims input", pName: "  Letting ", description: " rents out ", expName: "Letting", expDesc: "rents out"},
		{name: "Empty description is allowed", pName: "ETF investing", expName: "ETF investing"},
		{name: "Empty name", pName: "", expErr: domain.ErrPersonaNameRequired},
		{name: "Blank name", pName: "  \t", expErr: domain.ErrPersonaNameRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := domain.NewPersona("aaaa", tc.pName, tc.description)

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("err = %v, want %v", err, tc.expErr)
			}

			if err != nil {
				return
			}

			if p.Name != tc.expName || p.Description != tc.expDesc {
				t.Errorf("persona = %q/%q, want %q/%q", p.Name, p.Description, tc.expName, tc.expDesc)
			}

			if p.CollectorID != "aaaa" {
				t.Errorf("collector id = %q, want aaaa", p.CollectorID)
			}
		})
	}
}

func TestNewPersona_GeneratesUniqueIDs(t *testing.T) {
	a, _ := domain.NewPersona("aaaa", "A", "")
	b, _ := domain.NewPersona("aaaa", "A", "")

	if a.ID == b.ID {
		t.Error("two personas share an id")
	}
}
