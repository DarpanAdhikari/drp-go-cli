package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SchemaHistoryEntry represents one applied migration recorded in
// the schema_history table.
type SchemaHistoryEntry struct {
	ID        int64
	Migration string
	Batch     int
	AppliedAt time.Time
}

// createTableSQL returns the CREATE TABLE IF NOT EXISTS statement for
// schema_history, adapted for the given driver.
func createTableSQL(driverName string) string {
	switch driverName {
	case "mysql":
		return `
			CREATE TABLE IF NOT EXISTS schema_history (
				id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
				migration  VARCHAR(255) NOT NULL UNIQUE,
				batch      INT          NOT NULL,
				applied_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`
	default: // postgres
		return `
			CREATE TABLE IF NOT EXISTS schema_history (
				id         BIGSERIAL    PRIMARY KEY,
				migration  VARCHAR(255) NOT NULL UNIQUE,
				batch      INT          NOT NULL,
				applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
			)`
	}
}

// placeholder returns the positional parameter token for the given driver
// and position (1-based): "$1" for Postgres, "?" for MySQL.
func placeholder(driverName string, _ int) string {
	if driverName == "mysql" {
		return "?"
	}
	return "$1" // caller must use correct index for multi-arg queries; extend as needed
}

// EnsureSchemaHistoryTable creates the schema_history table if it does
// not already exist. driverName must be "postgres" or "mysql".
func EnsureSchemaHistoryTable(db *sql.DB, driverName string) error {
	if _, err := db.Exec(createTableSQL(driverName)); err != nil {
		return fmt.Errorf("schema_history: create table: %w", err)
	}
	return nil
}

// AppliedMigrations returns all rows from schema_history ordered by id,
// representing the full applied migration history in application order.
func AppliedMigrations(db *sql.DB) ([]SchemaHistoryEntry, error) {
	rows, err := db.Query(
		`SELECT id, migration, batch, applied_at FROM schema_history ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("schema_history: querying applied migrations: %w", err)
	}
	defer rows.Close()

	var entries []SchemaHistoryEntry
	for rows.Next() {
		var e SchemaHistoryEntry
		if err := rows.Scan(&e.ID, &e.Migration, &e.Batch, &e.AppliedAt); err != nil {
			return nil, fmt.Errorf("schema_history: scanning row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// LatestBatch returns the highest batch number in schema_history, or 0
// if no migrations have been applied yet.
func LatestBatch(db *sql.DB) (int, error) {
	var batch sql.NullInt64
	if err := db.QueryRow(`SELECT MAX(batch) FROM schema_history`).Scan(&batch); err != nil {
		return 0, fmt.Errorf("schema_history: reading latest batch: %w", err)
	}
	if !batch.Valid {
		return 0, nil
	}
	return int(batch.Int64), nil
}

// MigrationsInBatch returns all entries for the given batch number ordered
// by id descending — i.e. reverse application order, ready for rollback.
func MigrationsInBatch(db *sql.DB, driverName string, batch int) ([]SchemaHistoryEntry, error) {
	q := `SELECT id, migration, batch, applied_at FROM schema_history WHERE batch = $1 ORDER BY id DESC`
	if driverName == "mysql" {
		q = `SELECT id, migration, batch, applied_at FROM schema_history WHERE batch = ? ORDER BY id DESC`
	}

	rows, err := db.Query(q, batch)
	if err != nil {
		return nil, fmt.Errorf("schema_history: querying batch %d: %w", batch, err)
	}
	defer rows.Close()

	var entries []SchemaHistoryEntry
	for rows.Next() {
		var e SchemaHistoryEntry
		if err := rows.Scan(&e.ID, &e.Migration, &e.Batch, &e.AppliedAt); err != nil {
			return nil, fmt.Errorf("schema_history: scanning row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// RecordMigration inserts a new schema_history entry for an applied migration.
func RecordMigration(db *sql.DB, driverName, migration string, batch int) error {
	q := `INSERT INTO schema_history (migration, batch) VALUES ($1, $2)`
	if driverName == "mysql" {
		q = `INSERT INTO schema_history (migration, batch) VALUES (?, ?)`
	}
	if _, err := db.Exec(q, migration, batch); err != nil {
		return fmt.Errorf("schema_history: recording %q: %w", migration, err)
	}
	return nil
}

// RemoveMigration deletes the schema_history row for a given migration
// identifier. Used during rollback to keep history in sync with actual state.
func RemoveMigration(db *sql.DB, driverName, migration string) error {
	q := `DELETE FROM schema_history WHERE migration = $1`
	if driverName == "mysql" {
		q = `DELETE FROM schema_history WHERE migration = ?`
	}
	if _, err := db.Exec(q, migration); err != nil {
		return fmt.Errorf("schema_history: removing %q: %w", migration, err)
	}
	return nil
}
