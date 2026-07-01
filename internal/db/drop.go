package db

import (
	"database/sql"
	"fmt"
)

// DropAllTables drops every user-created table in the database in a single
// operation, respecting foreign key dependencies. Works for both Postgres
// and MySQL.
//
// Postgres: queries the information_schema, drops with CASCADE.
// MySQL:    disables FK checks, drops all tables, re-enables FK checks.
func DropAllTables(db *sql.DB, driverName string) error {
	switch driverName {
	case "mysql":
		return dropAllMySQL(db)
	default:
		return dropAllPostgres(db)
	}
}

func dropAllPostgres(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT tablename
		FROM pg_catalog.pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`)
	if err != nil {
		return fmt.Errorf("db:drop: listing tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return fmt.Errorf("db:drop: scanning table name: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(tables) == 0 {
		return nil
	}

	// Build a single DROP TABLE ... CASCADE statement for all tables.
	// Using CASCADE avoids ordering problems with foreign keys.
	quoted := make([]string, len(tables))
	for i, t := range tables {
		quoted[i] = fmt.Sprintf("%q", t)
	}

	stmt := "DROP TABLE IF EXISTS "
	for i, q := range quoted {
		if i > 0 {
			stmt += ", "
		}
		stmt += q
	}
	stmt += " CASCADE"

	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("db:drop: dropping tables: %w", err)
	}
	return nil
}

func dropAllMySQL(db *sql.DB) error {
	// Disable FK checks so we can drop in any order.
	if _, err := db.Exec(`SET FOREIGN_KEY_CHECKS = 0`); err != nil {
		return fmt.Errorf("db:drop: disabling FK checks: %w", err)
	}
	defer db.Exec(`SET FOREIGN_KEY_CHECKS = 1`)

	rows, err := db.Query(`SHOW TABLES`)
	if err != nil {
		return fmt.Errorf("db:drop: listing tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return fmt.Errorf("db:drop: scanning table name: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", t)); err != nil {
			return fmt.Errorf("db:drop: dropping table %q: %w", t, err)
		}
	}

	return nil
}

// TableNames returns all user table names in the database, used by db:tables.
func TableNames(db *sql.DB, driverName string) ([]string, error) {
	var (
		rows *sql.Rows
		err  error
	)
	switch driverName {
	case "mysql":
		rows, err = db.Query(`SHOW TABLES`)
	default:
		rows, err = db.Query(`
			SELECT tablename FROM pg_catalog.pg_tables
			WHERE schemaname = 'public'
			ORDER BY tablename
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("db: listing tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}
