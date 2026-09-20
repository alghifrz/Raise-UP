package sitesettings

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes site settings HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a site settings HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts site settings routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/site-settings")
	group.Use(middlewares...)

	group.GET("", h.Get)
	group.PATCH("", h.Update)
}

// RegisterPublicRoutes mounts read-only site settings for the public portal.
func (h *Handler) RegisterPublicRoutes(api *gin.RouterGroup) {
	api.GET("/public/site-settings", h.Get)
}

// Get handles GET /api/v1/site-settings.
func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context())
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Update handles PATCH /api/v1/site-settings.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.Update(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		message := err.Error()
		message = strings.TrimPrefix(message, ErrInvalidRequest.Error()+": ")
		if message == ErrInvalidRequest.Error() {
			message = "Invalid request"
		}
		response.BadRequest(c, message)
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "SITE_SETTINGS_NOT_FOUND", "Site settings not found")
	default:
		h.log.Error("site settings handler error", "error", err)
		response.Internal(c)
	}
}
