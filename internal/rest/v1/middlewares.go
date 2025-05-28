package v1

import (
	"time"

	prom "github.com/ASsssker/proxy/internal/monitoring/prometheus"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(services.RequestIDKey, requestID)

		c.Next()
	}
}

func RequestDurationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			prom.ProxyHttpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
		}()

		c.Next()
	}
}
