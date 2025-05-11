package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/storage"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TaskProvider interface {
	AddTask(ctx context.Context, taskID string) error
	GetTask(ctx context.Context, taskID string) (models.TaskResult, error)
	Close(ctx context.Context) error
}

type MessageSender interface {
	SendTask(ctx context.Context, task models.Task) error
	Close(ctx context.Context) error
}

type ProxyService struct {
	log          *slog.Logger
	taskProvider TaskProvider
	msgSender    MessageSender
	validator    *validator.Validate
}

func NewProxyService(log *slog.Logger, taskProvider TaskProvider, msgSender MessageSender,
	validator *validator.Validate) *ProxyService {
	return &ProxyService{
		log:          log,
		taskProvider: taskProvider,
		msgSender:    msgSender,
		validator:    validator,
	}
}

func (p ProxyService) AddTask(ctx context.Context, newTask models.NewTask) (string, error) {
	const op = "proxy_service.AddTask"
	requestID := ctx.Value(RequestIDKey).(string)

	log := p.log.With(slog.String("op", op), slog.String(RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	if err := p.validator.Struct(newTask); err != nil {
		return "", fmt.Errorf("%s request_id=%s failed to validate task: %w: %w", op, requestID, ErrValidation, err)
	}

	taskID := uuid.NewString()

	if err := p.taskProvider.AddTask(ctx, taskID); err != nil {
		return "", fmt.Errorf("%s request_id=%s failed to add task: %w", op, requestID, err)
	}

	task := models.Task{
		ID:      taskID,
		URL:     newTask.URL,
		Method:  newTask.Method,
		Headers: newTask.Headers,
		Body:    newTask.Body,
	}
	if err := p.msgSender.SendTask(ctx, task); err != nil {
		return "", fmt.Errorf("%s request_id=%s failed to send task %s: %w", op, requestID, taskID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return taskID, nil
}

func (p ProxyService) GetTaskInfo(ctx context.Context, taskID string) (models.TaskResult, error) {
	const op = "proxy_service.GetTaskInfo"
	requestID := ctx.Value(RequestIDKey).(string)

	log := p.log.With(slog.String("op", op), slog.String(RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	taskInfo, err := p.taskProvider.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			return models.TaskResult{}, fmt.Errorf("%s request_id=%s task not found: %w",
				op, requestID, ErrTaskNotFound)
		}

		return models.TaskResult{}, fmt.Errorf("%s request_id=%s failed to get task: %w", op, requestID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return taskInfo, nil
}

func (p ProxyService) Close(ctx context.Context) error {
	errCloseTaskProvider := p.taskProvider.Close(ctx)
	if errCloseTaskProvider != nil {
		errCloseTaskProvider = fmt.Errorf("failed to close task provider: %v", errCloseTaskProvider)

		if err := p.msgSender.Close(ctx); err != nil {
			return errors.Join(fmt.Errorf("failed to close message sender: %v", err), errCloseTaskProvider)
		}

		return errCloseTaskProvider
	}

	if err := p.msgSender.Close(ctx); err != nil {
		return fmt.Errorf("failed to close message sender: %v", err)
	}

	return nil
}
