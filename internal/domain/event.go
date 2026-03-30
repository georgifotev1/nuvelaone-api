package domain

import "time"

type Event struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ServiceID  string    `json:"service_id"`
	UserID     string    `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EventRequest struct {
	CustomerID string    `json:"customer_id" validate:"required"`
	ServiceID  string    `json:"service_id" validate:"required"`
	UserID     string    `json:"user_id" validate:"required"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
	Notes      string    `json:"notes"`
}

type EventUpdateRequest struct {
	CustomerID string    `json:"customer_id"`
	ServiceID  string    `json:"service_id"`
	UserID     string    `json:"user_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
}

type EventListFilter struct {
	StartDate string
	EndDate   string
}
