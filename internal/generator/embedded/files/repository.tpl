package {{.DomainName}}

import (
	"database/sql"
	"fmt"
	"time"
)

// {{.Name}}Repository handles all database operations for {{.Name}}.
type {{.Name}}Repository struct {
	db *sql.DB
}

// New{{.Name}}Repository constructs a new {{.Name}}Repository.
func New{{.Name}}Repository(db *sql.DB) *{{.Name}}Repository {
	return &{{.Name}}Repository{db: db}
}

// FindAll returns all {{.PluralName}} from the database.
func (r *{{.Name}}Repository) FindAll() ([]{{.Name}}, error) {
	rows, err := r.db.Query(`SELECT id, created_at, updated_at FROM {{.TableName}} ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}}: find all: %w", err)
	}
	defer rows.Close()

	var results []{{.Name}}
	for rows.Next() {
		var m {{.Name}}
		if err := rows.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("{{.TableName}}: scan: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// FindByID returns a single {{.Name}} by primary key.
func (r *{{.Name}}Repository) FindByID(id int64) (*{{.Name}}, error) {
	var m {{.Name}}
	err := r.db.QueryRow(
		`SELECT id, created_at, updated_at FROM {{.TableName}} WHERE id = $1`, id,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}}: find by id %d: %w", id, err)
	}
	return &m, nil
}

// Create inserts a new {{.Name}} and returns the created record.
func (r *{{.Name}}Repository) Create(m *{{.Name}}) (*{{.Name}}, error) {
	now := time.Now()
	err := r.db.QueryRow(
		`INSERT INTO {{.TableName}} (created_at, updated_at) VALUES ($1, $2) RETURNING id, created_at, updated_at`,
		now, now,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}}: create: %w", err)
	}
	return m, nil
}

// Update modifies an existing {{.Name}} record.
func (r *{{.Name}}Repository) Update(m *{{.Name}}) (*{{.Name}}, error) {
	m.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE {{.TableName}} SET updated_at = $1 WHERE id = $2`,
		m.UpdatedAt, m.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}}: update %d: %w", m.ID, err)
	}
	return m, nil
}

// Delete removes a {{.Name}} by primary key.
func (r *{{.Name}}Repository) Delete(id int64) error {
	if _, err := r.db.Exec(`DELETE FROM {{.TableName}} WHERE id = $1`, id); err != nil {
		return fmt.Errorf("{{.TableName}}: delete %d: %w", id, err)
	}
	return nil
}
