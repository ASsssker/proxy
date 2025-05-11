package services

import (
	"context"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/stretchr/testify/mock"
)

type mockMessageBroker struct {
	mock.Mock
}

func (mb *mockMessageBroker) SendTask(ctx context.Context, _ models.Task) error {
	args := mb.Called(ctx.Value(RequestIDKey))
	return args.Error(0)
}

func (mb *mockMessageBroker) Close(ctx context.Context) error {
	return nil
}

type mockTaskProvider struct {
	mock.Mock
}

func (tp *mockTaskProvider) AddTask(ctx context.Context, _ string) error {
	args := tp.Called(ctx.Value(RequestIDKey))
	return args.Error(0)
}

func (tp *mockTaskProvider) GetTask(ctx context.Context, _ string) (models.TaskResult, error) {
	args := tp.Called(ctx.Value(RequestIDKey))

	return args.Get(0).(models.TaskResult), args.Error(1)
}

func (mb *mockTaskProvider) Close(ctx context.Context) error {
	return nil
}
