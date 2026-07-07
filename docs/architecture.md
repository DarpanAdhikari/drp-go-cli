# Architecture

This document explains the internal structure of the `drp` repository —
how packages relate to each other, how the migration and seeder engines work,
and how templates are resolved.

---

## Repository Layout

```
drp/
├── main.go                  ← entry point — calls cmd.Execute()
│
├── cmd/                     ← Cobra CLI commands (thin layer)
│   ├── root.go              ← root command, global flags
│   ├── init.go              ← drp init
│   ├── doctor.go            ← drp doctor
│   ├── migrate.go           ← drp migrate, migrate:up/down/create/rollback/fresh/status/seed
│   ├── rollback.go          ← drp rollback (restore .old binary)
│   ├── seeder.go            ← drp seeder:create, seeder:status, db:seed
│   ├── db.go                ← drp db:status/tables/drop/reset
│   ├── create.go            ← drp create:crud
│   ├── run.go               ← drp run [target] (with --watch)
│   ├── run_unix.go          ← unix process management (linux + darwin)
│   ├── run_windows.go       ← windows process management
│   ├── build.go             ← drp build [target]
│   ├── test.go              ← drp test
│   ├── docs.go              ← drp docs:generate
│   ├── completion.go        ← drp completion [bash|zsh|fish|powershell]
│   ├── install.go           ← drp install
│   ├── upgrade.go           ← drp upgrade
│   ├── version.go           ← drp version, version:check
│   └── drp/
│       └── main.go          ← program entry point
│
├── internal/                ← all business logic (unexported to the outside world)
│   ├── config/              ← .env loading and validation
│   ├── db/                  ← database connection, ping, table helpers
│   ├── doctor/              ← environment health checks
│   ├── generator/           ← project scaffolding and CRUD code generation
│   ├── migration/           ← migration engine (up/down/rollback/fresh/status)
│   ├── output/              ← coloured terminal output helpers
│   └── seeder/              ← seeder engine (run/fresh)
│
├── templates/               ← Go text/template files embedded into the binary
│   ├── model.tpl
│   ├── repository.tpl
│   ├── service.tpl
│   ├── handler.tpl
│   ├── routes.tpl
│   ├── seeder.tpl
│   ├── migration_up.tpl
│   └── migration_down.tpl
│
├── pkg/                     ← optional shared utilities (currently empty)
├── examples/                ← example projects showing DRP in use
├── scripts/
│   └── build-release.sh     ← cross-compile release binaries (Linux/macOS/Windows)
├── .github/workflows/
│   ├── ci.yml               ← test + lint on every push
│   └── release.yml          ← publish binaries on git tag
├── docs/                    ← this documentation
├── go.mod
├── go.sum
└── CHANGELOG.md
```

---

## Design Principles

1. **Commands stay thin.** Everything inside `cmd/` only parses flags, calls
   `internal/` packages, and formats output via `internal/output`. Business
   logic never lives in `cmd/`.

2. **No runtime dependency on drp.** Generated code copies helper files
   (like `internal/config/config.go`) directly into the new project. Your
   application never has `import "github.com/yourorg/drp/..."` in production.

3. **No ORM, no migration library.** The migration and seeder engines run
   plain SQL files against `database/sql`. Every generated file is readable,
   editable, and independently understandable Go code.

4. **No global state.** All engines and helpers accept their dependencies
   (config, DB connection, directory paths) as arguments.

5. **Exported functions are documented.** Every exported type, function, and
   method has a Go doc comment.

---

## Package Responsibilities

### `internal/config`

Loads a `.env` file using `godotenv`, validates required fields (`DB_USER`,
`DB_NAME`), and validates that `DB_DRIVER` is one of the supported values
(`postgres`, `mysql`). Exposes `Config.DSN()` and `Config.AdminDSN()` for
the two drivers.

### `internal/db`

Opens and pings a `*sql.DB` connection using the driver chosen in the config.
Provides `TableNames()` (lists all user tables in a driver-agnostic way) and
`DropAllTables()` (drops everything in dependency order using
`FOREIGN_KEY_CHECKS=0` on MySQL and a `CASCADE` approach on PostgreSQL).

### `internal/doctor`

Runs a battery of environment checks and returns a `[]Result` slice — each
with a `Label`, `Detail`, and boolean `OK`. The `doctor` command iterates the
results and prints them without stopping on the first failure.

### `internal/generator`

Contains two sub-responsibilities:

- **Project scaffolding** (`project.go`): creates the full directory tree and
  starter files (`.env`, `go.mod`, `cmd/api/main.go`, etc.) for a new project.
- **CRUD generation** (`crud.go`, `render.go`, `names.go`): resolves naming
  conventions (singular/plural, PascalCase, snake_case), selects the correct
  template for each layer, renders it, and writes the files.

The `names.go` file is responsible for all name derivation — given the raw
resource name (e.g. `"product"`), it produces `Name`, `NamePlural`,
`TableName`, `PackageName`, `ModuleName`, etc. used in templates.

### `internal/migration`

Implements a simple, SQL-file-based migration engine:

- **File naming convention:** `<timestamp>_<name>.up.sql` / `<timestamp>_<name>.down.sql`
- **Tracking table:** `drp_migrations` (created automatically on first run),
  which stores `id`, `name`, and `batch` (an integer grouping migrations
  applied together in one `migrate:up` call).
- **`Engine.Up()`** — finds all `.up.sql` files not yet recorded in
  `drp_migrations`, runs them in timestamp order, records them in the
  same batch number.
- **`Engine.Down()`** — runs the `.down.sql` of the single most-recently
  applied migration, removes its row from `drp_migrations`.
- **`Engine.Rollback()`** — runs `.down.sql` for every migration in the last
  batch (same batch number), in reverse order.
- **`Engine.Fresh()`** — drops all tables (via `internal/db`), then calls
  `Up()`.
- **`Engine.Status()`** — returns a list of all migration files alongside
  their applied/pending state.

### `internal/seeder`

Implements a seeder engine that runs plain `.sql` files from `database/seeders/`:

- Tracks which seeders have already run in a `drp_seeders` table.
- **`Engine.Run(fresh bool)`** — if `fresh`, clears the tracking table first;
  then runs any un-seeded files in alphabetical order.

### `internal/output`

A small package providing coloured terminal output helpers:
`Success` (green ✔), `Fail` (red ✗), `Info` (cyan →), `Warn` (yellow ⚠).
Respects the `--no-color` global flag via `SetNoColor(bool)`.

---

## Template Resolution — Embedded vs. User Templates

When `drp create:crud` generates files, it follows a **precedence rule**:

1. **User-supplied template** — looks for a `.tpl` file in the project's
   `templates/` directory (e.g. `templates/handler.tpl`).
2. **Embedded template** — falls back to the template baked into the `drp`
   binary at compile time from the repository's `templates/` directory.

This means you can override any template for a specific project without
modifying `drp` itself.

Templates are Go `text/template` files. Available variables (set by `names.go`):

| Variable | Example (input: `"product"`) |
|---|---|
| `{{.Name}}` | `Product` |
| `{{.NameLower}}` | `product` |
| `{{.NamePlural}}` | `Products` |
| `{{.TableName}}` | `products` |
| `{{.PackageName}}` | `product` |
| `{{.ModuleName}}` | `github.com/acme/myapp` |

---

## Migration File Convention

```
database/migrations/
  20240101120000_create_users_table.up.sql
  20240101120000_create_users_table.down.sql
  20240102090000_create_products_table.up.sql
  20240102090000_create_products_table.down.sql
```

- The timestamp prefix (`YYYYMMDDHHmmss`) determines execution order.
- Each `.up.sql` must have a matching `.down.sql`.
- The migration engine runs each file as a single transaction where the driver supports it.

---

## Seeder File Convention

```
database/seeders/
  001_users.sql
  002_products.sql
```

- Seeders are run in **alphabetical order**.
- Use numeric prefixes (`001_`, `002_`) to control ordering.
- The seeder engine tracks which files have run in the `drp_seeders` table.

---

## Configuration Loading

Config is loaded from `.env` (or the path given by `--env-file`) using
`github.com/joho/godotenv`. The relevant keys are:

```dotenv
DB_DRIVER=postgres      # Required: "postgres" or "mysql"
DB_HOST=127.0.0.1       # Default: 127.0.0.1
DB_PORT=5432            # Default: 5432 (postgres) or 3306 (mysql)
DB_USER=                # Required
DB_PASSWORD=            # Default: (empty)
DB_NAME=                # Required
DB_SSLMODE=disable      # Postgres only; default: disable
APP_PORT=8080           # Used by the generated cmd/api/main.go
```

---

## CI / Release Workflow

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | Push / PR to any branch | Runs `go test ./...`, `go vet`, `staticcheck` |
| `release.yml` | Push of a `v*.*.*` tag | Cross-compiles binaries for Linux/macOS/Windows (amd64 + arm64) and uploads to a GitHub Release |

The cross-compile script is `scripts/build-release.sh`. Release binaries are
output to `dist/` with the `drp` version embedded via `-ldflags`.
