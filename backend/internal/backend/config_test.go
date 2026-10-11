package backend_test

import (
	"context"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
)

func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestConfigFromEnv(t *testing.T) {
	t.Run("defaults http addr", func(t *testing.T) {
		cfg := backend.ConfigFromEnv(envFrom(map[string]string{"DATABASE_URL": "postgres://x"}))

		if cfg.HTTPAddr != backend.DefaultHTTPAddr {
			t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, backend.DefaultHTTPAddr)
		}
	})

	t.Run("uses provided values", func(t *testing.T) {
		cfg := backend.ConfigFromEnv(envFrom(map[string]string{"DATABASE_URL": "postgres://x", "HTTP_ADDR": ":9000"}))

		if cfg.DatabaseURL != "postgres://x" || cfg.HTTPAddr != ":9000" {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})
}

func TestRun_RequiresDatabaseURL(t *testing.T) {
	if err := backend.Run(context.Background(), backend.ConfigFromEnv(envFrom(nil))); err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     backend.Config
		wantErr bool
	}{
		{"complete", backend.Config{DatabaseURL: "postgres://x", HTTPAddr: ":8080"}, false},
		{"missing database url", backend.Config{HTTPAddr: ":8080"}, true},
		{"missing http addr", backend.Config{DatabaseURL: "postgres://x"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
