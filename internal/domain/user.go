package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	TenantID  string    `json:"tenant_id"`
	Avatar    string    `json:"avatar"`
	Role      string    `json:"role"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role" validate:"required"`
}

type UpdateUserRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
