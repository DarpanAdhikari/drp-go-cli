package seeder

import (
	"database/sql"
	"fmt"
	"os"

	drpdb "github.com/DarpanAdhikari/drp-go-cli/internal/db"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
)

// Engine runs seeder files against a database, tracking which seeders have
// already run in the seed_history table.
type Engine struct {
	DB         *sql.DB
	DriverName string
	SeedersDir string
}

// NewEngine constructs a seeder Engine from a db.Connection.
func NewEngine(conn *drpdb.Connection, seedersDir string) *Engine {
	return &Engine{
		DB:         conn.DB,
		DriverName: conn.DriverName,
		SeedersDir: seedersDir,
	}
}

// Run executes all pending seeder files (those not already in seed_history)
// in timestamp order. If fresh is true, seed_history is cleared first so
// all seeders re-run regardless of prior state.
//
// Unlike migrations, seeders do not use transactional DDL wrapping — they
// contain INSERT statements which are data-only and easier to reason about
// without transactions. If a seeder fails mid-file, seed_history is NOT
// updated for that seeder so it can safely be re-run after fixing the data.
func (e *Engine) Run(fresh bool) error {
	if err := EnsureSeedHistoryTable(e.DB, e.DriverName); err != nil {
		return err
	}

	if fresh {
		if err := ClearSeedHistory(e.DB); err != nil {
			return fmt.Errorf("seeder: clearing history for fresh run: %w", err)
		}
	}

	files, err := DiscoverFiles(e.SeedersDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return ErrNoSeeders
	}

	applied, err := AppliedSeeders(e.DB)
	if err != nil {
		return err
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, a := range applied {
		appliedSet[a.Seeder] = true
	}

	ran := 0
	for _, f := range files {
		if appliedSet[f.Identifier()] {
			continue
		}
		if err := e.runFile(f); err != nil {
			return err
		}
		ran++
	}

	if ran == 0 {
		return ErrNothingToSeed
	}
	return nil
}

// runFile executes the SQL in a single seeder file and records it in
// seed_history on success.
func (e *Engine) runFile(f File) error {
	output.Info("Seeding: %s", f.Identifier())

	sqlBytes, err := os.ReadFile(f.Path)
	if err != nil {
		return fmt.Errorf("seeder: reading %q: %w", f.Path, err)
	}

	if _, err := e.DB.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("seeder: executing %q: %w", f.Identifier(), err)
	}

	if err := RecordSeeder(e.DB, e.DriverName, f.Identifier()); err != nil {
		return err
	}

	output.Success("Seeded:  %s", f.Identifier())
	return nil
}

// Sentinel errors so callers can distinguish no-op runs from real errors.
var (
	ErrNoSeeders     = fmt.Errorf("seeder: no seeder files found in seeders directory")
	ErrNothingToSeed = fmt.Errorf("seeder: nothing to seed — all seeders have already run")
)
