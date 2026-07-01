package seeder

import (
	"database/sql"
	"fmt"
	"time"
)

// HistoryEntry represents one applied seeder recorded in seed_history.
type HistoryEntry struct {
	ID        int64
	Seeder    string
	AppliedAt time.Time
}

// EnsureSeedHistoryTable creates the seed_history table if it does not
// already exist. driverName must be "postgres" or "mysql".
func EnsureSeedHistoryTable(db *sql.DB, driverName string) error {
	var q string
	switch driverName {
	case "mysql":
		q = `CREATE TABLE IF NOT EXISTS seed_history (
			id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
			seeder     VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		q = `CREATE TABLE IF NOT EXISTS seed_history (
			id         BIGSERIAL    PRIMARY KEY,
			seeder     VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`
	}
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("seed_history: create table: %w", err)
	}
	return nil
}

// AppliedSeeders returns all rows in seed_history ordered by id ascending.
func AppliedSeeders(db *sql.DB) ([]HistoryEntry, error) {
	rows, err := db.Query(
		`SELECT id, seeder, applied_at FROM seed_history ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("seed_history: query: %w", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		if err := rows.Scan(&e.ID, &e.Seeder, &e.AppliedAt); err != nil {
			return nil, fmt.Errorf("seed_history: scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// RecordSeeder inserts a seed_history entry for an applied seeder.
func RecordSeeder(db *sql.DB, driverName, seeder string) error {
	q := `INSERT INTO seed_history (seeder) VALUES ($1)`
	if driverName == "mysql" {
		q = `INSERT INTO seed_history (seeder) VALUES (?)`
	}
	if _, err := db.Exec(q, seeder); err != nil {
		return fmt.Errorf("seed_history: recording %q: %w", seeder, err)
	}
	return nil
}

// ClearSeedHistory truncates the seed_history table, used by --fresh.
func ClearSeedHistory(db *sql.DB) error {
	// TRUNCATE is safe here — seed_history has no foreign key dependents.
	if _, err := db.Exec(`TRUNCATE TABLE seed_history`); err != nil {
		// Fallback for engines where TRUNCATE requires special privileges.
		if _, err2 := db.Exec(`DELETE FROM seed_history`); err2 != nil {
			return fmt.Errorf("seed_history: clearing: %w", err)
		}
	}
	return nil
}
