// Package db handles database connectivity: opening connections, creating
// the target database if it doesn't exist, and ensuring the schema_history
// bookkeeping table exists.
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/yourorg/drp/internal/config"
	"github.com/yourorg/drp/internal/db/driver"
)

// Connection bundles an open *sql.DB with its resolved Driver so callers
// carry both in a single value.
type Connection struct {
	DB         *sql.DB
	Driver     driver.Driver
	DriverName string
}

// Connect opens a database connection using cfg, creating the target
// database first if it does not already exist, then ensures the
// schema_history table is present. Returns a ready-to-use Connection.
func Connect(cfg *config.Config) (*Connection, error) {
	drv, err := driver.New(cfg.DBDriver)
	if err != nil {
		return nil, err
	}

	// Step 1: admin connection (no target DB) — used only to create DB if missing.
	adminDB, err := open(drv.DriverName(), cfg.AdminDSN())
	if err != nil {
		return nil, fmt.Errorf(
			"db: cannot reach server at %s:%s as %q: %w\n  → check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD in .env",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, err,
		)
	}
	defer adminDB.Close()

	if err := drv.CreateDatabaseIfMissing(adminDB, cfg.DBName); err != nil {
		return nil, fmt.Errorf("db: ensuring database %q exists: %w", cfg.DBName, err)
	}

	// Step 2: connect to the target database.
	db, err := open(drv.DriverName(), cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db: connecting to database %q: %w", cfg.DBName, err)
	}

	// Step 3: ensure schema_history table exists.
	if err := EnsureSchemaHistoryTable(db, drv.DriverName()); err != nil {
		db.Close()
		return nil, fmt.Errorf("db: initialising schema_history: %w", err)
	}

	return &Connection{DB: db, Driver: drv, DriverName: drv.DriverName()}, nil
}

// Ping verifies the database is reachable. Used by `drp doctor` and `drp db:status`.
func Ping(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("db: ping failed: %w", err)
	}
	return nil
}

// open opens a *sql.DB with sensible connection pool defaults and verifies
// reachability. The returned DB is already confirmed to accept connections.
func open(driverName, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
