// Package router assembles the Gin engine: global middleware, health probes,
// and the versioned /api/v1 group that feature handlers attach to in later
// phases. This is the single composition point for HTTP routing.
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/handler"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
)

// Deps is the explicit dependency set the router needs. Passing a struct keeps
// the signature stable as the app grows (no positional-arg churn).
type Deps struct {
	Config *config.Config
	Logger *slog.Logger
	Health *handler.HealthHandler
}

// New builds the fully-wired Gin engine.
func New(d Deps) *gin.Engine {
	if d.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New (not Default) so we control the middleware stack precisely.
	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(d.Logger),
		middleware.Recovery(d.Logger),
		middleware.CORS(d.Config.HTTP.AllowedOrigins),
	)

	// --- operational probes (unversioned, no auth) ---
	r.GET("/livez", d.Health.Live)
	r.GET("/readyz", d.Health.Ready)

	// --- versioned API surface ---
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", d.Health.Ready)
		// Phase 4+ mount auth, devices, telemetry, alerts here.
	}

	return r
}
