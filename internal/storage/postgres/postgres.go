package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/ASsssker/proxy/internal/storage"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db  *sql.DB
	log *slog.Logger
}

func NewPostgresDB(ctx context.Context, log *slog.Logger, dns string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %v", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %v", err)
	}

	return &PostgresDB{db: db, log: log}, nil
}

func (p PostgresDB) GetTask(ctx context.Context, taskID string) (models.TaskResult, error) {
	const op = "postgres.GetTask"
	requestID := ctx.Value(services.RequestIDKey).(string)

	log := p.log.With(slog.String("op", op), slog.String(services.RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	stmt := `SELECT id, status, status_code, headers, body, content_length FROM tasks
			WHERE id = $1`

	taskResult := models.TaskResult{Headers: map[string]string{}}
	var status string
	if err := p.db.QueryRowContext(ctx, stmt, taskID).Scan(
		&taskResult.ID,
		&status,
		&taskResult.StatusCode,
		&taskResult.Headers,
		&taskResult.Body,
		&taskResult.ContentLength,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.TaskResult{}, fmt.Errorf("%s request_id=%s task not found: %w: %v",
				op, requestID, storage.ErrTaskNotFound, err)
		}

		return models.TaskResult{}, fmt.Errorf("%s request_id=%s task not found: %v",
			op, requestID, err)
	}

	taskResult.Status = models.TaskStatus(status)

	log.DebugContext(ctx, "the operation was successfully completed")

	return taskResult, nil
}

func (p PostgresDB) AddTask(ctx context.Context, taskID string) error {
	const op = "postgres.AddTask"
	requestID := ctx.Value(services.RequestIDKey).(string)

	log := p.log.With(slog.String("op", op), slog.String(services.RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	stmt := `INSERT INTO tasks (id, status)
			VALUES($1, $2)`

	if _, err := p.db.ExecContext(ctx, stmt, taskID, models.StatusNew); err != nil {
		return fmt.Errorf("%s request_id=%s failed to add new task: %v",
			op, requestID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil
}

func (p PostgresDB) UpdateTaskStatus(ctx context.Context, taskID string, newStatus models.TaskStatus) error {
	const op = "postgres.UpdateTaskStatus"

	log := p.log.With(slog.String("op", op), slog.String("task_id", taskID))
	log.DebugContext(ctx, "start operation")

	stmt := `UPDATE tasks
			SET status = $1
			WHERE id = $2`

	if _, err := p.db.ExecContext(ctx, stmt, string(newStatus), taskID); err != nil {
		return fmt.Errorf("%s task_id=%s failed to update task status: %v", op, taskID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil
}

func (p PostgresDB) UpdateTaskResult(ctx context.Context, taskResult models.TaskResult) error {
	const op = "postgres.UpdateTaskResult"

	log := p.log.With(slog.String("op", op), slog.String("task_id", taskResult.ID))
	log.DebugContext(ctx, "start operation")

	stmt := `UPDATE tasks
			SET status = $1,
				status_code = $2,
				headers = $3,
				body = $4,
				content_length = $5
			WHERE id = $6`

	if _, err := p.db.ExecContext(ctx, stmt,
		models.StatusDone,
		taskResult.StatusCode,
		taskResult.Headers,
		taskResult.Body,
		taskResult.ContentLength,
		taskResult.ID,
	); err != nil {
		return fmt.Errorf("%s task_id=%s failed to update task status: %v", op, taskResult.ID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil
}

func (p PostgresDB) Close(_ context.Context) error {
	return p.db.Close()
}
