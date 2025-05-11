package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/models"
	"github.com/alitto/pond/v2"
)

type TaskUpdater interface {
	UpdateTaskStatus(ctx context.Context, taskID string, newStatus models.TaskStatus) error
	UpdateTaskResult(ctx context.Context, taskResult models.TaskResult) error
	Close(ctx context.Context) error
}

type MessageReceiver interface {
	Subscribe(ctx context.Context, taskChan chan models.Task) (context.CancelFunc, error)
	Close(ctx context.Context) error
}

type TaskExecutor interface {
	Execute(ctx context.Context, task models.Task) (models.TaskResult, error)
}

type RequesterService struct {
	log          *slog.Logger
	taskUpdater  TaskUpdater
	msgReceiver  MessageReceiver
	taskExecutor TaskExecutor
	pool         pond.Pool
	taskChan     chan models.Task
	cancel       context.CancelFunc
}

func NewRequesterService(log slog.Logger, cfg config.Config, taskUpdater TaskUpdater,
	msgReceiver MessageReceiver, taskExecutor TaskExecutor) *RequesterService {

	return &RequesterService{
		log:          &log,
		taskUpdater:  taskUpdater,
		msgReceiver:  msgReceiver,
		taskExecutor: taskExecutor,
		pool:         pond.NewPool(int(cfg.RequesterWorkersCount), pond.WithNonBlocking(true)),
		taskChan:     make(chan models.Task),
	}
}

func (r RequesterService) Run(ctx context.Context) error {
	cancel, err := r.msgReceiver.Subscribe(ctx, r.taskChan)
	if err != nil {
		return fmt.Errorf("failed to run requester service: %w", err)
	}

	defer cancel()
	r.cancel = cancel
	for task := range r.taskChan {
		err := r.pool.Go(func() {
			r.processTask(task)
		})

		if err != nil {
			r.log.Error("failed to run task", slog.String("task_id", task.ID),
				slog.String("error", err.Error()))
		}
	}

	r.pool.WaitingTasks()

	return nil
}

func (r RequesterService) Close(ctx context.Context) error {
	defer r.pool.StopAndWait()
	r.cancel()

	errCloseTaskUpdater := r.taskUpdater.Close(ctx)
	if errCloseTaskUpdater != nil {
		errCloseTaskUpdater = fmt.Errorf("failed to close task updater: %v", errCloseTaskUpdater)

		if err := r.msgReceiver.Close(ctx); err != nil {
			return errors.Join(fmt.Errorf("failed to close message receiver: %v", err), errCloseTaskUpdater)
		}

		return errCloseTaskUpdater
	}

	if err := r.msgReceiver.Close(ctx); err != nil {
		return fmt.Errorf("failed to close message receiver: %v", err)
	}

	return nil

}

func (r RequesterService) processTask(task models.Task) {
	ctx := context.TODO()

	if err := r.taskUpdater.UpdateTaskStatus(ctx, task.ID, models.StatusInProcess); err != nil {
		r.log.Error("failed to update task status", slog.String("task_id", task.ID),
			slog.String("status", string(models.StatusInProcess)),
			slog.String("error", err.Error()),
		)

		return
	}

	taskResult, err := r.taskExecutor.Execute(context.TODO(), task)
	if err != nil {
		r.log.Error("failed to execute task", slog.String("task_id", task.ID),
			slog.String("error", err.Error()))

		if err := r.taskUpdater.UpdateTaskStatus(ctx, task.ID, models.StatusError); err != nil {
			r.log.Error("failed to update task status", slog.String("task_id", task.ID),
				slog.String("status", string(models.StatusError)),
				slog.String("error", err.Error()),
			)

			return
		}

		return
	}

	if err := r.taskUpdater.UpdateTaskResult(ctx, taskResult); err != nil {
		r.log.Error("failed to update task result", slog.String("task_id", task.ID),
			slog.String("error", err.Error()))

		return
	}
}
