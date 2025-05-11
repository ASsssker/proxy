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

	stmt := `INSERT INTO tasks (id)
			VALUES($1, $2)`

	if _, err := p.db.ExecContext(ctx, stmt, taskID, models.StatusNew); err != nil {
		return fmt.Errorf("%s request_id=%s failed to add new task: %v",
			op, requestID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil
}

func (p PostgresDB) Close(_ context.Context) error {
	return p.db.Close()
}
