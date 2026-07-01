// Package config loads and validates the project's .env configuration.
// It fails fast and loud on missing or malformed values, pointing the
// user at `drp doctor` rather than letting a cryptic DB error surface later.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// SupportedDrivers lists the DB drivers drp understands.
var SupportedDrivers = []string{"postgres", "mysql"}

// Config holds the resolved, validated configuration for a drp project.
type Config struct {
	DBDriver   string // "postgres" or "mysql"
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string // postgres only; ignored for mysql
}

// Load reads a .env file from envFile (defaults to ".env" in the working
// directory when empty) and returns a validated Config.
// Returns a descriptive error pointing at `drp doctor` if anything is wrong.
func Load(envFile string) (*Config, error) {
	if envFile == "" {
		envFile = ".env"
	}

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"config: %q not found — create one or run `drp doctor` to diagnose",
			envFile,
		)
	}

	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf(
			"config: failed to parse %q: %w\n  → run `drp doctor` for a full check",
			envFile, err,
		)
	}

	cfg := &Config{
		DBDriver:   strings.ToLower(getenv("DB_DRIVER", "postgres")),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", defaultPort(strings.ToLower(getenv("DB_DRIVER", "postgres")))),
		DBUser:     getenv("DB_USER", ""),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", ""),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
	}

	return cfg, cfg.validate()
}

// validate checks that all required fields are present and the driver is known.
func (c *Config) validate() error {
	var missing []string

	if c.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if c.DBName == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		return fmt.Errorf(
			"config: missing required .env variable(s): %s\n  → run `drp doctor` for a full check",
			strings.Join(missing, ", "),
		)
	}

	supported := false
	for _, d := range SupportedDrivers {
		if c.DBDriver == d {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf(
			"config: unknown DB_DRIVER %q — supported drivers: %s",
			c.DBDriver, strings.Join(SupportedDrivers, ", "),
		)
	}

	return nil
}

// DSN returns the driver-appropriate data source name.
func (c *Config) DSN() string {
	switch c.DBDriver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
		)
	case "mysql":
		// user:password@tcp(host:port)/dbname?parseTime=true&multiStatements=true
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
		)
	default:
		return ""
	}
}

// AdminDSN returns a DSN to the server without a specific database selected,
// used when creating the target database if it doesn't exist.
func (c *Config) AdminDSN() string {
	switch c.DBDriver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBSSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/?parseTime=true&multiStatements=true",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort,
		)
	default:
		return ""
	}
}

// getenv returns the environment variable named by key, or fallback if unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// defaultPort returns the conventional default port for the given driver name.
func defaultPort(driver string) string {
	switch driver {
	case "mysql":
		return "3306"
	default:
		return "5432"
	}
}
