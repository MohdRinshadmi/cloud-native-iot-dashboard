package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/application/health"
)

// HealthHandler exposes liveness and readiness probes. Kubernetes hits these:
//   - /livez   -> restart the pod if this fails
//   - /readyz  -> remove the pod from the Service if this fails
type HealthHandler struct {
	svc *health.Service
}

// NewHealthHandler wires the handler to the application health service.
func NewHealthHandler(svc *health.Service) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// Live is the liveness probe — cheap, never touches dependencies.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": h.svc.Live()})
}

// Ready is the readiness probe — checks every registered dependency.
// Returns 503 when any critical component is down so the orchestrator can
// route traffic away.
func (h *HealthHandler) Ready(c *gin.Context) {
	report := h.svc.Ready(c.Request.Context())
	code := http.StatusOK
	if report.Status == health.StatusDown {
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, report)
}
