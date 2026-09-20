package whatsapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/diuk/raiseup/config"
)

// Inbox is the chat-side ingest surface used by the webhook.
type Inbox interface {
	IngestWhatsAppInbound(ctx context.Context, phone, contactName, waMessageID, body string) error
	ApplyWhatsAppStatus(ctx context.Context, waMessageID, status string) error
}

// Service handles notifications and webhook processing.
type Service struct {
	cfg       configView
	client    *Client
	inbox     Inbox
	enabled   bool
}

type configView struct {
	VerifyToken string
	AppSecret   string
}

// NewService creates a WhatsApp application service.
func NewService(cfg config.WhatsAppConfig, messenger *MessengerAdapter, inbox Inbox) *Service {
	var client *Client
	if messenger != nil {
		client = messenger.Client()
	}
	return &Service{
		cfg: configView{
			VerifyToken: cfg.VerifyToken,
			AppSecret:   cfg.AppSecret,
		},
		client:  client,
		inbox:   inbox,
		enabled: cfg.Enabled,
	}
}

// Enabled reports whether WhatsApp integration is active.
func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

// VerifyToken returns the configured webhook verify token.
func (s *Service) VerifyToken() string {
	if s == nil {
		return ""
	}
	return s.cfg.VerifyToken
}

// AppSecret returns the Meta app secret for signature validation.
func (s *Service) AppSecret() string {
	if s == nil {
		return ""
	}
	return s.cfg.AppSecret
}

// NotificationRequest is the admin API body for outbound notifications.
type NotificationRequest struct {
	To           string              `json:"to"`
	Type         string              `json:"type"`
	Body         string              `json:"body"`
	TemplateName string              `json:"template_name"`
	Language     string              `json:"language"`
	Components   []TemplateComponent `json:"components"`
}

// NotificationResult is returned after a successful send.
type NotificationResult struct {
	To        string `json:"to"`
	MessageID string `json:"message_id"`
	Type      string `json:"type"`
}

// SendNotification sends a text or template message.
func (s *Service) SendNotification(ctx context.Context, req NotificationRequest) (*NotificationResult, error) {
	if !s.Enabled() || s.client == nil {
		return nil, ErrNotConfigured
	}

	to, err := NormalizePhone(req.To)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid to phone", ErrInvalidRequest)
	}

	kind := strings.ToLower(strings.TrimSpace(req.Type))
	switch kind {
	case "text":
		body := strings.TrimSpace(req.Body)
		if body == "" {
			return nil, fmt.Errorf("%w: body is required for text", ErrInvalidRequest)
		}
		result, err := s.client.SendText(ctx, to, body)
		if err != nil {
			return nil, err
		}
		return &NotificationResult{To: to, MessageID: result.MessageID, Type: "text"}, nil
	case "template":
		name := strings.TrimSpace(req.TemplateName)
		if name == "" {
			return nil, fmt.Errorf("%w: template_name is required", ErrInvalidRequest)
		}
		lang := strings.TrimSpace(req.Language)
		if lang == "" {
			lang = "id"
		}
		result, err := s.client.SendTemplate(ctx, to, name, lang, req.Components)
		if err != nil {
			return nil, err
		}
		return &NotificationResult{To: to, MessageID: result.MessageID, Type: "template"}, nil
	default:
		return nil, fmt.Errorf("%w: type must be text or template", ErrInvalidRequest)
	}
}

// ProcessWebhookPayload handles an inbound Meta webhook body.
func (s *Service) ProcessWebhookPayload(ctx context.Context, payload *WebhookPayload) error {
	if s.inbox == nil {
		return nil
	}
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			value := change.Value
			contactsByWaID := map[string]string{}
			for _, c := range value.Contacts {
				contactsByWaID[c.WaID] = c.Profile.Name
			}
			for _, msg := range value.Messages {
				if msg.Type != "" && msg.Type != "text" {
					continue
				}
				body := ""
				if msg.Text != nil {
					body = msg.Text.Body
				}
				if strings.TrimSpace(body) == "" {
					continue
				}
				name := contactsByWaID[msg.From]
				if err := s.inbox.IngestWhatsAppInbound(ctx, msg.From, name, msg.ID, body); err != nil {
					return err
				}
			}
			for _, st := range value.Statuses {
				if err := s.inbox.ApplyWhatsAppStatus(ctx, st.ID, st.Status); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
