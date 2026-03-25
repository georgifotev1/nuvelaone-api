package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/mock"
)

type MockTaskEnqueuer struct {
	mock.Mock
}

func (m *MockTaskEnqueuer) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}
