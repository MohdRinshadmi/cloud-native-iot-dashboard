package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery converts a panic into a clean 500 JSON response (never leaking a
// stack trace to the client) and logs the failure with the correlation id.
// This keeps a single bad request from taking down the worker.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					slog.Any("panic", r),
					slog.String("path", c.Request.URL.Path),
					slog.String("request_id", GetRequestID(c)),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "INTERNAL",
						"message":    "internal server error",
						"request_id": GetRequestID(c),
					},
				})
			}
		}()
		c.Next()
	}
}
