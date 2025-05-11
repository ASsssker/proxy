package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/models"
)

type RequestExecutor struct {
	client     *http.Client
	retryCount uint
}

func NewRequestExecutor(cfg config.Config) *RequestExecutor {
	return &RequestExecutor{
		client: &http.Client{
			Timeout: cfg.RequesterHTTPClientTimeout,
		},
		retryCount: cfg.RequesterRetryCount,
	}
}

func (r RequestExecutor) Execute(ctx context.Context, task models.Task) (models.TaskResult, error) {
	bodyReader := strings.NewReader(task.Body)
	request, err := http.NewRequest(task.Method, task.URL, bodyReader)
	if err != nil {
		return models.TaskResult{}, fmt.Errorf("task_id=%s failed to create new request: %v", task.ID, err)
	}

	request.Header = task.TaskHeadersToHTTPHeaders().Clone()

	for range r.retryCount {
		var resp *http.Response
		resp, err = r.client.Do(request)
		if err != nil {
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return models.TaskResult{}, fmt.Errorf("task_id=%s failed to read response body: %v", task.ID, err)
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

		return taskResult, nil
	}

	return models.TaskResult{}, err
}
