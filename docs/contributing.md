# Contributing to DRP

Thank you for your interest in contributing! This document covers engineering
standards, how to set up a development environment, and the PR process.

---

## Table of Contents

- [Development Setup](#development-setup)
- [Engineering Standards](#engineering-standards)
- [Project Conventions](#project-conventions)
- [Adding a New Command](#adding-a-new-command)
- [Adding or Modifying Templates](#adding-or-modifying-templates)
- [Testing](#testing)
- [PR Process](#pr-process)
- [Releasing](#releasing)

---

## Development Setup

```bash
# 1. Fork and clone the repository
git clone https://github.com/DarpanAdhikari/drp-go-cli.git
cd drp

# 2. Install Go (1.22 or later)
go version

# 3. Install dependencies
go mod download

# 4. Build the binary
go build -o drp .

# 5. Run the test suite
go test ./...

# 6. Optional: install staticcheck for linting
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

---

## Engineering Standards

All contributions must meet these standards before merging.

### Code quality

| Tool | Command | Requirement |
|---|---|---|
| `gofmt` | `gofmt -l .` | Zero output (no unformatted files) |
| `go vet` | `go vet ./...` | Zero warnings |
| `staticcheck` | `staticcheck ./...` | Zero issues |
| tests | `go test ./...` | All pass |

### Style rules

- **No global state.** All engines and helpers receive their dependencies
  (config, DB connection, directory paths) as arguments — never from package-level
  variables.
- **Exported symbols must have doc comments.** Every exported type, function,
  method, and constant requires a Go doc comment (`// FunctionName ...`).
- **Commands stay thin.** `cmd/` files only parse flags, call `internal/`
  packages, and format output. No business logic in `cmd/`.
- **Errors must be wrapped with context.** Use `fmt.Errorf("package: action: %w", err)`.
- **Do not import `drp` packages from generated code.** Generated files must
  be fully self-contained — they are copied into user projects that do not
  depend on `drp` at runtime.
- **Use `internal/output` for all terminal output.** Do not use `fmt.Println`
  or `log.Printf` directly in command files; use `output.Success`, `output.Fail`,
  `output.Info`, `output.Warn`.

---

## Project Conventions

### Naming

| Context | Convention | Example |
|---|---|---|
| Go packages | lowercase, single word | `migration`, `generator` |
| Go types | PascalCase | `Engine`, `Config` |
| Go functions | PascalCase (exported), camelCase (unexported) | `NewEngine`, `runSeeders` |
| Template files | `<layer>.tpl` | `handler.tpl`, `model.tpl` |
| Migration files | `<timestamp>_<snake_name>.up.sql` | `20240101_create_users_table.up.sql` |
| Seeder files | `<NNN>_<name>.sql` | `001_users.sql` |

### Error messages

- Start with the package name: `"migration: ..."`, `"config: ..."`.
- Wrap underlying errors: `fmt.Errorf("migration: reading file: %w", err)`.
- Where applicable, point the user at `drp doctor` or next steps.

### Directory structure

```
cmd/          → one file per top-level command (or command group)
internal/     → one sub-package per responsibility
templates/    → one .tpl file per generated layer
docs/         → markdown documentation
scripts/      → shell scripts for CI/release
examples/     → sample projects
```

---

## Adding a New Command

1. **Create** `cmd/<command>.go` with a `cobra.Command` variable.
2. **Register** it in that file's `init()` function via `rootCmd.AddCommand(...)`.
3. **Implement** business logic in a new or existing `internal/` package —
   not in the command file itself.
4. **Write tests** for the `internal/` logic.
5. **Document** the command in [`docs/commands.md`](commands.md).

### Minimal command template

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/yourorg/drp/internal/output"
)

var myNewCmd = &cobra.Command{
    Use:   "my:command [arg]",
    Short: "One-line description",
    Args:  cobra.ExactArgs(1),
    RunE: func(c *cobra.Command, args []string) error {
        // parse flags
        // call internal/ package
        // use output.Success / output.Fail / output.Info
        return nil
    },
}

func init() {
    rootCmd.AddCommand(myNewCmd)
    // register flags here
}
```

---

## Adding or Modifying Templates

Templates live in `templates/` and are embedded into the binary at compile
time. They are Go `text/template` files.

**Available template variables** (all derived from the resource name passed to `create:crud`):

| Variable | Description | Example (input: `"product"`) |
|---|---|---|
| `{{.Name}}` | PascalCase singular | `Product` |
| `{{.NameLower}}` | lowercase singular | `product` |
| `{{.NamePlural}}` | PascalCase plural | `Products` |
| `{{.TableName}}` | snake_case plural | `products` |
| `{{.PackageName}}` | Go package name | `product` |
| `{{.ModuleName}}` | Go module path | `github.com/acme/myapp` |

When adding a new template:
1. Create `templates/<layer>.tpl`.
2. Update `internal/generator/crud.go` to render and write the new file.
3. Add a corresponding flag to `cmd/create.go` if you want opt-in generation.
4. Write a test in `internal/generator/crud_test.go`.

---

## Testing

### Unit tests

```bash
go test ./...
go test ./internal/... -v
```

Tests live next to the packages they test (`*_test.go` files in the same directory).

### Running a specific test

```bash
go test ./internal/migration/... -run TestEngine_Up -v
```

### Test coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration / smoke tests

> Currently manual. Before opening a PR, verify the end-to-end flow:

```bash
go build -o drp .
./drp init testapp --force
cd testapp
# Configure .env for a local test database
./drp doctor
./drp migrate:create smoke_test
./drp migrate:up
./drp create:crud widget
cd ..
rm -rf testapp
```

---

## PR Process

1. **Branch from `main`** with a descriptive name: `feat/seeder-rollback`,
   `fix/mysql-drop-tables`, `docs/complete-architecture`.
2. **Make your changes**, following the engineering standards above.
3. **Run the full check** before pushing:
   ```bash
   gofmt -l .
   go vet ./...
   staticcheck ./...
   go test ./...
   ```
4. **Open a PR** against `main`. Fill in the PR description:
   - What problem does this solve?
   - What approach did you take?
   - Any trade-offs or future work?
5. **Address review comments.** Maintainers may request changes — respond
   promptly and push fixes to the same branch.
6. **Squash or rebase** before merge if the branch has noisy WIP commits.

---

## Releasing

Releases are managed via git tags and the `release.yml` GitHub Actions workflow.

```bash
# Tag a new release
git tag v0.2.0
git push origin v0.2.0
```

The workflow cross-compiles binaries for:

| OS | Architectures |
|---|---|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64 |

Binaries are uploaded to the GitHub Release automatically. The version string
is embedded at build time via `-ldflags "-X main.version=v0.2.0"`.

Versioning follows [Semantic Versioning](https://semver.org/). Update
`CHANGELOG.md` before tagging.
