package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskHandlerDeps struct {
	Mailer mailer.Mailer
	Logger *zap.SugaredLogger
}

func (d *TaskHandlerDeps) HandleWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("handleWelcomeEmail: %w", err)
	}

	if d.Mailer == nil {
		d.Logger.Warnw("mailer not configured, skipping welcome email", "user_id", p.UserID)
		return nil
	}

	d.Logger.Infow("sending welcome email", "user_id", p.UserID)

	if err := d.Mailer.Send(mailer.EmailData{
		To:       []string{p.Email},
		Subject:  "Welcome to NuvelaOne",
		Template: "welcome",
		Data:     map[string]any{"name": p.Name},
	}); err != nil {
		return fmt.Errorf("handleWelcomeEmail: %w", err)
	}

	d.Logger.Infow("welcome email sent", "user_id", p.UserID)
	return nil
}
