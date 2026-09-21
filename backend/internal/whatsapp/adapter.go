package whatsapp

import (
	"context"

	"github.com/diuk/raiseup/config"
)

// MessengerAdapter exposes the Graph client to the chat package.
type MessengerAdapter struct {
	client  *Client
	enabled bool
}

// NewMessengerAdapter builds an outbound messenger from config.
func NewMessengerAdapter(cfg config.WhatsAppConfig) *MessengerAdapter {
	return &MessengerAdapter{
		client:  NewClient(cfg.APIVersion, cfg.PhoneNumberID, cfg.AccessToken),
		enabled: cfg.Enabled,
	}
}

// Enabled reports whether WhatsApp sending is available.
func (m *MessengerAdapter) Enabled() bool {
	return m != nil && m.enabled && m.client != nil
}

// SendText sends a text message and returns the Meta message id.
func (m *MessengerAdapter) SendText(ctx context.Context, to, body string) (string, error) {
	if !m.Enabled() {
		return "", ErrNotConfigured
	}
	normalized, err := NormalizePhone(to)
	if err != nil {
		return "", err
	}
	result, err := m.client.SendText(ctx, normalized, body)
	if err != nil {
		return "", err
	}
	return result.MessageID, nil
}

// MarkAsRead marks an inbound WhatsApp message as read on Meta.
func (m *MessengerAdapter) MarkAsRead(ctx context.Context, waMessageID string) error {
	if !m.Enabled() {
		return ErrNotConfigured
	}
	return m.client.MarkAsRead(ctx, waMessageID)
}

// Client returns the underlying Graph client (for notifications).
func (m *MessengerAdapter) Client() *Client {
	if m == nil {
		return nil
	}
	return m.client
}

// PhoneHelper implements chat.PhoneNormalizer.
type PhoneHelper struct{}

// Normalize implements chat.PhoneNormalizer.
func (PhoneHelper) Normalize(raw string) (string, error) {
	return NormalizePhone(raw)
}

// MatchCandidates implements chat.PhoneNormalizer.
func (PhoneHelper) MatchCandidates(digits string) []string {
	return PhoneMatchCandidates(digits)
}
