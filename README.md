# DRP — Developer Rapid Productivity

A lightweight, SQL-first CLI for Go backend development. Inspired by Laravel
Artisan. Not a framework. Not an ORM. Just a tool that generates plain,
idiomatic Go code and gets out of your way.

## Why

DRP removes repetitive backend setup work — project scaffolding,
migrations, seeders, CRUD boilerplate — without locking your project into a
runtime dependency. Code DRP generates never imports a `drp` runtime
package, and there's no third-party ORM or migration library: everything is
built on `database/sql`.

## Features

- **Project Scaffolding**: `drp init` sets up a clean, robust standard Go layout.
- **Built-in Auth**: Option for built-in JWT authentication and user management (`drp init --auth`).
- **Migrations**: `drp migrate` creates, runs, and rolls back migrations with zero dependencies.
- **Code Generation**: `drp create:crud` instantly generates models, handlers, repositories, services, and routes.
- **Zero Lock-in**: Output is standard Go code using `database/sql` and standard libraries.

## Quickstart

```bash
# Install DRP
go install github.com/DarpanAdhikari/drp-go-cli/cmd/drp@latest

# Create a new project
drp init myapp
cd myapp
# edit .env
drp migrate:seed
drp create:crud product
drp run api --watch
```

See `docs/getting-started.md`.

## Repository Layout

See `docs/architecture.md` for the full breakdown of `cmd/`, `internal/`,
`templates/`, and `pkg/`.

## License

MIT — see `LICENSE`.
