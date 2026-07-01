package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectOptions configures `drp init`.
type ProjectOptions struct {
	// Name is the project directory name, e.g. "myapp".
	Name string
	// ModuleName is the Go module path, e.g. "github.com/yourorg/myapp".
	// Defaults to Name if empty.
	ModuleName string
	// Force allows writing into a non-empty directory without refusing.
	Force bool
}

// Project scaffolds a new DRP-backed Go project at opts.Name (relative to
// the current working directory). It refuses if the directory already
// exists and is non-empty, unless opts.Force is true.
func Project(opts ProjectOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("project: name cannot be empty")
	}
	if opts.ModuleName == "" {
		opts.ModuleName = opts.Name
	}

	root := filepath.Join(".", opts.Name)

	// Guard: refuse non-empty directory unless --force.
	if !opts.Force {
		if entries, err := os.ReadDir(root); err == nil && len(entries) > 0 {
			return fmt.Errorf(
				"project: directory %q already exists and is not empty\n"+
					"  → use --force to overwrite, or choose a different name", root,
			)
		}
	}

	dirs := []string{
		filepath.Join(root, "cmd", "api"),
		filepath.Join(root, "internal", "config"),
		filepath.Join(root, "internal", "handlers"),
		filepath.Join(root, "internal", "repositories"),
		filepath.Join(root, "internal", "services"),
		filepath.Join(root, "internal", "routes"),
		filepath.Join(root, "internal", "models"),
		filepath.Join(root, "database", "migrations"),
		filepath.Join(root, "database", "seeders"),
		filepath.Join(root, "pkg"),
		filepath.Join(root, "templates"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("project: creating directory %q: %w", d, err)
		}
	}

	files := projectFiles(opts)
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("project: creating parent of %q: %w", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("project: writing %q: %w", fullPath, err)
		}
	}

	return nil
}

// projectFiles returns a map of relative-path → content for all files that
// `drp init` writes into the new project root.
func projectFiles(opts ProjectOptions) map[string]string {
	mod := opts.ModuleName

	return map[string]string{
		// go.mod
		"go.mod": fmt.Sprintf("module %s\n\ngo 1.22\n", mod),

		// .env
		".env": strings.Join([]string{
			"# Database configuration",
			"DB_DRIVER=postgres",
			"DB_HOST=127.0.0.1",
			"DB_PORT=5432",
			"DB_USER=postgres",
			"DB_PASSWORD=secret",
			fmt.Sprintf("DB_NAME=%s", opts.Name),
			"DB_SSLMODE=disable",
			"",
			"# Application",
			"APP_PORT=8080",
		}, "\n") + "\n",

		// .env.example (safe to commit)
		".env.example": strings.Join([]string{
			"DB_DRIVER=postgres",
			"DB_HOST=127.0.0.1",
			"DB_PORT=5432",
			"DB_USER=",
			"DB_PASSWORD=",
			fmt.Sprintf("DB_NAME=%s", opts.Name),
			"DB_SSLMODE=disable",
			"APP_PORT=8080",
		}, "\n") + "\n",

		// .gitignore
		".gitignore": ".env\n*.exe\n*.exe~\n*.test\ndist/\n",

		// cmd/api/main.go — runnable entrypoint
		filepath.Join("cmd", "api", "main.go"): fmt.Sprintf(`package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found: %%v", err)
	}

	dsn := fmt.Sprintf(
		"host=%%s port=%%s user=%%s password=%%s dbname=%%s sslmode=%%s",
		getenv("DB_HOST", "127.0.0.1"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "postgres"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_NAME", "%s"),
		getenv("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}
	log.Println("connected to database")

	mux := http.NewServeMux()
	// TODO: register routes here, e.g.:
	//   routes.RegisterProductRoutes(mux, db)

	port := getenv("APP_PORT", "8080")
	log.Printf("listening on :%%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
`, opts.Name),

		// internal/config/config.go — local copy, not importing drp
		filepath.Join("internal", "config", "config.go"): fmt.Sprintf(`// Package config loads .env and exposes typed config values.
// This file is a standalone copy — it does not import drp at runtime.
package config

import (
	"fmt"
	"os"
)

// Config holds application configuration loaded from .env.
type Config struct {
	DBDriver   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	AppPort    string
}

// Load reads environment variables (call godotenv.Load() before this).
func Load() *Config {
	return &Config{
		DBDriver:   getenv("DB_DRIVER", "postgres"),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", "%s"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		AppPort:    getenv("APP_PORT", "8080"),
	}
}

// PostgresDSN returns a libpq-style connection string.
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%%s port=%%s user=%%s password=%%s dbname=%%s sslmode=%%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
`, opts.Name),

		// README
		"README.md": fmt.Sprintf("# %s\n\nGenerated by [DRP](https://github.com/DarpanAdhikari/drp-go-cli).\n\n## Getting started\n\n```bash\ncp .env.example .env\n# edit .env\ndrp migrate:seed\ngo run ./cmd/api\n```\n", opts.Name),
	}
}
