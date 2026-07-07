# Commands Reference

Complete reference for every `drp` command, organised in the order you'd
typically use them during a project lifecycle.

**Global flags** (apply to all commands):

| Flag | Default | Description |
|---|---|---|
| `--env-file <path>` | `.env` | Path to the `.env` file to load |
| `--no-color` | `false` | Disable coloured terminal output |
| `-h, --help` | — | Show help for any command |

---

## 1. Setup & Self-Management

Commands for installing, upgrading, and getting information about `drp`
itself. Run these once to set up your toolchain.

### `drp install`

Install the current binary as `drp` (copies itself into `~/.local/bin/`
or a custom directory).

```
drp install [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--bin-dir <path>` | `~/.local/bin` | Directory to install drp into |
| `--no-shell` | `false` | Do not update shell profile PATH |

#### Examples

```bash
drp install
drp install --bin-dir /usr/local/bin
```

---

### `drp upgrade`

Download the latest (or a specific) `drp` release from GitHub and replace
the current binary.

```
drp upgrade [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--repo <owner/name>` | `DarpanAdhikari/drp-go-cli` | GitHub repository to fetch from |
| `--bin-dir <path>` | _(current dir or ~/.local/bin)_ | Directory containing drp |
| `--asset-url <url>` | — | Direct release asset URL (for testing) |
| `--version <tag>` | _(latest)_ | Specific version to install, e.g. `v0.1.0` |

`--version` and `--asset-url` cannot be used together.

#### Examples

```bash
drp upgrade                          # latest stable
drp upgrade --version v0.2.0         # specific version
```

---

### `drp rollback`

Restore the previous `drp` binary from the `.old` backup left behind by
`drp upgrade` or `drp install`.

```
drp rollback
```

No flags. Run if the latest upgrade introduced a problem.

---

### `drp version`

Print the current `drp` CLI version.

```
drp version
```

### `drp version:check`

Check whether a newer `drp` release is available on GitHub.

```
drp version:check
```

Compares the running version against the latest GitHub release tag and
suggests `drp upgrade` if a newer version exists.

---

### `drp completion`

Generate shell auto-completion scripts for bash, zsh, fish, or PowerShell.

```
drp completion [bash|zsh|fish|powershell]
```

#### Examples — enable completion

```bash
# Bash — source for current session
source <(drp completion bash)

# Bash — permanent
drp completion bash | sudo tee /etc/bash_completion.d/drp

# Zsh — permanent
drp completion zsh > "${fpath[1]}/_drp"
```

---

## 2. Environment & Diagnostics

### `drp doctor`

Check your environment for common DRP setup issues.

```
drp doctor
```

Checks performed (all run even if an earlier one fails):

- Go installation and version
- Presence and parseability of `.env`
- Required `.env` variables (`DB_USER`, `DB_NAME`)
- Database reachability (connect + ping)
- Expected project directory structure

#### Examples

```bash
drp doctor
drp doctor --env-file .env.test
```

---

## 3. Project Scaffolding

### `drp init`

Scaffold a new DRP-backed Go project.

```
drp init [project-name] [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--module <path>` | _(project name)_ | Go module path (e.g. `github.com/acme/myapp`) |
| `--auth` | `true` | Scaffold JWT auth, users, token handling, and auth routes |
| `--infra <components>` | `""` (none) | Generate infrastructure files: `all`, `docker`, `ci`, `make`, `lint` (comma-separated) |
| `--driver <driver>` | `postgres` | Database driver: `postgres` or `mysql` |
| `--force` | `false` | Overwrite a non-empty directory |
| `--no-interaction` | `false` | Skip interactive prompts |

> **`--auth` defaults to `true`.** Pass `--auth=false` to skip the auth preset.

> **`--infra`** accepts `"all"` for everything, or a comma-separated list
> like `"docker,ci"` for specific components.

> **`--driver`** controls DSN format, migration SQL syntax, and default
> ports (5432 for postgres, 3306 for mysql).

#### What it generates (base)

```
<project-name>/
├── cmd/api/main.go
├── internal/config/config.go
├── database/migrations/
├── database/seeders/
├── pkg/
├── templates/
├── .env
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

A `GET /healthz` endpoint is always generated (uses `database/sql.Ping()`
on every request).

#### With `--auth`

```
<project-name>/
├── cmd/api/main.go
├── internal/
│   ├── auth/                  ← JWT, token store, auth handlers
│   ├── config/config.go
│   ├── middleware/             ← CORS, rate limiting, request ID, auth
│   ├── routes/
│   ├── shared/                ← base types, context keys, response helpers
│   └── user/                  ← user model, repository, service, handler
├── database/migrations/
├── database/seeders/
├── docs/                      ← swagger documentation
├── pkg/
├── templates/
├── .env
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

#### With `--infra`

Depending on the components selected, additional files are generated:

| Component | Files |
|---|---|
| `docker` | `docker/Dockerfile`, `.dockerignore`, `docker-compose.yml` |
| `ci` | `.github/workflows/ci.yml` (Go CI workflow) |
| `make` | `Makefile` (build, test, lint, run targets) |
| `lint` | `.editorconfig`, `.golangci.yml` |

#### Examples

```bash
# Basic usage (interactive prompts)
drp init myapp

# Skip interactive prompts
drp init myapp --no-interaction

# With authentication preset
drp init myapp --auth

# Skip auth (bare project only)
drp init myapp --auth=false

# Custom Go module path
drp init myapp --module github.com/acme/myapp

# MySQL driver
drp init myapp --driver mysql

# All infrastructure files
drp init myapp --infra all

# Docker + CI only
drp init myapp --infra docker,ci

# Re-scaffold into an existing directory
drp init myapp --force

# Full non-interactive example
drp init myapp --no-interaction --auth --infra docker,ci --driver postgres
```

After creation, `drp` automatically runs `git init` in the new project
directory.

---

### `drp create:crud`

Generate all CRUD layers for a resource inside the current project directory.

```
drp create:crud [name] [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--model` | `-m` | `false` | Generate model only |
| `--repository` | `-r` | `false` | Generate repository only |
| `--service` | `-s` | `false` | Generate service only |
| `--handler` | — | `false` | Generate handler only |
| `--routes` | — | `false` | Generate routes only |
| `--module <path>` | — | _(auto-detected)_ | Override Go module name |
| `--no-interaction` | — | `false` | Skip interactive prompts (implies all layers) |

When no layer flags are given and `--no-interaction` is not set, `drp`
enters **interactive mode**:

1. Asks which layers to generate (checkboxes)
2. Shows a preview of files that will be created
3. Asks for confirmation before writing

With no layer flags in non-interactive mode, all five files are generated:

```
internal/<domain>/model.go
internal/<domain>/repository.go
internal/<domain>/service.go
internal/<domain>/handler.go
internal/routes/<domain>_routes.go
```

All four domain layers share the same package — no cross-package imports
between model, repository, service, and handler.

#### Examples

```bash
# Interactive mode (recommended)
drp create:crud product

# Generate all layers non-interactively
drp create:crud product --no-interaction

# Generate only the model and repository
drp create:crud product -m -r

# Specify module explicitly
drp create:crud product --module github.com/acme/myapp
```

After generation, register the routes in `cmd/api/main.go`:

```go
routes.RegisterProductRoutes(mux, db)
```

---

## 4. Database Commands

All database commands read `database/migrations/` and connect to the
database configured in `.env`.

### 4a. Migrations

#### `drp migrate` / `drp migrate:up`

Run all pending (not yet applied) migrations.

```
drp migrate
drp migrate:up
```

Both commands are equivalent. Prints an info message if there is nothing
to run.

---

#### `drp migrate:create`

Create a new timestamped migration file pair (up + down).

```
drp migrate:create [name] [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--table <name>` | _(inferred)_ | Table name to use in starter SQL |

Creates two files in `database/migrations/`:

```
database/migrations/
  20240101120000_<name>.up.sql
  20240101120000_<name>.down.sql
```

#### Example

```bash
drp migrate:create create_users_table
drp migrate:create add_email_to_orders --table orders
```

---

#### `drp migrate:down`

Roll back the single most-recently applied migration.

```
drp migrate:down [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be stepped down without executing |

---

#### `drp migrate:rollback`

Roll back the **entire last batch** of migrations (all that ran together
in the last `migrate:up` call).

```
drp migrate:rollback [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be rolled back without executing |

---

#### `drp migrate:fresh`

Drop **all tables**, then re-run all migrations from scratch.

```
drp migrate:fresh [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be dropped/applied without executing |

> ⚠️ **Destructive.** All data will be lost.

---

#### `drp migrate:status`

Show which migrations have been applied and which are still pending.

```
drp migrate:status
```

Output example:

```
  ✔  20240101120000_create_users_table
  ✔  20240102090000_create_products_table
  ✗  20240103150000_add_category_to_products   (pending)
```

---

#### `drp migrate:seed`

Run pending migrations, then immediately run the database seeders.
Shortcut for `migrate:up` followed by `db:seed`.

```
drp migrate:seed [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--fresh` | `false` | Drop all tables, re-run migrations, then re-run all seeders |

---

### 4b. Seeders

#### `drp seeder:create`

Create a new seeder SQL file in `database/seeders/`.

```
drp seeder:create [name]
```

#### Example

```bash
drp seeder:create users
drp seeder:create products
```

Creates: `database/seeders/<name>.sql`

---

#### `drp seeder:status`

Show which seeders have run and which are pending.

```
drp seeder:status
```

Output example:

```
  ✔  users
  ✗  products   (pending)
```

---

#### `drp db:seed`

Run the database seeders.

```
drp db:seed [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--fresh` | `false` | Drop all tables, re-run migrations, then re-run **all** seeders from scratch |

---

### 4c. Database Utilities

#### `drp db:status`

Show the current database connection status and basic info.

```
drp db:status
```

Output example:

```
✔  Connected to "myapp" on 127.0.0.1:5432 (postgres)
```

---

#### `drp db:tables`

List all tables currently in the configured database.

```
drp db:tables
```

---

#### `drp db:drop`

Drop **all tables** in the configured database.

```
drp db:drop [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | List tables that would be dropped without executing |

> ⚠️ **Destructive.** All data and schema will be lost.

---

#### `drp db:reset`

Drop all tables and re-run all migrations from scratch (alias for
`migrate:fresh`).

```
drp db:reset [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be dropped/applied without executing |

---

## 5. Development Workflow

### `drp run`

Run a Go command from `./cmd/<target>`.

```
drp run [target] [args...] [flags]
```

Target is required (e.g. `api`, `worker`).

| Flag | Short | Default | Description |
|---|---|---|---|
| `--watch` | `-w` | `false` | Watch for changes in `.go` and `.env` files and auto-restart |

#### Examples

```bash
drp run api
drp run api --watch
drp run worker
```

With `--watch`, `drp` monitors all `.go` files (excluding `.git/`,
`tmp/`, `node_modules/`, `database/`) and `.env` for changes and
restarts the process. The watch mode:

- Polls every 800ms for file modifications
- Debounces multiple changes into a single restart
- Gracefully stops the old process before starting the new one
- Waits for the port to become free before restarting

---

### `drp test`

Run all project tests. Runs `go mod tidy` then `go test ./...`.

```
drp test [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--verbose` | `-v` | `false` | Verbose output (each test name) |
| `--cover` | — | `false` | Include coverage report |

#### Examples

```bash
drp test          # compact output
drp test -v       # verbose
drp test --cover   # with coverage
```

---

### `drp build`

Compile a production binary from `./cmd/<target>`.

```
drp build [target] [flags]
```

Target defaults to `"api"` (builds `./cmd/api`).

| Flag | Short | Default | Description |
|---|---|---|---|
| `--output <path>` | `-o` | `./dist/<target>` | Custom output path |

#### Examples

```bash
drp build            # builds ./cmd/api → ./dist/api
drp build worker     # builds ./cmd/worker → ./dist/worker
drp build -o myapp   # builds ./cmd/api → ./myapp
```

---

### `drp docs:generate`

Regenerate swagger documentation. Parses all Swaggo annotations
(`@Summary`, `@Router`, `@Param`, etc.) and produces full
`docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml`.

If the `swag` CLI binary is not installed, it is fetched automatically.

```
drp docs:generate
```

> Run this after adding or modifying handler annotations to keep your
> API documentation in sync with the code.
