package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery catches any panic that occurs during request handling, logs the
// stack trace, and returns a 500 JSON response via the response wrapper
// instead of Gin's default plain-text or HTML panic page.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)
				response.Error(c, http.StatusInternalServerError, "An unexpected error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
