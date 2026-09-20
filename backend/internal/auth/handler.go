package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/diuk/raiseup/pkg/authctx"
	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes authentication HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates an auth HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts auth routes on the given /api/v1 group.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	authGroup := api.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.GET("/me", requireAuth, h.Me)
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, result)
}

// Me handles GET /api/v1/auth/me.
func (h *Handler) Me(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	user, err := h.service.Me(c.Request.Context(), principal.UserID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, user)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		response.BadRequest(c, "Invalid email or password")
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(c, "INVALID_CREDENTIALS", "Invalid email or password")
	case errors.Is(err, ErrUnauthorized):
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "Insufficient permissions")
	default:
		h.log.Error("auth handler error", "error", err)
		response.Internal(c)
	}
}
