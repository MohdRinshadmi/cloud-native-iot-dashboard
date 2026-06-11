package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Limiter is the port the middleware needs (implemented by cache.RateLimiter).
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RateLimit guards a route with `limit` requests per `window`, keyed by
// client IP + route. FAIL-OPEN by design: if Redis is unreachable we log and
// admit the request — availability of login beats strictness of the limiter.
func RateLimit(l Limiter, limit int, window time.Duration, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), c.ClientIP())

		allowed, err := l.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			log.Warn("rate limiter unavailable, failing open",
				slog.String("error", err.Error()), slog.String("request_id", GetRequestID(c)))
			c.Next()
			return
		}
		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":       "RATE_LIMITED",
					"message":    "too many requests, slow down",
					"request_id": GetRequestID(c),
				},
			})
			return
		}
		c.Next()
	}
}
