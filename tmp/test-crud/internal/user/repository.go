package user

import (
	"database/sql"
	"errors"
)

// Repository handles user database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// RepositoryInterface defines the contract for user storage.
type RepositoryInterface interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id int64) (*User, error)
}

// Compile-time check that *Repository satisfies RepositoryInterface.
var _ RepositoryInterface = (*Repository)(nil)

// Create inserts a new user into the database.
func (r *Repository) Create(u *User) error {
	return r.db.QueryRow(
		`INSERT INTO users (name, email, password_hash, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, true, NOW(), NOW())
		 RETURNING id, is_active, created_at, updated_at`,
		u.Name, u.Email, u.PasswordHash,
	).Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
}

// FindByEmail looks up a user by email.
func (r *Repository) FindByEmail(email string) (*User, error) {
	return r.find(`SELECT id, name, email, password_hash, is_active, created_at, updated_at FROM users WHERE email = $1 LIMIT 1`, email)
}

// FindByID looks up a user by primary key.
func (r *Repository) FindByID(id int64) (*User, error) {
	return r.find(`SELECT id, name, email, password_hash, is_active, created_at, updated_at FROM users WHERE id = $1 LIMIT 1`, id)
}

func (r *Repository) find(query string, arg any) (*User, error) {
	var u User
	err := r.db.QueryRow(query, arg).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
