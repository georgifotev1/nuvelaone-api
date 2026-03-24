package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type WelcomeEmailHandler struct {
	mailer mailer.Mailer
	logger *zap.SugaredLogger
}

func NewWelcomeEmailHandler(m mailer.Mailer, logger *zap.SugaredLogger) *WelcomeEmailHandler {
	return &WelcomeEmailHandler{mailer: m, logger: logger}
}

func (h *WelcomeEmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("WelcomeEmailHandler: %w", asynq.SkipRetry)
	}

	if h.mailer == nil {
		h.logger.Warnw("mailer not configured, skipping welcome email", "userID", p.UserID)
		return nil
	}

	h.logger.Infow("sending welcome email", "userID", p.UserID)

	if err := h.mailer.Send(mailer.EmailData{
		To:       []string{p.Email},
		Subject:  "Welcome to NuvelaOne",
		Template: "welcome",
		Data:     map[string]any{"name": p.Name},
	}); err != nil {
		return fmt.Errorf("WelcomeEmailHandler send: %w", err)
	}

	h.logger.Infow("welcome email sent", "userID", p.UserID)
	return nil
}
