package domain

import "time"

// User swagger:model
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	TenantID  string    `json:"tenant_id"`
	Avatar    string    `json:"avatar"`
	Role      Role      `json:"role"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequest swagger:model
type UpdateUserRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
