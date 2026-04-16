package domain

import "time"

type Service struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
	Buffer      int       `json:"buffer"`
	Cost        int       `json:"cost"`
	Visible     bool      `json:"visible"`
	TenantID    string    `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ProviderIDs []string  `json:"provider_ids"`
}

type ServiceRequest struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description"`
	Duration    int      `json:"duration" validate:"required"`
	Buffer      int      `json:"buffer"`
	Cost        int      `json:"cost" validate:"required"`
	Visible     bool     `json:"visible"`
	UserIDs     []string `json:"user_ids"`
}

type TimeslotRequest struct {
	ServiceID string `json:"service_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
	Date      string `json:"date" validate:"required"`
}

type TimeslotResponse struct {
	Timeslots []string `json:"timeslots"`
}
