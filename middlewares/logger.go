package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger records method, path, status code, latency, and client IP for every
// request. It runs c.Next() first so the log line is written after the handler
// completes and the status code is known.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}
