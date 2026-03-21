package tasks

import "github.com/hibiken/asynq"

func RegisterTasks(deps *TaskHandlerDeps) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeWelcomeEmail, deps.HandleWelcomeEmail)
	return mux
}
