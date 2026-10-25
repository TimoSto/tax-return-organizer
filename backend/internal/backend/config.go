package backend

import (
	"errors"
	"time"
)

// DefaultHTTPAddr is used when HTTP_ADDR is not set.
const DefaultHTTPAddr = ":8080"

// DefaultDrainDelay is used when DRAIN_DELAY is not set.
const DefaultDrainDelay = 5 * time.Second

// Config is the runtime configuration of the backend.
type Config struct {
	// DatabaseURL must be a postgres:// URL.
	DatabaseURL string
	// HTTPAddr defaults to DefaultHTTPAddr.
	HTTPAddr string
	// DrainDelay is how long the server reports itself not ready before it stops accepting connections on
	// shutdown, so the orchestrator stops routing to it first. Zero skips the wait.
	DrainDelay time.Duration
}

// ConfigFromEnv reads DATABASE_URL, HTTP_ADDR and DRAIN_DELAY (a Go duration such as "5s"). It only applies defaults; call Validate to check the result.
func ConfigFromEnv(getenv func(string) string) Config {
	cfg := Config{
		DatabaseURL: getenv("DATABASE_URL"),
		HTTPAddr:    getenv("HTTP_ADDR"),
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
	}

	cfg.DrainDelay = DefaultDrainDelay

	if v := getenv("DRAIN_DELAY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			d = -1 // rejected by Validate
		}

		cfg.DrainDelay = d
	}

	return cfg
}

// Validate reports whether the config is complete.
func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.HTTPAddr == "" {
		return errors.New("HTTP_ADDR is required")
	}

	if c.DrainDelay < 0 {
		return errors.New("DRAIN_DELAY must be a non-negative duration such as 5s")
	}

	return nil
}
