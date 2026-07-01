package driver

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver with database/sql
)

// MySQLDriver implements Driver for MySQL / MariaDB.
//
// IMPORTANT: MySQL does not support transactional DDL. Statements such as
// CREATE TABLE and ALTER TABLE cause an implicit commit, so a transaction
// cannot be used to atomically roll back a partially-applied migration file.
// The migration engine checks SupportsTransactionalDDL() and emits a clear
// warning when a MySQL migration fails mid-file, since some schema changes
// may have been committed already.
type MySQLDriver struct{}

// DriverName returns the database/sql registration name for go-sql-driver/mysql.
func (MySQLDriver) DriverName() string { return "mysql" }

// SupportsTransactionalDDL reports false: MySQL cannot roll back DDL.
func (MySQLDriver) SupportsTransactionalDDL() bool { return false }

// CreateDatabaseIfMissing connects via adminDB (no database selected)
// and creates dbname using IF NOT EXISTS so the call is idempotent.
func (MySQLDriver) CreateDatabaseIfMissing(adminDB *sql.DB, dbname string) error {
	// MySQL supports CREATE DATABASE IF NOT EXISTS natively and idempotently.
	// Character set and collation are set to sensible modern defaults.
	_, err := adminDB.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		dbname,
	))
	if err != nil {
		return fmt.Errorf("mysql: creating database %q: %w", dbname, err)
	}
	return nil
}
