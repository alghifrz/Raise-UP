package village

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes village profile and officials HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a village HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts village profile/official routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/village-profile")
	group.Use(middlewares...)

	group.GET("", h.GetProfile)
	group.PATCH("", h.UpdateProfile)

	group.GET("/officials", h.ListOfficials)
	group.POST("/officials", h.CreateOfficial)
	group.PATCH("/officials/:id", h.UpdateOfficial)
	group.DELETE("/officials/:id", h.DeleteOfficial)
}

// RegisterPublicRoutes mounts read-only village profile data for the public portal.
func (h *Handler) RegisterPublicRoutes(api *gin.RouterGroup) {
	group := api.Group("/public/village-profile")
	group.GET("", h.GetProfile)
	group.GET("/officials", h.ListOfficials)
}

// GetProfile handles GET /api/v1/village-profile.
func (h *Handler) GetProfile(c *gin.Context) {
	item, err := h.service.GetProfile(c.Request.Context())
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// UpdateProfile handles PATCH /api/v1/village-profile.
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.UpdateProfile(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// ListOfficials handles GET /api/v1/village-profile/officials.
func (h *Handler) ListOfficials(c *gin.Context) {
	items, err := h.service.ListOfficials(c.Request.Context())
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, items)
}

// CreateOfficial handles POST /api/v1/village-profile/officials.
func (h *Handler) CreateOfficial(c *gin.Context) {
	var req CreateOfficialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.CreateOfficial(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// UpdateOfficial handles PATCH /api/v1/village-profile/officials/:id.
func (h *Handler) UpdateOfficial(c *gin.Context) {
	var req UpdateOfficialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.UpdateOfficial(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// DeleteOfficial handles DELETE /api/v1/village-profile/officials/:id.
func (h *Handler) DeleteOfficial(c *gin.Context) {
	if err := h.service.DeleteOfficial(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.NoContent(c)
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
	case errors.Is(err, ErrProfileNotFound):
		response.NotFound(c, "VILLAGE_PROFILE_NOT_FOUND", "Village profile not found")
	case errors.Is(err, ErrOfficialNotFound):
		response.NotFound(c, "VILLAGE_OFFICIAL_NOT_FOUND", "Village official not found")
	default:
		h.log.Error("village handler error", "error", err)
		response.Internal(c)
	}
}
