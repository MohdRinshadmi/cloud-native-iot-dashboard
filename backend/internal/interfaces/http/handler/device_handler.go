package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/application/devices"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// DeviceHandler exposes tenant-scoped device CRUD. The tenant id always comes
// from the verified JWT principal — request payloads cannot influence it.
type DeviceHandler struct {
	svc *devices.Service
}

// NewDeviceHandler wires the handler.
func NewDeviceHandler(svc *devices.Service) *DeviceHandler { return &DeviceHandler{svc: svc} }

// List returns a filtered, paginated device page.
// GET /api/v1/devices?q=&status=&limit=&offset=
func (h *DeviceHandler) List(c *gin.Context) {
	p := middleware.GetPrincipal(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.svc.List(c.Request.Context(), devices.ListInput{
		TenantID: p.TenantID,
		Q:        c.Query("q"),
		Status:   c.Query("status"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]deviceResponse, len(items))
	for i, d := range items {
		out[i] = toDeviceResponse(d)
	}
	c.JSON(http.StatusOK, paginated[deviceResponse]{
		Data: out,
		Meta: paginationMeta{Total: total, Limit: clampLimit(limit), Offset: max(offset, 0)},
	})
}

// Create registers a device.
// POST /api/v1/devices  (admin|operator)
func (h *DeviceHandler) Create(c *gin.Context) {
	p := middleware.GetPrincipal(c)

	var req createDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperror.InvalidInput(bindingMessage(err)))
		return
	}

	d, err := h.svc.Create(c.Request.Context(), devices.CreateInput{
		TenantID: p.TenantID, Name: req.Name, Model: req.Model, Firmware: req.Firmware,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDeviceResponse(d))
}

// Get fetches one device.
// GET /api/v1/devices/:id
func (h *DeviceHandler) Get(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	d, err := h.svc.Get(c.Request.Context(), p.TenantID, c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDeviceResponse(d))
}

// Update applies a partial update.
// PATCH /api/v1/devices/:id  (admin|operator)
func (h *DeviceHandler) Update(c *gin.Context) {
	p := middleware.GetPrincipal(c)

	var req updateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperror.InvalidInput(bindingMessage(err)))
		return
	}

	d, err := h.svc.Update(c.Request.Context(), p.TenantID, c.Param("id"), devices.UpdateInput{
		Name: req.Name, Model: req.Model, Firmware: req.Firmware,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDeviceResponse(d))
}

// Delete removes a device.
// DELETE /api/v1/devices/:id  (admin)
func (h *DeviceHandler) Delete(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	if err := h.svc.Delete(c.Request.Context(), p.TenantID, c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func clampLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}
