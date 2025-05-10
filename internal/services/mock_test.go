package services

import (
	"context"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/stretchr/testify/mock"
)

type mockMessageBroker struct {
	mock.Mock
}

func (mb *mockMessageBroker) Send(ctx context.Context, task models.Task) error {
	args := mb.Called(ctx.Value(RequestIDKey))
	return args.Error(0)
}

type mockTaskProvider struct {
	mock.Mock
}

func (tp *mockTaskProvider) AddTask(ctx context.Context, taskResult models.TaskResult) error {
	args := tp.Called(ctx.Value(RequestIDKey))
	return args.Error(0)
}

func (tp *mockTaskProvider) GetTask(ctx context.Context, taskID string) (models.TaskResult, error) {
	args := tp.Called(ctx.Value(RequestIDKey))

	return args.Get(0).(models.TaskResult), args.Error(1)
}
