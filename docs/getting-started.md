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

What gets created:

```
myapp/
├── cmd/api/main.go          ← runnable HTTP server entry point
├── internal/
│   ├── config/config.go     ← standalone config loader (no drp import)
│   ├── handlers/
│   ├── repositories/
│   ├── services/
│   ├── routes/
│   └── models/
├── database/
│   ├── migrations/
│   └── seeders/
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

This generates five files under your project:

```
internal/models/product.go
internal/repositories/product_repository.go
internal/services/product_service.go
internal/handlers/product_handler.go
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
   go build -o myapi ./cmd/api
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
- See [`contributing.md`](contributing.md) if you want to add commands or templates.
