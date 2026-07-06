# Commands Reference

Complete reference for every `drp` command.

**Global flags** (apply to all commands):

| Flag | Default | Description |
|---|---|---|
| `--env-file <path>` | `.env` | Path to the `.env` file to load |
| `--no-color` | `false` | Disable coloured terminal output |
| `-h, --help` | — | Show help for any command |

---

## `drp init`

Scaffold a new DRP-backed Go project.

```
drp init [project-name] [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--module <path>` | _(project name)_ | Go module path (e.g. `github.com/acme/myapp`) |
| `--force` | `false` | Overwrite a non-empty directory |

### What it generates (base)

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

### With `--auth`

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

> **Note:** The generated `internal/config/config.go` is a standalone file —
> it does **not** import the `drp` package at runtime.

### Examples

```bash
# Basic usage
drp init myapp

# With authentication preset
drp init myapp --auth

# Custom Go module path
drp init myapp --module github.com/acme/myapp

# Re-scaffold into an existing directory
drp init myapp --force
```

---

## `drp doctor`

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

### Examples

```bash
drp doctor
drp doctor --env-file .env.test
```

---

## Migration Commands

All migration commands read `database/migrations/` and connect to the
database configured in `.env`.

### `drp migrate` / `drp migrate:up`

Run all pending (not yet applied) migrations.

```
drp migrate
drp migrate:up
```

Both commands are equivalent. Prints an info message if there is nothing to run.

---

### `drp migrate:create`

Create a new timestamped migration file pair (up + down).

```
drp migrate:create [name]
```

Creates two files in `database/migrations/`:

```
database/migrations/
  20240101120000_<name>.up.sql
  20240101120000_<name>.down.sql
```

#### Example

```bash
drp migrate:create create_users_table
drp migrate:create add_email_to_orders
```

---

### `drp migrate:down`

Roll back the single most-recently applied migration.

```
drp migrate:down [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be stepped down without executing |

---

### `drp migrate:rollback`

Roll back the **entire last batch** of migrations (all that ran together in the last `migrate:up` call).

```
drp migrate:rollback [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be rolled back without executing |

---

### `drp migrate:fresh`

Drop **all tables**, then re-run all migrations from scratch. Equivalent to `db:reset`.

```
drp migrate:fresh [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be dropped/applied without executing |

> ⚠️ **Destructive.** All data will be lost.

---

### `drp migrate:status`

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

### `drp migrate:seed`

Run all pending migrations, then immediately run the database seeders.
Shortcut for `migrate:up` followed by `db:seed`.

```
drp migrate:seed
```

---

## Seeder Commands

### `drp seeder:create`

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

### `drp db:seed`

Run the database seeders.

```
drp db:seed [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--fresh` | `false` | Clear seed history and re-run **all** seeders |

---

## Database Commands

### `drp db:status`

Show the current database connection status and basic info.

```
drp db:status
```

Output example:

```
✔  Connected to "myapp" on 127.0.0.1:5432 (postgres)
```

---

### `drp db:tables`

List all tables currently in the configured database.

```
drp db:tables
```

---

### `drp db:drop`

Drop **all tables** in the configured database.

```
drp db:drop [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | List tables that would be dropped without executing |

> ⚠️ **Destructive.** All data and schema will be lost.

---

### `drp db:reset`

Drop all tables and re-run all migrations from scratch (alias for `migrate:fresh`).

```
drp db:reset [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print what would be dropped/applied without executing |

---

## Code Generation

### `drp create:crud`

Generate all CRUD layers for a resource inside the current project directory.

```
drp create:crud [name] [flags]
```

With no layer flags, **all five files** are generated (domain-based layout):

```
internal/<domain>/model.go
internal/<domain>/repository.go
internal/<domain>/service.go
internal/<domain>/handler.go
internal/routes/<domain>_routes.go
```

All four domain layers share the same package — no cross-package imports
between model, repository, service, and handler. For example, `drp create:crud product`
creates `internal/product/model.go`, `internal/product/repository.go`, etc.,
all under `package product`.

Use layer flags to generate specific files only:

| Flag | Short | Description |
|---|---|---|
| `--model` | `-m` | Generate model only |
| `--repository` | `-r` | Generate repository only |
| `--service` | `-s` | Generate service only |
| `--handler` | — | Generate handler only |
| `--routes` | — | Generate routes only |
| `--module <path>` | — | Override Go module name (auto-detected from `go.mod` by default) |

#### Examples

```bash
# Generate all layers for "product"
drp create:crud product

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

## `drp test`

Run all project tests. Runs `go mod tidy` then `go test ./...`.

```
drp test [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-v, --verbose` | `false` | Verbose output (each test name) |
| `--cover` | `false` | Include coverage report |

### Examples

```bash
drp test          # compact output
drp test -v       # verbose
drp test -cover   # with coverage
```

---

## `drp build`

Compile a production binary from `./cmd/<target>`.

```
drp build [target] [flags]
```

Target defaults to `"api"` (builds `./cmd/api`).

### Flags

| Flag | Default | Description |
|---|---|---|
| `-o, --output <path>` | `./dist/<target>` | Custom output path |

### Examples

```bash
drp build            # builds ./cmd/api → ./dist/api
drp build worker     # builds ./cmd/worker → ./dist/worker
drp build -o myapp   # builds ./cmd/api → ./myapp
```

---

## `drp docs:generate`

Regenerate swagger documentation. Parses all Swaggo annotations
(`@Summary`, `@Router`, `@Param`, etc.) and produces full
`docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml`.

If the `swag` CLI binary is not installed, it is fetched automatically.

```
drp docs:generate
```

> Run this after adding or modifying handler annotations to keep your
> API documentation in sync with the code.

---

## Other Commands

### `drp version`

Print the current DRP CLI version.

```
drp version
```

### `drp completion`

Generate shell auto-completion scripts (bash, zsh, fish, PowerShell).

```
drp completion [bash|zsh|fish|powershell]
```

#### Example — enable bash completion

```bash
drp completion bash > /etc/bash_completion.d/drp
```
