package user

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Service contains user business logic.
type Service struct {
	repo RepositoryInterface
}

// NewService constructs a new Service.
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// ServiceInterface defines the contract for user business logic.
type ServiceInterface interface {
	Register(req RegisterRequest) (*User, error)
	Authenticate(email, password string) (*User, error)
	FindByID(id int64) (*User, error)
}

// Compile-time check that *Service satisfies ServiceInterface.
var _ ServiceInterface = (*Service)(nil)

// Register creates a new user after validating input.
func (s *Service) Register(req RegisterRequest) (*User, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Name == "" {
		return nil, fmt.Errorf("name and email are required")
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{Name: req.Name, Email: req.Email, PasswordHash: string(hash)}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Authenticate verifies credentials and returns the user.
func (s *Service) Authenticate(email, password string) (*User, error) {
	u, err := s.repo.FindByEmail(strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, fmt.Errorf("user is inactive")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return u, nil
}

// FindByID returns a user by primary key.
func (s *Service) FindByID(id int64) (*User, error) {
	return s.repo.FindByID(id)
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain an uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain a lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain a digit")
	}
	return nil
}
