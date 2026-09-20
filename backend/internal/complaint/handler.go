package complaint

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes complaint HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a complaint HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts complaint routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/complaints")
	group.Use(middlewares...)

	group.GET("", h.List)
	group.POST("", h.Create)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.PATCH("/:id/status", h.UpdateStatus)
	group.DELETE("/:id", h.Delete)
}

// List handles GET /api/v1/complaints.
func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return
	}

	result, err := h.service.List(
		c.Request.Context(),
		page,
		pageSize,
		c.Query("search"),
		c.Query("status"),
		c.Query("urgency"),
		c.Query("category"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// Get handles GET /api/v1/complaints/:id.
func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Create handles POST /api/v1/complaints.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// Update handles PATCH /api/v1/complaints/:id.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// UpdateStatus handles PATCH /api/v1/complaints/:id/status.
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.UpdateStatus(c.Request.Context(), c.Param("id"), req.Status)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Delete handles DELETE /api/v1/complaints/:id.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
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
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "COMPLAINT_NOT_FOUND", "Complaint not found")
	case errors.Is(err, ErrResidentNotFound):
		response.NotFound(c, "RESIDENT_NOT_FOUND", "Resident not found")
	case errors.Is(err, ErrInvalidStatusTransition):
		response.Conflict(c, "INVALID_STATUS_TRANSITION", "invalid complaint status transition")
	default:
		h.log.Error("complaint handler error", "error", err)
		response.Internal(c)
	}
}
