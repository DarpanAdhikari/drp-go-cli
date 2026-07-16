package user

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	return db
}

func insertTestUser(t *testing.T, db *sql.DB, name, email, passwordHash string) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO users (name, email, password_hash, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		name, email, passwordHash,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

func TestRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id := insertTestUser(t, db, "Alice", "alice@example.com", "hash")
	require.True(t, id > 0)

	u, err := repo.FindByEmail("alice@example.com")
	require.NoError(t, err)
	require.Equal(t, "Alice", u.Name)
	require.Equal(t, "alice@example.com", u.Email)
}

func TestRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id := insertTestUser(t, db, "Bob", "bob@example.com", "hash")

	u, err := repo.FindByID(id)
	require.NoError(t, err)
	require.Equal(t, "Bob", u.Name)
	require.Equal(t, "bob@example.com", u.Email)
}
