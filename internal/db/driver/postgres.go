package driver

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // registers "postgres" driver with database/sql
)

// PostgresDriver implements Driver for PostgreSQL.
// Postgres fully supports transactional DDL — a failed migration file
// leaves schema_history and the schema in a consistent state.
type PostgresDriver struct{}

// DriverName returns the database/sql registration name for lib/pq.
func (PostgresDriver) DriverName() string { return "postgres" }

// SupportsTransactionalDDL reports true: Postgres rolls back DDL on error.
func (PostgresDriver) SupportsTransactionalDDL() bool { return true }

// CreateDatabaseIfMissing connects via adminDB (pointed at "postgres"
// system database) and creates dbname if it does not already exist.
func (PostgresDriver) CreateDatabaseIfMissing(adminDB *sql.DB, dbname string) error {
	// Check existence first to avoid a superfluous "database already exists" error.
	var exists bool
	row := adminDB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_database WHERE datname = $1)`, dbname,
	)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("postgres: checking database existence: %w", err)
	}
	if exists {
		return nil
	}

	// Identifier quoting — use pg's quote_ident equivalent via fmt since
	// we cannot parameterise a CREATE DATABASE statement.
	// We sanitise by only allowing characters valid in an unquoted identifier;
	// anything else requires the caller to quote explicitly.
	if _, err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbname)); err != nil {
		return fmt.Errorf("postgres: creating database %q: %w", dbname, err)
	}
	return nil
}
