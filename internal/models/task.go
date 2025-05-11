package models

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"strings"
)

type TaskStatus string

var (
	StatusDone      = TaskStatus("done")
	StatusInProcess = TaskStatus("in process")
	StatusError     = TaskStatus("error")
	StatusNew       = TaskStatus("new")
)

type TaskResult struct {
	ID            string     `json:"id"`
	Status        TaskStatus `json:"status"`
	StatusCode    int        `json:"http_status_code"`
	Headers       Headers    `json:"headers"`
	Body          string     `json:"body"`
	ContentLength int        `json:"content_length"`
}

type NewTask struct {
	URL     string            `json:"url" validate:"required,http_url"`
	Method  string            `json:"method" validate:"required,httpmethod"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Task struct {
	ID      string            `json:"id"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (t Task) TaskHeadersToHTTPHeaders() http.Header {
	headers := make(http.Header)
	for key, values := range t.Headers {
		for _, value := range strings.Split(values, ";") {
			headers.Add(key, value)
		}
	}

	return headers
}

type Headers map[string]string

func (j Headers) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (h Headers) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	}
	return json.Unmarshal(data, &h)
}
