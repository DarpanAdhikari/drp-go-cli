package user

import (
	"/tmp/test-crud/internal/shared"
)

// User represents a user record in the database.
type User struct {
	shared.Base
	Name         string    `json:"name" validate:"required,min=2,max=120"`
	Email        string    `json:"email" validate:"required,email"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"is_active"`
}

// UserResponse is the public-safe representation of a user.
type UserResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, IsActive: u.IsActive}
}

// ── Request DTOs ────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=120"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	MacAddress string `json:"mac_address,omitempty"`
	FCMToken   string `json:"fcm_token,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	MacAddress   string `json:"mac_address,omitempty"`
	FCMToken     string `json:"fcm_token,omitempty"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
