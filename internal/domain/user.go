package domain

import "time"

// User represents the core user entity.
// Add or rename fields to match your domain.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
}
