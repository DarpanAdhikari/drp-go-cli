# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/), and this
project adheres to [Semantic Versioning](https://semver.org/) starting at v0.1.0.

## [Unreleased]
### Added
- Interactive CRUD generation: `drp create:crud` now launches an interactive
  checklist to select layers (model/repository/service/handler/routes), shows
  a preview, and asks for confirmation before generating.
- Module name prompt: when `go.mod` is missing, DRP prompts for the module
  name interactively instead of failing.
- `--no-interaction` flag: skip interactive prompts for CI/scripting use.
- Non-TTY detection: interactive mode automatically falls back to generating
  all layers when stdout is not a terminal.
- Signal handler: `drp run --watch` now catches Ctrl+C and cleanly stops the
  child process before exiting (no more orphaned servers).

### Changed
- Port killing on macOS/Linux: `stopProcess` now sends `SIGINT` instead of
  `SIGTERM`. `go run` handles `SIGINT` by terminating the child binary before
  exiting, ensuring the port is reliably freed on restart.
- Windows upgrade: `copyExecutable` now renames the existing binary to `.old`
  before placing the new one, fixing "Access denied" errors when upgrading the
  running binary on Windows.
- `go.mod` bumped from Go 1.22 to 1.23 (required by `charmbracelet/huh`).

### Added
- Initial repository skeleton (Step 1 of the build plan): directory layout,
  package stubs, Cobra command scaffolding, CI/release workflow stubs.
