package backend

import "errors"

// DefaultHTTPAddr is used when HTTP_ADDR is not set.
const DefaultHTTPAddr = ":8080"

// Config is the runtime configuration of the backend.
type Config struct {
	// DatabaseURL must be a postgres:// URL.
	DatabaseURL string
	// HTTPAddr defaults to DefaultHTTPAddr.
	HTTPAddr string
}

// ConfigFromEnv reads DATABASE_URL and HTTP_ADDR. It only applies defaults; call Validate to check the result.
func ConfigFromEnv(getenv func(string) string) Config {
	cfg := Config{
		DatabaseURL: getenv("DATABASE_URL"),
		HTTPAddr:    getenv("HTTP_ADDR"),
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
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

	return nil
}
