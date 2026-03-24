package tasks

import (
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type HandlerDeps struct {
	Mailer    mailer.Mailer
	TokenRepo repository.TokenRepository
	Logger    *zap.SugaredLogger
}

func Register(deps HandlerDeps) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.Handle(TypeWelcomeEmail,
		NewWelcomeEmailHandler(deps.Mailer, deps.Logger))

	mux.Handle(TypeCleanupExpiredTokens,
		NewCleanupExpiredTokensHandler(deps.TokenRepo, deps.Logger))

	return mux
}
