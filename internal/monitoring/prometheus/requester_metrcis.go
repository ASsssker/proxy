package prom

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequesterTaskExecuteDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "requester_task_execute_duration",
			Help:    "Duration if execute task in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
		},
		[]string{"task_result_status"},
	)

	RequesterTasksStarted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requester_tasks_started",
			Help: "Total number of tasks successfully started",
		},
	)

	RequesterTasksRejected = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requester_tasks_rejected",
			Help: "Total number of tasks rejected due to fill worker pool",
		},
	)
)

func MustRegisterRequesterMetrics(handler *gin.Engine) {
	prometheus.MustRegister(RequesterTaskExecuteDuration, RequesterTasksStarted, RequesterTasksRejected)

	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
