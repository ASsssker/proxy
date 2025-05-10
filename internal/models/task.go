package models

type TaskStatus string

var (
	DoneStatus      = TaskStatus("done")
	InProcessStatus = TaskStatus("in process")
	ErrorStatus     = TaskStatus("error")
	NewStatus       = TaskStatus("new")
)

type TaskInfo struct {
	ID            string            `json:"id"`
	Status        TaskStatus        `json:"status"`
	StatusCode    int               `json:"http_status_code"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	ContentLength int               `json:"content_length"`
}

type NewTask struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
