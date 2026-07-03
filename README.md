# DRP — Developer Rapid Productivity

A lightweight, SQL-first CLI for Go backend development. Inspired by Laravel
Artisan. Not a framework. Not an ORM. Just a tool that generates plain,
idiomatic Go code and gets out of your way.

## Features

- **Project Scaffolding** — `drp init` creates a clean, production-ready Go
  project with config, database, migrations, and a runnable API entry point.
  Optionally includes JWT authentication (`drp init --auth`).
- **Migrations** — `drp migrate` creates, runs, and rolls back database
  migrations with zero dependencies on `database/sql`.
- **Seeders** — `drp db:seed` populates your database with test data.
- **Interactive CRUD Generation** — `drp create:crud` uses an interactive
  checklist to select which layers to generate (model, repository, service,
  handler, routes). Generates plain Go files — no runtime dependency on DRP.
- **Hot Reload** — `drp run api --watch` watches `.go` and `.env` files and
  auto-restarts your server. Reliable cross-platform process termination with
  automatic port management.
- **Production Readiness Checks** — `drp doctor` validates your Go
  environment, database connectivity, and project structure.
- **Self-Upgrade** — `drp upgrade` downloads the latest release from GitHub
  and replaces the running binary. Works on Windows, macOS, and Linux.
- **SQL-First Design** — All generated code uses `database/sql` directly.
  No third-party ORM, no migration library, no framework lock-in.

## Installation

### Pre-built binary (recommended)

Download the latest release for your platform from the
[releases page](https://github.com/DarpanAdhikari/drp-go-cli/releases).

Or use the built-in upgrade command once installed:

```bash
drp upgrade
```

### From source

```bash
go install github.com/DarpanAdhikari/drp-go-cli/cmd/drp@latest
```

Requires Go 1.23+ to build.

## Quickstart

```bash
# Create a new project
drp init myapp
cd myapp

# Edit .env with your database credentials
# Then run migrations and seeders
drp migrate:seed

# Generate CRUD for a resource (interactive)
drp create:crud product

# Run the API with hot reload
drp run api --watch
```

The `create:crud` command launches an interactive checklist:

```
Generate CRUD for "product" — which layers? (space to toggle)
  ◉ Model
  ◉ Repository
  ◉ Service
  ◉ Handler
  ◉ Routes
```

After selecting layers, it shows a preview before generating:

```
The following files will be created:
  ✓ internal/models/products.go
  ✓ internal/repositories/product_repository.go
  ✓ internal/services/product_service.go
  ✓ internal/handlers/product_handler.go
  ✓ internal/routes/product_routes.go

Proceed with generation?  [Yes]  [Cancel]
```

For scripting or non-TTY environments, use `--no-interaction`:

```bash
drp create:crud product --no-interaction       # all layers
drp create:crud product -m -r --no-interaction  # model + repository only
```

## Commands

| Command | Description |
|---------|-------------|
| `drp init [name]` | Scaffold a new project |
| `drp init [name] --auth` | Scaffold with JWT authentication |
| `drp create:crud [name]` | Generate CRUD layers (interactive or `--no-interaction`) |
| `drp run [target]` | Run `./cmd/<target>` |
| `drp run [target] --watch` | Run with hot reload |
| `drp migrate` | Run pending migrations |
| `drp migrate:create` | Create a new migration |
| `drp migrate:rollback` | Roll back the last batch |
| `drp migrate:fresh` | Drop all tables and re-run all migrations |
| `drp migrate:status` | Show migration status |
| `drp db:seed` | Run database seeders |
| `drp db:status` | Check database connection |
| `drp db:tables` | List database tables |
| `drp db:drop` | Drop all tables |
| `drp doctor` | Check environment, database, and project health |
| `drp install` | Install the current binary to `~/.local/bin` |
| `drp upgrade` | Download and replace with the latest release |
| `drp version` | Print version |

## Requirements

- **To run generated projects**: Go 1.22+ and a running PostgreSQL or MySQL
  database.
- **To build DRP from source**: Go 1.23+
- **To use pre-built binaries**: No Go installation needed — binaries are
  statically compiled.

## Project Structure

A project created by `drp init` follows standard Go conventions:

```
myapp/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Environment configuration
│   ├── handlers/            # HTTP handlers
│   ├── middleware/           # HTTP middleware
│   ├── models/              # Database models
│   ├── repositories/        # Database access layer
│   ├── routes/              # Route registrations
│   └── services/            # Business logic
├── database/
│   ├── migrations/          # SQL migration files
│   └── seeders/             # SQL seeder files
├── templates/               # Custom template overrides
├── .env                     # Local environment variables
├── .env.example             # Committable env template
├── go.mod
└── README.md
```

All generated CRUD code lives under `internal/` and uses `database/sql`
directly — no runtime dependency on DRP after generation.

## Architecture

DRP is designed with zero lock-in as a core principle:

1. **Generated code never imports DRP**. Your project has no dependency on
   this tool after initial generation.
2. **Templates are overrideable**. Place a `.tpl` file in your project's
   `templates/` directory and DRP will use it instead of the embedded default.
3. **SQL-first**. Migrations, seeders, and queries are raw SQL. The migration
   engine tracks state in a `schema_history` table using standard DDL.

See `docs/architecture.md` for details.

## Contributing

Pull requests are welcome. For major changes, open an issue first to discuss.

Make sure tests pass:

```bash
go test ./...
```

## License

MIT — see `LICENSE`.
