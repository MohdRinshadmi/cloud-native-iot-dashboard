package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/application/devices"
	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// TelemetryHandler exposes per-device telemetry reads. Tenant ownership of the
// device is verified before any telemetry query — no cross-tenant leakage via
// guessed device ids.
type TelemetryHandler struct {
	devices *devices.Service
	history telemetry.Repository
	latest  telemetry.LatestStore
}

// NewTelemetryHandler wires the handler.
func NewTelemetryHandler(devices *devices.Service, history telemetry.Repository, latest telemetry.LatestStore) *TelemetryHandler {
	return &TelemetryHandler{devices: devices, history: history, latest: latest}
}

// History returns recent readings, newest first.
// GET /api/v1/devices/:id/telemetry?minutes=60&limit=500
func (h *TelemetryHandler) History(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	deviceID := c.Param("id")

	// Ownership check doubles as 404 for foreign/unknown devices.
	if _, err := h.devices.Get(c.Request.Context(), p.TenantID, deviceID); err != nil {
		respondError(c, err)
		return
	}

	minutes, _ := strconv.Atoi(c.DefaultQuery("minutes", "60"))
	if minutes <= 0 || minutes > 24*60 {
		minutes = 60
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))

	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)
	readings, err := h.history.ListRecent(c.Request.Context(), p.TenantID, deviceID, since, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": readings})
}

// Latest returns the newest reading from the hot store.
// GET /api/v1/devices/:id/telemetry/latest
func (h *TelemetryHandler) Latest(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	deviceID := c.Param("id")

	if _, err := h.devices.Get(c.Request.Context(), p.TenantID, deviceID); err != nil {
		respondError(c, err)
		return
	}

	r, err := h.latest.GetLatest(c.Request.Context(), deviceID)
	if err != nil {
		respondError(c, err)
		return
	}
	// Defense in depth: the cache is keyed by device id; re-check tenant.
	if r.TenantID != p.TenantID {
		respondError(c, apperror.NotFound("no telemetry for device"))
		return
	}
	c.JSON(http.StatusOK, r)
}
