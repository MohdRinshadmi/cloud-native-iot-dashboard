package ws

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
)

// Handler upgrades authenticated HTTP requests to WebSocket connections.
//
// Auth: browsers cannot set an Authorization header on the WS handshake, so
// the access token arrives as `?token=` (standard practice; mitigated by
// short-lived tokens). The token is verified BEFORE the upgrade — an invalid
// token never gets a socket.
type Handler struct {
	hub      *Hub
	verifier appauth.TokenVerifier
	upgrader websocket.Upgrader
	log      *slog.Logger
}

// NewHandler wires the WS endpoint. allowedOrigins mirrors the CORS config —
// the browser Origin must match (anti cross-site WebSocket hijacking); empty
// Origin (non-browser clients) is allowed since the token gates access anyway.
func NewHandler(hub *Hub, verifier appauth.TokenVerifier, allowedOrigins []string, log *slog.Logger) *Handler {
	return &Handler{
		hub:      hub,
		verifier: verifier,
		log:      log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "" || slices.Contains(allowedOrigins, origin) || slices.Contains(allowedOrigins, "*")
			},
		},
	}
}

// Serve handles GET /api/v1/ws?token=<access-token>.
func (h *Handler) Serve(c *gin.Context) {
	principal, err := h.verifier.VerifyAccess(c.Query("token"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid or missing token"},
		})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already wrote the HTTP error response.
		h.log.Warn("ws upgrade failed", slog.String("error", err.Error()))
		return
	}

	client := newClient(h.hub, conn, principal.TenantID, principal.UserID)
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}
