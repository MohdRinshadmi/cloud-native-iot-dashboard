package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

const ctxKeyPrincipal = "principal"

// AuthRequired validates the Bearer access token and attaches the principal
// to the request context. 401 on anything malformed, expired or unsigned.
func AuthRequired(verifier appauth.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			unauthorized(c, "missing bearer token")
			return
		}

		principal, err := verifier.VerifyAccess(token)
		if err != nil {
			unauthorized(c, "invalid or expired token")
			return
		}

		c.Set(ctxKeyPrincipal, principal)
		c.Next()
	}
}

// RequireRoles gates a route to the given roles. Must run after AuthRequired.
func RequireRoles(roles ...user.Role) gin.HandlerFunc {
	allowed := make(map[user.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p == nil {
			unauthorized(c, "missing bearer token")
			return
		}
		if _, ok := allowed[p.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":       "FORBIDDEN",
					"message":    "insufficient permissions",
					"request_id": GetRequestID(c),
				},
			})
			return
		}
		c.Next()
	}
}

// GetPrincipal returns the authenticated principal, or nil when unauthenticated.
func GetPrincipal(c *gin.Context) *appauth.Principal {
	if v, ok := c.Get(ctxKeyPrincipal); ok {
		if p, ok := v.(*appauth.Principal); ok {
			return p
		}
	}
	return nil
}

func unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":       "UNAUTHORIZED",
			"message":    msg,
			"request_id": GetRequestID(c),
		},
	})
}
