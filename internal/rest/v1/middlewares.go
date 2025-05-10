package v1

import (
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
