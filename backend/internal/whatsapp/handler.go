package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes WhatsApp webhook and notification HTTP endpoints.
type Handler struct {
	service *Service
	log     *slog.Logger
}

// NewHandler creates a WhatsApp HTTP handler.
func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterWebhookRoutes mounts public Meta webhook routes (no JWT).
func (h *Handler) RegisterWebhookRoutes(api *gin.RouterGroup) {
	group := api.Group("/webhooks/whatsapp")
	group.GET("", h.Verify)
	group.POST("", h.Receive)
}

// RegisterRoutes mounts protected admin notification routes.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/whatsapp")
	group.Use(middlewares...)
	group.POST("/notifications", h.SendNotification)
	group.GET("/status", h.Status)
	group.PATCH("/bot", h.UpdateBot)
}

// Verify handles GET /api/v1/webhooks/whatsapp (Meta challenge).
func (h *Handler) Verify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode != "subscribe" || token == "" || challenge == "" {
		response.BadRequest(c, "invalid webhook verification request")
		return
	}
	if h.service == nil || token != h.service.VerifyToken() {
		response.Forbidden(c, "verify token mismatch")
		return
	}
	c.String(http.StatusOK, challenge)
}

// Receive handles POST /api/v1/webhooks/whatsapp.
func (h *Handler) Receive(c *gin.Context) {
	if h.service == nil || !h.service.Enabled() {
		response.Error(c, http.StatusServiceUnavailable, "WHATSAPP_DISABLED", "WhatsApp is not configured")
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		response.BadRequest(c, "unable to read body")
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if !validSignature(h.service.AppSecret(), signature, body) {
		response.Forbidden(c, "invalid webhook signature")
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		response.BadRequest(c, "invalid webhook payload")
		return
	}

	if err := h.service.ProcessWebhookPayload(c.Request.Context(), &payload); err != nil {
		h.log.Error("whatsapp webhook processing failed", "error", err)
		response.Internal(c)
		return
	}

	c.Status(http.StatusOK)
}

// SendNotification handles POST /api/v1/whatsapp/notifications.
func (h *Handler) SendNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.SendNotification(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, result)
}

// Status handles GET /api/v1/whatsapp/status.
func (h *Handler) Status(c *gin.Context) {
	if h.service == nil {
		response.JSON(c, http.StatusOK, Status{})
		return
	}
	response.JSON(c, http.StatusOK, h.service.CurrentStatus(c.Request.Context()))
}

type updateBotRequest struct {
	Enabled *bool `json:"enabled"`
}

// UpdateBot handles PATCH /api/v1/whatsapp/bot.
func (h *Handler) UpdateBot(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "WHATSAPP_DISABLED", "WhatsApp is not configured")
		return
	}

	var req updateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		response.BadRequest(c, "enabled is required")
		return
	}

	status, err := h.service.SetBotEnabled(c.Request.Context(), *req.Enabled)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, status)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotConfigured):
		response.Error(c, http.StatusServiceUnavailable, "WHATSAPP_DISABLED", "WhatsApp is not configured")
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, "WHATSAPP_SETTINGS_NOT_FOUND", "WhatsApp settings were not found")
	case errors.Is(err, ErrInvalidRequest):
		message := err.Error()
		message = strings.TrimPrefix(message, ErrInvalidRequest.Error()+": ")
		if message == ErrInvalidRequest.Error() {
			message = "Invalid request"
		}
		response.BadRequest(c, message)
	case errors.Is(err, ErrSendFailed):
		message := err.Error()
		message = strings.TrimPrefix(message, ErrSendFailed.Error()+": ")
		response.Error(c, http.StatusBadGateway, "WHATSAPP_SEND_FAILED", message)
	default:
		h.log.Error("whatsapp handler error", "error", err)
		response.Internal(c)
	}
}

func validSignature(appSecret, header string, body []byte) bool {
	if appSecret == "" || header == "" {
		return false
	}
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	provided := strings.TrimPrefix(header, prefix)
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(provided))
}
