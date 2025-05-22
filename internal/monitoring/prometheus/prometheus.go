package prom

import "github.com/prometheus/client_golang/prometheus"

var PingCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ping_request_count",
		Help: "No of request handled by Ping handler",
	},
)
