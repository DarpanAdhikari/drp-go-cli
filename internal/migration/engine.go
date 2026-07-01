package migration

import (
	"database/sql"
	"fmt"
	"os"

	drpdb "github.com/DarpanAdhikari/drp-go-cli/internal/db"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
)

// Engine runs migrations against a database, recording state in
// schema_history. It is driver-aware for transactional DDL support.
type Engine struct {
	DB               *sql.DB
	DriverName       string
	MigrationsDir    string
	TransactionalDDL bool // true for Postgres, false for MySQL
}

// NewEngine constructs a migration Engine from a db.Connection.
func NewEngine(conn *drpdb.Connection, migrationsDir string) *Engine {
	return &Engine{
		DB:               conn.DB,
		DriverName:       conn.DriverName,
		MigrationsDir:    migrationsDir,
		TransactionalDDL: conn.Driver.SupportsTransactionalDDL(),
	}
}

// Up runs all pending migrations in timestamp order as a single new batch.
// On Postgres, each file executes inside a transaction — failure rolls back
// cleanly. On MySQL, DDL cannot be rolled back; a clear error is returned
// with a warning that partial changes may have been applied.
func (e *Engine) Up() error {
	files, err := DiscoverFiles(e.MigrationsDir)
	if err != nil {
		return err
	}

	applied, err := drpdb.AppliedMigrations(e.DB)
	if err != nil {
		return err
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, a := range applied {
		appliedSet[a.Migration] = true
	}

	var pending []File
	for _, f := range files {
		if !appliedSet[f.Identifier()] {
			pending = append(pending, f)
		}
	}
	if len(pending) == 0 {
		return ErrNothingToMigrate
	}

	batch, err := drpdb.LatestBatch(e.DB)
	if err != nil {
		return err
	}
	batch++

	for _, f := range pending {
		if err := e.runFile(f.UpPath, f.Identifier(), batch, directionUp); err != nil {
			return err
		}
	}
	return nil
}

// Rollback reverts the most recent batch of applied migrations, executing
// each migration's down file in reverse-application order.
func (e *Engine) Rollback() error {
	batch, err := drpdb.LatestBatch(e.DB)
	if err != nil {
		return err
	}
	if batch == 0 {
		return ErrNothingToRollback
	}

	entries, err := drpdb.MigrationsInBatch(e.DB, e.DriverName, batch)
	if err != nil {
		return err
	}

	files, err := DiscoverFiles(e.MigrationsDir)
	if err != nil {
		return err
	}
	fileIndex := make(map[string]File, len(files))
	for _, f := range files {
		fileIndex[f.Identifier()] = f
	}

	for _, entry := range entries {
		f, ok := fileIndex[entry.Migration]
		if !ok {
			return fmt.Errorf("migration: cannot rollback %q — down file not found on disk", entry.Migration)
		}
		if err := e.runFile(f.DownPath, f.Identifier(), batch, directionDown); err != nil {
			return err
		}
	}
	return nil
}

// Down reverts only the single most-recently applied migration.
func (e *Engine) Down() error {
	applied, err := drpdb.AppliedMigrations(e.DB)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		return ErrNothingToRollback
	}

	last := applied[len(applied)-1]
	files, err := DiscoverFiles(e.MigrationsDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.Identifier() == last.Migration {
			return e.runFile(f.DownPath, f.Identifier(), last.Batch, directionDown)
		}
	}
	return fmt.Errorf("migration: cannot step down %q — down file not found on disk", last.Migration)
}

// Fresh drops all tables (via schema_history-tracked DDL) and re-runs all
// migrations from scratch. When dryRun is true, prints the plan without
// making any changes.
func (e *Engine) Fresh(dryRun bool) error {
	if dryRun {
		files, err := DiscoverFiles(e.MigrationsDir)
		if err != nil {
			return err
		}
		fmt.Println("[dry-run] Would drop all tables and re-run migrations:")
		for _, f := range files {
			fmt.Printf("  → %s\n", f.Identifier())
		}
		return nil
	}

	// Drop and recreate schema_history, then run all up migrations.
	if _, err := e.DB.Exec(`DROP TABLE IF EXISTS schema_history`); err != nil {
		return fmt.Errorf("migration: fresh — dropping schema_history: %w", err)
	}
	if err := drpdb.EnsureSchemaHistoryTable(e.DB, e.DriverName); err != nil {
		return err
	}
	return e.Up()
}

// Status returns a list of all known migrations with their applied/pending state.
func (e *Engine) Status() ([]StatusEntry, error) {
	files, err := DiscoverFiles(e.MigrationsDir)
	if err != nil {
		return nil, err
	}

	applied, err := drpdb.AppliedMigrations(e.DB)
	if err != nil {
		return nil, err
	}
	appliedMap := make(map[string]int, len(applied))
	for _, a := range applied {
		appliedMap[a.Migration] = a.Batch
	}

	entries := make([]StatusEntry, 0, len(files))
	for _, f := range files {
		id := f.Identifier()
		batch, ok := appliedMap[id]
		entries = append(entries, StatusEntry{
			Migration: id,
			Applied:   ok,
			Batch:     batch,
		})
	}
	return entries, nil
}

// direction tells runFile whether to record or remove the migration from history.
type direction int

const (
	directionUp   direction = 1
	directionDown direction = -1
)

// runFile executes the SQL in path and updates schema_history accordingly.
// On Postgres it wraps execution in a transaction; on MySQL it executes
// directly with a warning comment that partial changes cannot be rolled back.
func (e *Engine) runFile(path, identifier string, batch int, dir direction) error {
	sql, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("migration: reading %q: %w", path, err)
	}

	if dir == directionUp {
		output.Info("Migrating: %s", identifier)
	} else {
		output.Info("Rolling back: %s", identifier)
	}

	var runErr error
	if e.TransactionalDDL {
		runErr = e.runInTransaction(string(sql), identifier, batch, dir)
	} else {
		runErr = e.runDirect(string(sql), identifier, batch, dir)
	}

	if runErr != nil {
		return runErr
	}

	if dir == directionUp {
		output.Success("Migrated:  %s", identifier)
	} else {
		output.Success("Rolled back: %s", identifier)
	}
	return nil
}

// runInTransaction wraps a migration file's SQL in a transaction (Postgres).
// schema_history is updated inside the same transaction so both the schema
// change and the history record succeed or fail atomically.
func (e *Engine) runInTransaction(sqlStr, identifier string, batch int, dir direction) error {
	tx, err := e.DB.Begin()
	if err != nil {
		return fmt.Errorf("migration: beginning transaction for %q: %w", identifier, err)
	}

	if _, err := tx.Exec(sqlStr); err != nil {
		tx.Rollback()
		return fmt.Errorf("migration: executing %q: %w", identifier, err)
	}

	if dir == directionUp {
		q := `INSERT INTO schema_history (migration, batch) VALUES ($1, $2)`
		if _, err := tx.Exec(q, identifier, batch); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration: recording %q in schema_history: %w", identifier, err)
		}
	} else {
		q := `DELETE FROM schema_history WHERE migration = $1`
		if _, err := tx.Exec(q, identifier); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration: removing %q from schema_history: %w", identifier, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migration: committing %q: %w", identifier, err)
	}
	return nil
}

// runDirect executes a migration file outside a transaction (MySQL).
// If execution fails, a warning explains that partial DDL changes may have
// been committed and the user must inspect and repair manually.
func (e *Engine) runDirect(sqlStr, identifier string, batch int, dir direction) error {
	if _, err := e.DB.Exec(sqlStr); err != nil {
		return fmt.Errorf(
			"migration: executing %q: %w\n"+
				"  ⚠ MySQL does not support transactional DDL — some statements in this\n"+
				"  file may have committed. Inspect the schema manually before retrying.",
			identifier, err,
		)
	}

	if dir == directionUp {
		if err := drpdb.RecordMigration(e.DB, e.DriverName, identifier, batch); err != nil {
			return err
		}
	} else {
		if err := drpdb.RemoveMigration(e.DB, e.DriverName, identifier); err != nil {
			return err
		}
	}
	return nil
}

// Sentinel errors returned by Engine methods so callers can inspect them.
var (
	ErrNothingToMigrate  = fmt.Errorf("migration: nothing to migrate — all migrations are up to date")
	ErrNothingToRollback = fmt.Errorf("migration: nothing to roll back — no migrations have been applied")
)
