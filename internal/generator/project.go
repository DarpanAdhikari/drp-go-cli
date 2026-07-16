package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// Auth adds JWT authentication, user management routes, and token storage.
	Auth bool
	// Infra selects infrastructure files to generate: "all", or comma-separated
	// like "docker,ci,make,lint". Empty means none.
	Infra string
	// DBDriver selects the database driver for generated code: "postgres" or "mysql".
	DBDriver string
}

// hasInfra reports whether opts.Infra includes the given component.
func hasInfra(opts ProjectOptions, name string) bool {
	infra := opts.Infra
	if infra == "" {
		return false
	}
	if infra == "all" {
		return true
	}
	for _, s := range strings.Split(infra, ",") {
		if strings.TrimSpace(s) == name {
			return true
		}
	}
	return false
}

// dbImport returns the Go import path for the database driver.
func dbImport(driver string) string {
	switch driver {
	case "mysql":
		return "github.com/go-sql-driver/mysql"
	default:
		return "github.com/lib/pq"
	}
}

// dbOpen returns the driver name for sql.Open.
func dbOpen(driver string) string {
	switch driver {
	case "mysql":
		return "mysql"
	default:
		return "postgres"
	}
}

// dsnTemplate returns a go-dsl format for the database DSN.
// base/main.go uses fmt.Sprintf with this as the format string.
func dsnTemplate(driver string) string {
	switch driver {
	case "mysql":
		return "%s:%s@tcp(%s:%s)/%s?parseTime=true"
	default:
		return "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"
	}
}

func dsnArgs(driver string) string {
	switch driver {
	case "mysql":
		return `getenv("DB_USER", "root"), getenv("DB_PASSWORD", ""), getenv("DB_HOST", "127.0.0.1"), getenv("DB_PORT", "3306"), getenv("DB_NAME", "` + "%s" + `")`
	default:
		return `getenv("DB_HOST", "127.0.0.1"), getenv("DB_PORT", "5432"), getenv("DB_USER", "postgres"), getenv("DB_PASSWORD", ""), getenv("DB_NAME", "` + "%s" + `"), getenv("DB_SSLMODE", "disable")`
	}
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
	if opts.DBDriver == "" {
		opts.DBDriver = "postgres"
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
		filepath.Join(root, "internal", "auth"),
		filepath.Join(root, "internal", "config"),
		filepath.Join(root, "internal", "middleware"),
		filepath.Join(root, "internal", "routes"),
		filepath.Join(root, "internal", "shared"),
		filepath.Join(root, "internal", "user"),
		filepath.Join(root, "tests"),
		filepath.Join(root, "docs"),
		filepath.Join(root, "database", "migrations"),
		filepath.Join(root, "database", "seeders"),
		filepath.Join(root, "pkg"),
		filepath.Join(root, "templates"),
	}

	// Add infra directories if requested.
	if hasInfra(opts, "docker") {
		dirs = append(dirs, filepath.Join(root, "docker"))
	}
	if hasInfra(opts, "ci") {
		dirs = append(dirs, filepath.Join(root, ".github", "workflows"))
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
	drv := opts.DBDriver
	files := map[string]string{
		// go.mod
		"go.mod": fmt.Sprintf("module %s\n\ngo %s\n", mod, getGoVersion()),

		// .env
		".env": envContents(opts),
		// .env.example (safe to commit)
		".env.example": envExampleContents(opts),

		// .gitignore
		".gitignore": ".env\n*.exe\n*.exe~\n*.test\ndist/\n",

		// cmd/api/main.go — runnable entrypoint
		filepath.Join("cmd", "api", "main.go"): mainGo(drv, opts.Name),

		// internal/config/config.go — local copy, not importing drp
		filepath.Join("internal", "config", "config.go"): configGo(drv, opts.Name),

		// internal/routes/health.go — always generated
		filepath.Join("internal", "routes", "health.go"): routesHealthGo(),

		// README
		"README.md": fmt.Sprintf("# %s\n\nGenerated by [DRP](https://github.com/DarpanAdhikari/drp-go-cli).\n\n## Getting started\n\n```bash\ncp .env.example .env\n# edit .env\ndrp migrate:seed\ndrp run api\n```\n", opts.Name),
	}

	// Infrastructure files (optional).
	if hasInfra(opts, "docker") {
		files["docker/Dockerfile"] = dockerfileGo()
		files[".dockerignore"] = dockerignoreGo()
		files["docker-compose.yml"] = dockerComposeGo(opts.Name)
	}
	if hasInfra(opts, "make") {
		files["Makefile"] = makefileGo()
	}
	if hasInfra(opts, "lint") {
		files[".editorconfig"] = editorconfigGo()
		files[".golangci.yml"] = golangciYmlGo()
	}
	if hasInfra(opts, "ci") {
		files[filepath.Join(".github", "workflows", "ci.yml")] = ciWorkflowGo()
	}

	if opts.Auth {
		for path, content := range authProjectFiles(opts) {
			files[path] = content
		}
	}

	return files
}

func envContents(opts ProjectOptions) string {
	drv := opts.DBDriver
	defaultPort := "5432"
	defaultUser := "postgres"
	if drv == "mysql" {
		defaultPort = "3306"
		defaultUser = "root"
	}
	return strings.Join([]string{
		"# Database configuration",
		fmt.Sprintf("DB_DRIVER=%s", drv),
		"DB_HOST=127.0.0.1",
		fmt.Sprintf("DB_PORT=%s", defaultPort),
		fmt.Sprintf("DB_USER=%s", defaultUser),
		"DB_PASSWORD=secret",
		fmt.Sprintf("DB_NAME=%s", opts.Name),
		"DB_SSLMODE=disable",
		"",
		"# Application",
		"APP_PORT=8080",
	}, "\n") + "\n"
}

func envExampleContents(opts ProjectOptions) string {
	drv := opts.DBDriver
	defaultPort := "5432"
	defaultUser := "postgres"
	if drv == "mysql" {
		defaultPort = "3306"
		defaultUser = "root"
	}
	return strings.Join([]string{
		fmt.Sprintf("DB_DRIVER=%s", drv),
		"DB_HOST=127.0.0.1",
		fmt.Sprintf("DB_PORT=%s", defaultPort),
		fmt.Sprintf("DB_USER=%s", defaultUser),
		"DB_PASSWORD=",
		fmt.Sprintf("DB_NAME=%s", opts.Name),
		"DB_SSLMODE=disable",
		"APP_PORT=8080",
	}, "\n") + "\n"
}

func mainGo(driver, name string) string {
	imp := dbImport(driver)
	open := dbOpen(driver)
	tpl := dsnTemplate(driver)
	args := dsnArgs(driver)

	return fmt.Sprintf(`package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "%s"
	"github.com/joho/godotenv"
	"%s/internal/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found: %%v", err)
	}

	dsn := fmt.Sprintf("%s",
		%s,
	)

	db, err := sql.Open("%s", dsn)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}
	log.Println("connected to database")

	mux := http.NewServeMux()
	routes.RegisterHealthRoute(mux, db)
	// TODO: register other routes here, e.g.:
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
`, imp, name, tpl, args, open)
}

func configGo(driver, name string) string {
	dsnBody := ""
	if driver == "mysql" {
		dsnBody = `// DSN returns a MySQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%%s@tcp(%%s:%%s)/%%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}`
	} else {
		dsnBody = `// DSN returns a PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%%s port=%%s user=%%s password=%%s dbname=%%s sslmode=%%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}`
	}

	return fmt.Sprintf(`// Package config loads .env and exposes typed config values.
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
		DBDriver:   getenv("DB_DRIVER", "%s"),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "%s"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", "%s"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		AppPort:    getenv("APP_PORT", "8080"),
	}
}

%s

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
`, driver, defaultPort(driver), name, dsnBody)
}

func defaultPort(driver string) string {
	if driver == "mysql" {
		return "3306"
	}
	return "5432"
}

// ── Health check ──────────────────────────────────────────────────────────

func routesHealthGo() string {
	return `package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// RegisterHealthRoute adds a GET /healthz endpoint that reports database status.
func RegisterHealthRoute(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
}
`
}

// ── Docker ────────────────────────────────────────────────────────────────

func dockerfileGo() string {
	return `# Multi-stage build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
`
}

func dockerignoreGo() string {
	return `.env
.git
dist/
tmp/
*.md
`
}

func dockerComposeGo(name string) string {
	return fmt.Sprintf(`version: "3.9"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file: .env
    depends_on:
      db:
        condition: service_healthy
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: %s
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      retries: 5
volumes:
  pgdata:
`, name)
}

// ── Makefile ──────────────────────────────────────────────────────────────

func makefileGo() string {
	return `.PHONY: test build run migrate docs clean

test:
	go mod tidy
	go test ./...

build:
	go build -o dist/api ./cmd/api

run:
	go run ./cmd/api

migrate:
	drp migrate:seed

docs:
	drp docs:generate

clean:
	rm -rf dist/
`
}

// ── EditorConfig ──────────────────────────────────────────────────────────

func editorconfigGo() string {
	return `root = true

[*]
indent_style = tab
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.md]
trim_trailing_whitespace = false
`
}

// ── golangci-lint ─────────────────────────────────────────────────────────

func golangciYmlGo() string {
	return `linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - unused

run:
  timeout: 5m
`
}

// ── GitHub Actions CI ─────────────────────────────────────────────────────

func ciWorkflowGo() string {
	return `name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go mod tidy
      - run: go build ./...
      - run: go vet ./...
      - run: go test ./...
`
}

// ── Helpers ───────────────────────────────────────────────────────────────

func getGoVersion() string {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return "1.22"
}
