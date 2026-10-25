package backend

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

const (
	// DefaultHTTPAddr is used when HTTP_ADDR is not set.
	DefaultHTTPAddr = ":8080"
	// DefaultMaxConns is the pool size used when DB_MAX_CONNS is not set.
	DefaultMaxConns = 10
	// DefaultMaxConnLifetime is used when DB_MAX_CONN_LIFETIME is not set.
	DefaultMaxConnLifetime = 30 * time.Minute
	// DefaultHealthCheckPeriod is used when DB_HEALTH_CHECK_PERIOD is not set.
	DefaultHealthCheckPeriod = time.Minute
)

// Config is the runtime configuration of the backend.
type Config struct {
	// DatabaseURL must be a postgres:// URL.
	DatabaseURL string
	// HTTPAddr defaults to DefaultHTTPAddr.
	HTTPAddr string
	// LogLevel defaults to info.
	LogLevel slog.Level
	// SkipMigrate disables the migrations on startup, for when a separate job (`docker migrate`) runs them.
	SkipMigrate bool
	// MaxConns, MaxConnLifetime and HealthCheckPeriod tune the connection pool; zero keeps the pgx default.
	MaxConns          int32
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration

	envErr error // unparsable environment values, reported by Validate
}

// ConfigFromEnv reads the environment. It only applies defaults; call Validate to check the result.
// Unparsable values are reported by Validate, so ConfigFromEnv itself never fails.
func ConfigFromEnv(getenv func(string) string) Config {
	cfg := Config{
		DatabaseURL:       getenv("DATABASE_URL"),
		HTTPAddr:          getenv("HTTP_ADDR"),
		MaxConns:          DefaultMaxConns,
		MaxConnLifetime:   DefaultMaxConnLifetime,
		HealthCheckPeriod: DefaultHealthCheckPeriod,
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
	}

	var errs []error

	if v := getenv("LOG_LEVEL"); v != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
		}
	}

	if v := getenv("MIGRATE_ON_START"); v != "" {
		on, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("MIGRATE_ON_START: %w", err))
		}

		cfg.SkipMigrate = !on
	}

	if v := getenv("DB_MAX_CONNS"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 1 {
			errs = append(errs, errors.New("DB_MAX_CONNS: must be a positive integer"))
		} else {
			cfg.MaxConns = int32(n)
		}
	}

	cfg.MaxConnLifetime = envDuration(getenv, "DB_MAX_CONN_LIFETIME", cfg.MaxConnLifetime, &errs)
	cfg.HealthCheckPeriod = envDuration(getenv, "DB_HEALTH_CHECK_PERIOD", cfg.HealthCheckPeriod, &errs)
	cfg.envErr = errors.Join(errs...)

	return cfg
}

func envDuration(getenv func(string) string, key string, def time.Duration, errs *[]error) time.Duration {
	v := getenv(key)
	if v == "" {
		return def
	}

	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		*errs = append(*errs, fmt.Errorf("%s: must be a positive duration like 30m", key))
		return def
	}

	return d
}

// Validate reports whether the config is complete.
func (c Config) Validate() error {
	if c.envErr != nil {
		return c.envErr
	}

	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.HTTPAddr == "" {
		return errors.New("HTTP_ADDR is required")
	}

	return nil
}
