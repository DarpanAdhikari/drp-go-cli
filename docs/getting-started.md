# Getting Started

A 5-minute guide to go from zero to a running DRP-generated Go backend.

## Prerequisites

| Requirement | Minimum version | Check |
|---|---|---|
| Go | 1.22+ | `go version` |
| PostgreSQL **or** MySQL | any recent | `psql --version` / `mysql --version` |
| Git | any | `git --version` |

---

## Step 1 — Install the `drp` CLI

```bash
go install github.com/DarpanAdhikari/drp-go-cli/cmd/drp@latest
```

Verify the install:

```bash
drp version
drp --help
```

> **Note:** `drp` does **not** need to be imported into your project. It is a
> standalone developer tool — like `git` or `make` — that generates plain Go
> files and then gets out of the way.

---

## Step 2 — Check your environment

Before creating a project, confirm that your system is ready:

```bash
drp doctor
```

`doctor` checks:

- Go installation and version
- Presence and validity of a `.env` file
- Database reachability
- Expected project directory structure

Fix any ❌ items it reports before continuing.

---

## Step 3 — Scaffold a new project

```bash
drp init myapp
cd myapp
```

For a project with built-in authentication (JWT, user management, middleware):

```bash
drp init myapp --auth
cd myapp
```

What gets created (with `--auth`):

```
myapp/
├── cmd/api/main.go          ← runnable HTTP server entry point
├── internal/
│   ├── auth/                ← JWT, token store, auth handlers
│   ├── config/config.go     ← standalone config loader (no drp import)
│   ├── middleware/           ← CORS, rate limiting, request ID, auth
│   ├── routes/
│   ├── shared/              ← base types, context helpers, response helpers
│   └── user/                ← user model, repository, service, handler
├── database/
│   ├── migrations/
│   └── seeders/
├── dist/                    ← production binaries (drp build)
├── docs/                    ← swagger documentation (drp docs:generate)
├── pkg/
├── templates/
├── .env                     ← your local secrets (git-ignored)
├── .env.example             ← safe to commit
├── .gitignore
├── go.mod
└── README.md
```

> **Custom module path:** if your GitHub path differs from the project name,
> pass `--module`:
>
> ```bash
> drp init myapp --module github.com/acme/myapp
> ```

---

## Step 4 — Configure the database

Open `.env` and fill in your real credentials:

```dotenv
# Database configuration
DB_DRIVER=postgres        # "postgres" or "mysql"
DB_HOST=127.0.0.1
DB_PORT=5432              # 5432 for postgres, 3306 for mysql
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=myapp
DB_SSLMODE=disable        # postgres only

# Application
APP_PORT=8080
```

Required variables: `DB_USER`, `DB_NAME`.  
All others have sensible defaults.

---

## Step 4b — Run tests

Your project (especially with `--auth`) comes with ready-to-run tests:

```bash
drp test        # runs all tests
drp test -v     # verbose output
```

This runs `go mod tidy` and `go test ./...` for you.

---

## Step 5 — Create and run your first migration

```bash
# Create a migration file pair (up + down)
drp migrate:create create_products_table

# Edit the generated SQL files in database/migrations/
# Then apply them:
drp migrate:up

# Or check what has run / is pending first:
drp migrate:status
```

---

## Step 6 — Seed the database

```bash
# Create a seeder file
drp seeder:create products

# Edit database/seeders/products.sql
# Then run migrations + seeders together:
drp migrate:seed
```

---

## Step 7 — Generate a full CRUD resource

```bash
drp create:crud product
```

This generates five files under your project (domain-based layout):

```
internal/product/model.go
internal/product/repository.go
internal/product/service.go
internal/product/handler.go
internal/routes/product_routes.go
```

Then register the routes in `cmd/api/main.go`:

```go
routes.RegisterProductRoutes(mux, db)
```

---

## Step 8 — Run the API

```bash
drp run api --watch
# Watching for changes...
# Starting go run ./cmd/api
```

---

## Full quickstart in one block

```bash
drp init myapp
cd myapp
# edit .env with your DB credentials
drp doctor                          # verify everything is OK
drp migrate:create create_users_table
# edit the generated SQL
drp migrate:seed                    # migrate + seed
drp create:crud user                # generate CRUD for "user"
drp run api --watch                 # start the server with hot-reloading
```

---

## Step 9 — Deploying to Production

When you are ready to take your DRP-generated backend to production, do not use `drp run api`. Instead, compile your application into a standalone binary:

1. **Build the binary:**
   ```bash
   drp build
   # or with a custom name:
   drp build -o myapi
   ```
   This gives you a single, executable file (`myapi`) that contains everything your app needs.

2. **Set Environment Variables:**
   In production, you should pass configuration securely via environment variables (e.g. in your Systemd service, Docker container, or Kubernetes deployment) rather than checking in a `.env` file. However, if an `.env` file is present next to the binary, the application will still load it.

3. **Run migrations (optional but recommended in deployment pipeline):**
   ```bash
   # You must have DRP installed on your deployment server/pipeline
   drp migrate:up
   ```
   *(Note: You can also choose to run migrations as part of your CI/CD pipeline rather than directly on the app server).*

4. **Run the binary:**
   ```bash
   ./myapi
   ```

---

## Step 9b — Generate API documentation (auth preset only)

If you used `--auth`, your project includes Swaggo annotations on all
handlers. Generate the swagger UI docs:

```bash
drp docs:generate
```

This installs the `swag` CLI if missing, parses all `@Router`, `@Param`,
etc. annotations, and produces the OpenAPI spec under `docs/`. View the
UI at `http://localhost:8080/swagger/` when the API is running.

---

## Useful flags

| Flag | Applies to | Effect |
|---|---|---|
| `--env-file <path>` | all commands | Use a custom `.env` file instead of `.env` |
| `--no-color` | all commands | Disable coloured terminal output |
| `--force` | `init` | Overwrite a non-empty directory |
| `--module <path>` | `init`, `create:crud` | Set / override Go module path |
| `--dry-run` | `migrate:rollback`, `migrate:down`, `migrate:fresh`, `db:drop`, `db:reset` | Preview what *would* happen without making changes |
| `--watch`, `-w` | `run` | Watch for changes in `.go` and `.env` files and auto-restart |

---

## Next steps

- See [`commands.md`](commands.md) for the full command reference.
- See [`architecture.md`](architecture.md) for how DRP is structured internally.
- See [`swagger_settings.md`](swagger_settings.md) for how to view and customise your API documentation.
- See [`test_generation.md`](test_generation.md) for how to run and extend the generated test suite.
- See [`contributing.md`](contributing.md) if you want to add commands or templates.
