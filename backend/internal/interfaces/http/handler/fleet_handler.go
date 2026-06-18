package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/application/devices"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
)

// FleetHandler serves tenant-level fleet aggregates for the dashboard.
type FleetHandler struct {
	devices *devices.Service
}

// NewFleetHandler wires the handler.
func NewFleetHandler(d *devices.Service) *FleetHandler { return &FleetHandler{devices: d} }

type fleetSummaryResponse struct {
	Total    int64            `json:"total"`
	Online   int64            `json:"online"`
	Offline  int64            `json:"offline"`
	Degraded int64            `json:"degraded"`
	Other    int64            `json:"other"`
	ByStatus map[string]int64 `json:"by_status"`
}

// Summary returns the fleet status distribution.
// GET /api/v1/fleet/summary
func (h *FleetHandler) Summary(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	sum, err := h.devices.Summary(c.Request.Context(), p.TenantID)
	if err != nil {
		respondError(c, err)
		return
	}

	byStatus := make(map[string]int64, len(sum.ByStatus))
	for status, count := range sum.ByStatus {
		byStatus[string(status)] = count
	}
	c.JSON(http.StatusOK, fleetSummaryResponse{
		Total:    sum.Total,
		Online:   sum.Online,
		Offline:  sum.Offline,
		Degraded: sum.Degraded,
		Other:    sum.Other,
		ByStatus: byStatus,
	})
}
