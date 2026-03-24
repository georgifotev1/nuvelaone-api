// tasks/cleanup.go
package tasks

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type CleanupExpiredTokensHandler struct {
	tokenRepo repository.TokenRepository
	logger    *zap.SugaredLogger
}

func NewCleanupExpiredTokensHandler(tokenRepo repository.TokenRepository, logger *zap.SugaredLogger) *CleanupExpiredTokensHandler {
	return &CleanupExpiredTokensHandler{tokenRepo: tokenRepo, logger: logger}
}

func (h *CleanupExpiredTokensHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if err := h.tokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("CleanupExpiredTokensHandler: %w", err)
	}
	h.logger.Infow("cleaned up expired tokens")
	return nil
}
