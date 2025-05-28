package prom

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ProxyPingCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "proxy_ping_request_count",
			Help: "No of request handled by Ping handler",
		},
	)

	ProxyHttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)
)

func MustRegisterProxyMetrics(handler *gin.Engine) {
	prometheus.MustRegister(ProxyPingCounter, ProxyHttpRequestDuration)

	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
