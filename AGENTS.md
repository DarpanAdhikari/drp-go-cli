# AGENTS.md — drp-go-cli

This file preserves the full context for AI agents to continue work without losing the plan.

---

## Project

`drp-go-cli` is a CLI tool (`drp`) that scaffolds Go projects. Its template system lives in `internal/generator/`.

### Key files

| File | Purpose |
|---|---|
| `internal/generator/project.go` | `drp init` — creates project dirs + base files |
| `internal/generator/auth_preset.go` | `drp init --auth` — adds auth/user/middleware files |
| `internal/generator/crud.go` | `drp create:crud` — generates CRUD layers for a resource |
| `internal/generator/project_test.go` | Tests for init + auth preset |
| `templates/*.tpl` | User-overridable CRUD templates (disk) |
| `internal/generator/embedded/files/*.tpl` | Embedded CRUD template fallbacks |
| `internal/generator/render.go` | Resolves templates (disk over embedded) |

---

## What's been done: Phase 1 (Tier 1+2)

### Flat-layer → Domain-based structure

**Before** `drp init --auth` generated:
```
internal/handlers/     ← auth_handler.go, user_handler.go, helpers.go
internal/models/       ← user.go
internal/repositories/ ← user_repository.go
internal/services/     ← user_service.go
internal/auth/         ← jwt.go, token_store.go
internal/middleware/
internal/routes/
internal/config/
```

**After:**
```
internal/shared/       ← base.go, context.go, response.go
internal/auth/         ← jwt.go, token_store.go, handler.go
internal/user/         ← model.go, repository.go, service.go, handler.go
internal/middleware/   ← auth.go, cors.go, requestid.go, rate_limiter.go, rate_limit.go
internal/routes/
internal/config/
```

### Features shipped

1. **Minimal JWT** — no PII (Email removed from Claims). `GenerateTokenPair(userID int64, cfg)`.
2. **UpsertDeviceSession** — finds existing active session by (user_id + mac_address) or (user_id + user_agent), updates in-place preserving `trust_level`, `authorized_at`, `created_at`. Falls back to INSERT.
3. **Input validation** — `go-playground/validator/v10` struct tags on all request DTOs.
4. **Password policy** — 8+ chars, must contain uppercase + lowercase + digit (in `user/service.go`).
5. **Structured responses** — `shared.WriteJSON`, `shared.WriteError`, `shared.DecodeJSON`.
6. **Graceful shutdown** — `signal.NotifyContext(SIGINT, SIGTERM)` with 10s timeout.
7. **Structured logging** — `slog` everywhere, level configurable via `LOG_LEVEL` env var.
8. **CORS middleware** — configurable origins via `CORS_ORIGINS` env (comma-separated; `*` = allow all).
9. **Request ID middleware** — reads `X-Request-ID` header or generates 16-byte hex UUID.
10. **Rate limiting** — per-IP token bucket (`golang.org/x/time/rate`), toggled via `RATE_LIMIT_ENABLED` env. Rate/burst configurable.
11. **Updated migration schema** — `mac_address`, `fcm_token`, `trust_level`, `authorized_at`, `updated_at` columns on `user_tokens`.

### Generated env vars (new beyond base)

```
CORS_ORIGINS=*
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
LOG_LEVEL=info
```

### Generated go.mod deps (new beyond base)

```
github.com/go-playground/validator/v10 v10.20.0
golang.org/x/time v0.5.0
```

---

## Architecture decisions & gotchas

### Circular import: `auth → user → middleware → auth`

**The cycle:**
- `auth/handler.go` imports `user` (for `*user.Service`, request DTOs)
- `user/handler.go` imports `middleware` (for `UserID()`)
- `middleware/auth.go` imports `auth` (for `ParseToken()`)

**Solution:** Move `UserID()`, `TokenJTI()`, and context key constants out of `middleware/auth.go` into `shared/context.go`. Now:
- `user/handler.go` imports only `shared` (no cycle)
- `middleware/auth.go` imports `shared` for the keys
- `auth/handler.go` does NOT import `middleware` at all
- Logout handler in `auth/handler.go` parses the access token from the Authorization header directly instead of reading it from context via `middleware.TokenJTI(r)`

### Auth handler vs middleware boundary

- `middleware/auth.go` — sets context keys (`shared.UserIDKey`, `shared.TokenJTIKey`). Does NOT contain `UserID()` / `TokenJTI()` helper functions (those are in `shared/context.go`).
- `auth/handler.go` — all auth endpoints. Imports `user` for `user.Service`. Does NOT import `middleware` (breaks cycle).
- `user/handler.go` — GET /me endpoint only. Uses `shared.UserID(r)` to get authenticated user.

### Template function patterns

- Functions that **don't** reference the module name (`%s`): return plain Go raw strings (`return \`...\``)
- Functions that **do** reference the module name: use `fmt.Sprintf` with backtick template and `mod` parameter
- Never import `"time"` in generated `user/model.go` — `shared.Base` provides `time.Time` fields via embedding
- Never import `"strings"` in generated `middleware/cors.go` — it's not needed there

### File naming in `authProjectFiles()`

All generated file paths use `filepath.Join()` for cross-platform compatibility. The key paths:

```
internal/shared/base.go        → sharedBaseGo()
internal/shared/context.go     → sharedContextGo()
internal/shared/response.go    → sharedResponseGo()
internal/user/model.go         → userModelGo(mod)
internal/user/repository.go    → userRepositoryGo()
internal/user/service.go       → userServiceGo()
internal/user/handler.go       → userHandlerGo(mod)
internal/auth/jwt.go           → authJWTGo(mod)
internal/auth/token_store.go   → authTokenStoreGo()
internal/auth/handler.go       → authHandlerGo(mod)
internal/middleware/auth.go    → authMiddlewareGo(mod)
internal/middleware/cors.go    → corsMiddlewareGo(mod)
internal/middleware/requestid.go → requestIDMiddlewareGo(mod)
internal/middleware/rate_limiter.go → rateLimiterGo()
internal/middleware/rate_limit.go → rateLimitMiddlewareGo()
internal/routes/routes.go      → authRoutesGo(mod)
internal/config/config.go      → authConfigGo(mod, name)
cmd/api/main.go                → authMainGo(mod)
database/migrations/000001_... → authUsersMigrationUp()
```

---

## Phase 2 (Tier 3) — Complete ✅

### 3. CRUD template update ✅

The `templates/*.tpl` and `internal/generator/crud.go` now generate into domain-based directories:
- `internal/<domain>/model.go`
- `internal/<domain>/repository.go`
- `internal/<domain>/service.go`
- `internal/<domain>/handler.go`
- `internal/routes/<domain>_routes.go`

All layers within a domain share the same package (named after the singular snake_case domain name), so there are no cross-package imports between model/repo/service/handler. The routes layer is the only one that imports the domain package. A new `DomainName` field was added to the `Names` struct for this purpose.

---

## Test strategy

1. Run `go test ./internal/generator/` in drp-go-cli itself
2. Run `drp init myapp --auth` in a temp dir
3. `cd myapp && go mod tidy && go build ./...` — must compile
4. For template content correctness: use `strings.Contains` on generated files
5. For template compilation: install drp binary, generate project, build it

**Key test file:** `internal/generator/project_test.go` — checks dirs, files, .env vars, route strings

---

## Custom commands (convenience layer)

Three commands wrap common Go tasks so users don't need to remember raw Go CLI:

| Command | What it does |
|---|---|
| `drp test` | Runs `go mod tidy` then `go test ./...` (flags: `-v`, `--cover`) |
| `drp build [target]` | Builds `./cmd/<target>` → `./dist/<target>` (default: `api`) |
| `drp docs:generate` | Auto-installs `swag` CLI if missing, runs `swag init -g cmd/api/main.go` |

---

## Phase 3 (Session 3) — Complete ✅

### Features shipped
1. **`drp completion [bash|zsh|fish|powershell]`** — shell autocomplete via Cobra built-ins (`cmd/completion.go`)
2. **Auto `git init`** — runs after project creation (`cmd/init.go`)
3. **Health check endpoint** — `GET /healthz` always generated (not tied to `--auth`). Uses `database/sql.Ping()` on every request (`internal/routes/health.go`). Registered in both base `main.go` and auth `main.go`.
4. **Infrastructure generation** (`--infra` flag):
   - `"docker"` → `docker/Dockerfile`, `.dockerignore`, `docker-compose.yml`
   - `"ci"` → `.github/workflows/ci.yml` (Go CI workflow)
   - `"make"` → `Makefile` (build, test, lint, run targets)
   - `"lint"` → `.editorconfig`, `.golangci.yml`
   - `"all"` → all of the above
5. **MySQL driver support** (`--driver` flag):
   - `--driver postgres` (default): uses `github.com/lib/pq`, BIGSERIAL/TIMESTAMPTZ in migrations, `host=...` DSN format
   - `--driver mysql`: uses `github.com/go-sql-driver/mysql`, BIGINT UNSIGNED AUTO_INCREMENT/DATETIME(6) in migrations, `user:pass@tcp(...)` DSN format
   - Driver-aware defaults: DB_PORT (5432/3306), DB_USER (postgres/root)
   - `.env` and `.env.example` are driver-aware
   - Auth template functions updated: `authMainGo(mod, drv)`, `authConfigGo(mod, name, drv)`, `authUsersMigrationUp(drv)`

### What changed

| File | Changes |
|---|---|
| `cmd/completion.go` | New — shell completion command |
| `cmd/init.go` | New `--infra`, `--driver` flags; calls `git init` post-creation |
| `internal/generator/project.go` | Complete rewrite — `ProjectOptions.Infra`, `DBDriver`; driver-aware base config/main.go templates; infrastructure files (Docker, CI, Makefile, lint); health check route; `dbImport()`, `dsnTemplate()`, `dsnArgs()` helpers |
| `internal/generator/auth_preset.go` | Updated — `authMainGo`, `authConfigGo`, `authUsersMigrationUp` all accept `drv` parameter; driver-aware .env/DSN/migration SQL; health route registration |
| `internal/generator/project_test.go` | Updated — tests for new infra files, driver-aware generation |

### Key patterns
- **`--infra` flag**: single string `"all"` for everything, comma-separated for specific components (e.g. `"docker,ci"`), empty for none.
- **`DBDriver` field**: `ProjectOptions.DBDriver` set via `--driver` flag (default `"postgres"`). Not parsed from .env.
- **Template driver params**: Auth template functions now take a `drv string` parameter to conditionally generate mysql vs postgres code paths.
- **Health check**: Always generated via `routes.RegisterHealthRoute(mux, db)` in both base and auth main.go.

### Generated env vars (Phase 3 additions)
```
# Docker/CI/Makefile — no additional env vars beyond base + auth
```

### Generated go.mod deps (Phase 3 additions)
```
github.com/go-sql-driver/mysql v1.8.1  # only when --driver mysql
```

## How to continue

1. Open `internal/generator/auth_preset.go` or `internal/generator/project.go`
2. Add new template functions or modify existing ones
3. Wire new files into `authProjectFiles()` or `projectFiles()` maps
4. Update `internal/generator/project_test.go` for new expected files
5. Run `go install ./cmd/drp && go test ./...`
6. Generate a project in /tmp to verify it compiles
7. For driver-aware changes: test both `--driver postgres` and `--driver mysql`
