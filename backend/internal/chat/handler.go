package chat

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/diuk/raiseup/pkg/authctx"
	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes chat HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a chat HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes mounts chat routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/conversations")
	group.Use(middlewares...)

	group.GET("", h.ListConversations)
	group.POST("", h.CreateConversation)
	group.GET("/contacts", h.ListContacts)
	group.GET("/:id", h.GetConversation)
	group.GET("/:id/messages", h.ListMessages)
	group.POST("/:id/messages", h.SendMessage)
	group.POST("/:id/read", h.MarkRead)
}

// ListConversations handles GET /api/v1/conversations.
func (h *Handler) ListConversations(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListConversations(
		c.Request.Context(),
		principal.UserID,
		page,
		pageSize,
		c.Query("search"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// GetConversation handles GET /api/v1/conversations/:id.
func (h *Handler) GetConversation(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	item, err := h.service.GetConversation(c.Request.Context(), principal.UserID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// CreateConversation handles POST /api/v1/conversations.
func (h *Handler) CreateConversation(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.CreateConversation(c.Request.Context(), principal.UserID, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// ListMessages handles GET /api/v1/conversations/:id/messages.
func (h *Handler) ListMessages(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListMessages(
		c.Request.Context(),
		principal.UserID,
		c.Param("id"),
		page,
		pageSize,
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// SendMessage handles POST /api/v1/conversations/:id/messages.
func (h *Handler) SendMessage(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.SendMessage(c.Request.Context(), principal.UserID, c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// MarkRead handles POST /api/v1/conversations/:id/read.
func (h *Handler) MarkRead(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	item, err := h.service.MarkRead(c.Request.Context(), principal.UserID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// ListContacts handles GET /api/v1/conversations/contacts.
func (h *Handler) ListContacts(c *gin.Context) {
	principal, ok := authctx.GetPrincipal(c)
	if !ok {
		response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListContacts(
		c.Request.Context(),
		principal.UserID,
		page,
		pageSize,
		c.Query("search"),
	)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
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
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "CONVERSATION_NOT_FOUND", "Conversation not found")
	case errors.Is(err, ErrUserNotFound):
		response.NotFound(c, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "You are not a participant of this conversation")
	default:
		h.log.Error("chat handler error", "error", err)
		response.Internal(c)
	}
}
