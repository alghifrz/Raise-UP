package finance

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes finance HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a finance HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts finance routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/finance")
	group.Use(middlewares...)

	group.GET("/summary", h.Summary)
	group.GET("/transactions", h.List)
	group.POST("/transactions", h.Create)
	group.GET("/transactions/:id", h.Get)
	group.PATCH("/transactions/:id", h.Update)
	group.DELETE("/transactions/:id", h.Delete)
}

// List handles GET /api/v1/finance/transactions.
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
		c.Query("type"),
		c.Query("category"),
		c.Query("search"),
		c.Query("from"),
		c.Query("to"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// Get handles GET /api/v1/finance/transactions/:id.
func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Create handles POST /api/v1/finance/transactions.
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

// Update handles PATCH /api/v1/finance/transactions/:id.
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

// Delete handles DELETE /api/v1/finance/transactions/:id.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

// Summary handles GET /api/v1/finance/summary.
func (h *Handler) Summary(c *gin.Context) {
	item, err := h.service.Summary(c.Request.Context(), c.Query("from"), c.Query("to"))
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
		response.NotFound(c, "TRANSACTION_NOT_FOUND", "Transaction not found")
	default:
		h.log.Error("finance handler error", "error", err)
		response.Internal(c)
	}
}
