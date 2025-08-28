package handlers

import (
	"MTConnect/internal/middleware/logging"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware(parentLogger *logging.Logger) gin.HandlerFunc {
	logger := parentLogger.WithPrefix("HTTP")

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		start := time.Now()
		logger.Info("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.Request.RemoteAddr,
		)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		logger.Info("Request completed",
			"status", status,
			"latency", latency,
			"client_ip", c.ClientIP(),
		)
	}
}
