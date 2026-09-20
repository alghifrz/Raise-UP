package dues

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes dues HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a dues HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts dues routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/dues")
	group.Use(middlewares...)

	group.GET("/periods", h.ListPeriods)
	group.POST("/periods", h.CreatePeriod)
	group.GET("/periods/:id/summary", h.PeriodSummary)
	group.GET("/periods/:id/status", h.ListResidentPaymentStatus)
	group.GET("/periods/:id", h.GetPeriod)
	group.PATCH("/periods/:id", h.UpdatePeriod)
	group.DELETE("/periods/:id", h.DeletePeriod)

	group.GET("/payments", h.ListPayments)
	group.POST("/payments", h.CreatePayment)
	group.GET("/payments/:id", h.GetPayment)
	group.PATCH("/payments/:id", h.UpdatePayment)
	group.DELETE("/payments/:id", h.DeletePayment)
}

// ListPeriods handles GET /api/v1/dues/periods.
func (h *Handler) ListPeriods(c *gin.Context) {
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListPeriods(
		c.Request.Context(),
		page,
		pageSize,
		c.Query("year"),
		c.Query("month"),
		c.Query("half"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// GetPeriod handles GET /api/v1/dues/periods/:id.
func (h *Handler) GetPeriod(c *gin.Context) {
	item, err := h.service.GetPeriod(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// CreatePeriod handles POST /api/v1/dues/periods.
func (h *Handler) CreatePeriod(c *gin.Context) {
	var req CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.CreatePeriod(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// UpdatePeriod handles PATCH /api/v1/dues/periods/:id.
func (h *Handler) UpdatePeriod(c *gin.Context) {
	var req UpdatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.UpdatePeriod(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// DeletePeriod handles DELETE /api/v1/dues/periods/:id.
// Deleting a period cascades and permanently deletes associated payments.
func (h *Handler) DeletePeriod(c *gin.Context) {
	if err := h.service.DeletePeriod(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

// PeriodSummary handles GET /api/v1/dues/periods/:id/summary.
func (h *Handler) PeriodSummary(c *gin.Context) {
	item, err := h.service.PeriodSummary(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// ListResidentPaymentStatus handles GET /api/v1/dues/periods/:id/status.
func (h *Handler) ListResidentPaymentStatus(c *gin.Context) {
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListResidentPaymentStatus(
		c.Request.Context(),
		c.Param("id"),
		page,
		pageSize,
		c.Query("status"),
		c.Query("search"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// ListPayments handles GET /api/v1/dues/payments.
func (h *Handler) ListPayments(c *gin.Context) {
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListPayments(
		c.Request.Context(),
		page,
		pageSize,
		c.Query("period_id"),
		c.Query("resident_id"),
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

// GetPayment handles GET /api/v1/dues/payments/:id.
func (h *Handler) GetPayment(c *gin.Context) {
	item, err := h.service.GetPayment(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// CreatePayment handles POST /api/v1/dues/payments.
func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.CreatePayment(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// UpdatePayment handles PATCH /api/v1/dues/payments/:id.
func (h *Handler) UpdatePayment(c *gin.Context) {
	var req UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.UpdatePayment(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// DeletePayment handles DELETE /api/v1/dues/payments/:id.
func (h *Handler) DeletePayment(c *gin.Context) {
	if err := h.service.DeletePayment(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

func parsePagination(c *gin.Context) (page, pageSize int, ok bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return 0, 0, false
	}
	pageSize, err = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return 0, 0, false
	}
	return page, pageSize, true
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
	case errors.Is(err, ErrInvalidPaymentAmount):
		response.Error(c, http.StatusBadRequest, "INVALID_PAYMENT_AMOUNT", "Payment amount must equal the period amount")
	case errors.Is(err, ErrPeriodNotFound):
		response.NotFound(c, "DUES_PERIOD_NOT_FOUND", "Dues period not found")
	case errors.Is(err, ErrPaymentNotFound):
		response.NotFound(c, "DUES_PAYMENT_NOT_FOUND", "Dues payment not found")
	case errors.Is(err, ErrResidentNotFound):
		response.NotFound(c, "RESIDENT_NOT_FOUND", "Resident not found")
	case errors.Is(err, ErrPeriodAlreadyExists):
		response.Conflict(c, "DUES_PERIOD_ALREADY_EXISTS", "Dues period already exists")
	case errors.Is(err, ErrPaymentAlreadyExists):
		response.Conflict(c, "DUES_PAYMENT_ALREADY_EXISTS", "Dues payment already exists for this period and resident")
	default:
		h.log.Error("dues handler error", "error", err)
		response.Internal(c)
	}
}
