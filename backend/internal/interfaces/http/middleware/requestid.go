package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID is the canonical correlation header.
const HeaderRequestID = "X-Request-ID"

// ctxKeyRequestID is the Gin context key for the request id.
const ctxKeyRequestID = "request_id"

// RequestID ensures every request carries a correlation id. It honours an
// inbound X-Request-ID (so ids propagate across services) or mints a new one,
// then echoes it back on the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = newID()
		}
		c.Set(ctxKeyRequestID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}

// GetRequestID extracts the correlation id from the Gin context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b)
}
