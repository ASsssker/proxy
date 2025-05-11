package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/models"
)

type RequestExecutor struct {
	log        *slog.Logger
	client     *http.Client
	retryCount uint
}

func NewRequestExecutor(cfg config.Config, log *slog.Logger) *RequestExecutor {
	retryCount := cfg.RequesterRetryCount
	if retryCount < 1 {
		retryCount = 1
	}

	return &RequestExecutor{
		client: &http.Client{
			Timeout: cfg.RequesterHTTPClientTimeout,
		},
		retryCount: retryCount,
		log:        log,
	}
}

func (r RequestExecutor) Execute(ctx context.Context, task models.Task) (models.TaskResult, error) {
	const op = "task_executor_service.Execute"
	log := r.log.With(slog.String("op", op), slog.String("task_id", task.ID))
	log.DebugContext(ctx, "start operation")

	bodyReader := strings.NewReader(task.Body)
	request, err := http.NewRequest(task.Method, task.URL, bodyReader)
	if err != nil {
		return models.TaskResult{}, fmt.Errorf("%s task_id=%s failed to create new request: %v", op, task.ID, err)
	}

	request.Header = task.TaskHeadersToHTTPHeaders().Clone()

	for range r.retryCount {
		var resp *http.Response
		resp, err = r.client.Do(request)
		if err != nil {
			log.ErrorContext(ctx, "failed to send request: %v", slog.String("error", err.Error()))
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return models.TaskResult{}, fmt.Errorf("%s task_id=%s failed to read response body: %v", op, task.ID, err)
		}

		headers := make(map[string]string, len(resp.Header))
		for key, values := range resp.Header {
			headers[key] = strings.Join(values, ";")
		}

		taskResult := models.TaskResult{
			ID:            task.ID,
			Status:        models.StatusDone,
			StatusCode:    resp.StatusCode,
			Headers:       headers,
			Body:          string(body),
			ContentLength: int(resp.ContentLength),
		}

		log.DebugContext(ctx, "the operation was successfully completed")

		return taskResult, nil
	}

	return models.TaskResult{}, fmt.Errorf("%s task_id=%sfailed to execute task: %v", op, task.ID, err)
}
