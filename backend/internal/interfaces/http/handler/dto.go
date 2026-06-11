package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// ---- response envelopes -------------------------------------------------------

// paginated is the standard list envelope.
type paginated[T any] struct {
	Data []T            `json:"data"`
	Meta paginationMeta `json:"meta"`
}

type paginationMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// ---- error mapping --------------------------------------------------------------

// respondError converts an apperror (or unknown error) into the canonical
// `{ "error": { code, message, request_id } }` JSON shape. Internal details
// never leak to the client.
func respondError(c *gin.Context, err error) {
	code := apperror.CodeInternal
	message := "internal server error"

	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		code = appErr.Code
		if code != apperror.CodeInternal {
			message = appErr.Message
		}
	}

	c.AbortWithStatusJSON(httpStatus(code), gin.H{
		"error": gin.H{
			"code":       string(code),
			"message":    message,
			"request_id": middleware.GetRequestID(c),
		},
	})
}

func httpStatus(code apperror.Code) int {
	switch code {
	case apperror.CodeInvalidInput:
		return http.StatusBadRequest
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// ---- device DTOs ---------------------------------------------------------------

type deviceResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Model      string     `json:"model"`
	Firmware   string     `json:"firmware"`
	Status     string     `json:"status"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func toDeviceResponse(d *device.Device) deviceResponse {
	return deviceResponse{
		ID: d.ID, Name: d.Name, Model: d.Model, Firmware: d.Firmware,
		Status: string(d.Status), LastSeenAt: d.LastSeenAt,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type createDeviceRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=120"`
	Model    string `json:"model" binding:"max=120"`
	Firmware string `json:"firmware" binding:"max=60"`
}

type updateDeviceRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=120"`
	Model    *string `json:"model" binding:"omitempty,max=120"`
	Firmware *string `json:"firmware" binding:"omitempty,max=60"`
}

// ---- auth DTOs ------------------------------------------------------------------

type registerRequest struct {
	TenantName string `json:"tenant_name" binding:"required,min=2,max=80"`
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required,min=1,max=120"`
	Password   string `json:"password" binding:"required,min=8,max=128"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{ID: u.ID, TenantID: u.TenantID, Email: u.Email, Name: u.Name, Role: string(u.Role)}
}

type sessionResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresIn   int64        `json:"expires_in"` // seconds
	User        userResponse `json:"user"`
}
