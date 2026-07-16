package {{.DomainName}}_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"{{.ModuleName}}/internal/{{.DomainName}}"
)

func setup{{.Name}}TestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE {{.TableName}} (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`)
	require.NoError(t, err)
	return db
}

func insertTest{{.Name}}(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	now := time.Now()
	var id int64
	err := db.QueryRow(
		`INSERT INTO {{.TableName}} (created_at, updated_at) VALUES (?, ?) RETURNING id`,
		now, now,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func Test{{.Name}}Repository_FindAll(t *testing.T) {
	db := setup{{.Name}}TestDB(t)
	repo := {{.DomainName}}.New{{.Name}}Repository(db)

	insertTest{{.Name}}(t, db)
	insertTest{{.Name}}(t, db)

	items, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func Test{{.Name}}Repository_FindByID(t *testing.T) {
	db := setup{{.Name}}TestDB(t)
	repo := {{.DomainName}}.New{{.Name}}Repository(db)

	id := insertTest{{.Name}}(t, db)

	item, err := repo.FindByID(id)
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, id, item.ID)
}

func Test{{.Name}}Repository_Create(t *testing.T) {
	db := setup{{.Name}}TestDB(t)
	repo := {{.DomainName}}.New{{.Name}}Repository(db)

	m := &{{.DomainName}}.{{.Name}}{}
	created, err := repo.Create(m)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotZero(t, created.ID)
}

func Test{{.Name}}Repository_Update(t *testing.T) {
	db := setup{{.Name}}TestDB(t)
	repo := {{.DomainName}}.New{{.Name}}Repository(db)

	id := insertTest{{.Name}}(t, db)

	original, err := repo.FindByID(id)
	require.NoError(t, err)
	require.NotNil(t, original)

	updated, err := repo.Update(original)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.True(t, updated.UpdatedAt.Equal(original.UpdatedAt) || updated.UpdatedAt.After(original.UpdatedAt))
}

func Test{{.Name}}Repository_Delete(t *testing.T) {
	db := setup{{.Name}}TestDB(t)
	repo := {{.DomainName}}.New{{.Name}}Repository(db)

	id := insertTest{{.Name}}(t, db)

	err := repo.Delete(id)
	require.NoError(t, err)

	item, err := repo.FindByID(id)
	require.NoError(t, err)
	require.Nil(t, item)
}
