package backend

const DefaultHTTPAddr = ":8080"

type Config struct {
	// DatabaseURL must be a postgres:// URL.
	DatabaseURL string
	// HTTPAddr defaults to DefaultHTTPAddr.
	HTTPAddr string
}

// ConfigFromEnv reads DATABASE_URL and HTTP_ADDR. It does not validate: Run rejects an empty database URL.
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
