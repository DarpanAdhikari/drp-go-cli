# DRP — Developer Rapid Productivity

A lightweight, SQL-first CLI for Go backend development. Inspired by Laravel
Artisan. Not a framework. Not an ORM. Just a tool that generates plain,
idiomatic Go code and gets out of your way.

> **Status:** repository skeleton only (Step 1 of the build plan). No
> commands are functional yet — see `docs/architecture.md` and the build
> plan for the implementation roadmap.

## Why

DRP removes repetitive backend setup work — project scaffolding,
migrations, seeders, CRUD boilerplate — without locking your project into a
runtime dependency. Code DRP generates never imports a `drp` runtime
package, and there's no third-party ORM or migration library: everything is
built on `database/sql`.

## Quickstart (target Phase 1 experience)

```bash
drp init myapp
cd myapp
# edit .env
drp migrate:seed
drp create:crud product
```

See `docs/getting-started.md`.

## Repository Layout

See `docs/architecture.md` for the full breakdown of `cmd/`, `internal/`,
`templates/`, and `pkg/`.

## License

MIT — see `LICENSE`.
