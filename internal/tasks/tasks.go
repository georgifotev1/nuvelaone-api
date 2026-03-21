package tasks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskMaxRetry = 5
	TaskTimeout  = 30 * time.Second

	TypeWelcomeEmail = "email:welcome"
)

type WelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewWelcomeEmailTask(p WelcomeEmailPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("tasks.NewWelcomeEmailTask: %w", err)
	}
	return asynq.NewTask(TypeWelcomeEmail, payload,
		asynq.MaxRetry(TaskMaxRetry),
		asynq.Timeout(TaskTimeout),
	), nil
}
