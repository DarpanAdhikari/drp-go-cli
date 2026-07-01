// Package driver defines the Driver interface that every database driver
// must satisfy, and provides a factory for resolving the correct driver
// from a config string.
package driver

import (
	"database/sql"
	"fmt"
)

// Driver encapsulates behaviour that differs between database engines:
// DSN building, admin-level DB creation, and DDL transaction support.
type Driver interface {
	// SupportsTransactionalDDL reports whether this engine can safely wrap
	// DDL statements (CREATE TABLE, ALTER TABLE, etc.) in a transaction
	// that rolls back on failure. Postgres: yes. MySQL: no.
	SupportsTransactionalDDL() bool

	// CreateDatabaseIfMissing connects to the server's admin/default
	// database and issues CREATE DATABASE if the target db is absent.
	CreateDatabaseIfMissing(adminDB *sql.DB, dbname string) error

	// DriverName returns the database/sql driver registration name
	// (e.g. "postgres", "mysql") for sql.Open.
	DriverName() string
}

// New returns the Driver implementation for the given driver name string.
func New(name string) (Driver, error) {
	switch name {
	case "postgres":
		return &PostgresDriver{}, nil
	case "mysql":
		return &MySQLDriver{}, nil
	default:
		return nil, fmt.Errorf("driver: unsupported driver %q", name)
	}
}
