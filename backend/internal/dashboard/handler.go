package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes dashboard HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a dashboard HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts dashboard routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/dashboard")
	group.Use(middlewares...)
	group.GET("/summary", h.Summary)
}

// Summary handles GET /api/v1/dashboard/summary.
func (h *Handler) Summary(c *gin.Context) {
	item, err := h.service.Summary(c.Request.Context())
	if err != nil {
		h.log.Error("dashboard handler error", "error", err)
		response.Internal(c)
		return
	}
	response.JSON(c, http.StatusOK, item)
}
