// Package router assembles the Gin engine: global middleware, health probes,
// and the versioned /api/v1 surface. This is the single composition point for
// HTTP routing — every route, guard and limit is visible in one file.
package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/handler"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
)

// Deps is the explicit dependency set the router needs.
type Deps struct {
	Config   *config.Config
	Logger   *slog.Logger
	Health   *handler.HealthHandler
	Auth     *handler.AuthHandler
	Devices  *handler.DeviceHandler
	Verifier appauth.TokenVerifier
	Limiter  middleware.Limiter
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

	authRequired := middleware.AuthRequired(d.Verifier)
	manageRoles := middleware.RequireRoles(user.RoleAdmin, user.RoleOperator)
	adminOnly := middleware.RequireRoles(user.RoleAdmin)
	// Brute-force guard on credential endpoints: 10 attempts/min per IP.
	loginLimit := middleware.RateLimit(d.Limiter, 10, time.Minute, d.Logger)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", d.Health.Ready)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", loginLimit, d.Auth.Register)
			auth.POST("/login", loginLimit, d.Auth.Login)
			auth.POST("/refresh", d.Auth.Refresh)
			auth.POST("/logout", d.Auth.Logout)
			auth.GET("/me", authRequired, d.Auth.Me)
		}

		devices := v1.Group("/devices", authRequired)
		{
			devices.GET("", d.Devices.List)
			devices.POST("", manageRoles, d.Devices.Create)
			devices.GET("/:id", d.Devices.Get)
			devices.PATCH("/:id", manageRoles, d.Devices.Update)
			devices.DELETE("/:id", adminOnly, d.Devices.Delete)
		}
	}

	return r
}
