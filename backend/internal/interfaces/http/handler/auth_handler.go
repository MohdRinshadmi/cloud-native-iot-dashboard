package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/middleware"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

const (
	refreshCookieName = "refresh_token"
	// The cookie is scoped to the auth endpoints only — it is never sent with
	// regular API traffic, shrinking its exposure surface.
	refreshCookiePath = "/api/v1/auth"
)

// AuthHandler exposes registration, login, refresh rotation and logout.
// The refresh token travels exclusively in an httpOnly cookie: JavaScript
// can never read it, which is the XSS defense the spec demands.
type AuthHandler struct {
	svc    *appauth.Service
	secure bool // Secure cookie flag (true outside development)
}

// NewAuthHandler wires the handler.
func NewAuthHandler(svc *appauth.Service, secureCookies bool) *AuthHandler {
	return &AuthHandler{svc: svc, secure: secureCookies}
}

// Register provisions a tenant + admin and signs them in.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperror.InvalidInput(bindingMessage(err)))
		return
	}

	session, err := h.svc.Register(c.Request.Context(), appauth.RegisterInput{
		TenantName: req.TenantName, Email: req.Email, Name: req.Name, Password: req.Password,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	h.setRefreshCookie(c, session)
	c.JSON(http.StatusCreated, toSessionResponse(session))
}

// Login authenticates with email + password.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperror.InvalidInput(bindingMessage(err)))
		return
	}

	session, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	h.setRefreshCookie(c, session)
	c.JSON(http.StatusOK, toSessionResponse(session))
}

// Refresh rotates the refresh token and mints a new access token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	raw, _ := c.Cookie(refreshCookieName)

	session, err := h.svc.Refresh(c.Request.Context(), raw)
	if err != nil {
		h.clearRefreshCookie(c)
		respondError(c, err)
		return
	}

	h.setRefreshCookie(c, session)
	c.JSON(http.StatusOK, toSessionResponse(session))
}

// Logout revokes the current refresh token and clears the cookie.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	raw, _ := c.Cookie(refreshCookieName)
	if err := h.svc.Logout(c.Request.Context(), raw); err != nil {
		respondError(c, err)
		return
	}
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

// Me returns the authenticated user.
// GET /api/v1/auth/me  (requires AuthRequired)
func (h *AuthHandler) Me(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	u, err := h.svc.Me(c.Request.Context(), p.UserID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

// ---- helpers --------------------------------------------------------------------

func toSessionResponse(s *appauth.Session) sessionResponse {
	return sessionResponse{
		AccessToken: s.AccessToken,
		ExpiresIn:   int64(s.ExpiresIn / time.Second),
		User:        toUserResponse(s.User),
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, s *appauth.Session) {
	c.SetSameSite(http.SameSiteLaxMode)
	maxAge := int(time.Until(s.RefreshExpiresAt) / time.Second)
	c.SetCookie(refreshCookieName, s.RefreshToken, maxAge, refreshCookiePath, "", h.secure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, refreshCookiePath, "", h.secure, true)
}

// bindingMessage keeps validator errors human-readable without leaking
// reflection internals.
func bindingMessage(err error) string {
	if err == nil {
		return "invalid request body"
	}
	return "invalid request: " + err.Error()
}
