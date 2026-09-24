package whatsapp

import (
	"context"
	"testing"

	"github.com/diuk/raiseup/config"
)

type memoryInbox struct {
	froms []string
}

func (m *memoryInbox) IngestWhatsAppInbound(_ context.Context, phone, _, _, _ string) (bool, error) {
	m.froms = append(m.froms, phone)
	return true, nil
}

func (m *memoryInbox) ApplyWhatsAppStatus(context.Context, string, string) error {
	return nil
}

func textPayload(phoneNumberID, from, id, body string) *WebhookPayload {
	return &WebhookPayload{
		Entry: []WebhookEntry{{
			Changes: []WebhookChange{{
				Value: WebhookValue{
					Metadata: WebhookMetadata{PhoneNumberID: phoneNumberID},
					Messages: []WebhookMessage{{
						From: from,
						ID:   id,
						Type: "text",
						Text: &WebhookTextBody{Body: body},
					}},
				},
			}},
		}},
	}
}

func TestProcessWebhookPayloadIgnoresOtherPhoneNumberID(t *testing.T) {
	inbox := &memoryInbox{}
	svc := NewService(config.WhatsAppConfig{
		Enabled:       true,
		PhoneNumberID: "111111111111111",
		VerifyToken:   "verify",
		AppSecret:     "secret",
	}, nil, inbox, nil, nil)

	if err := svc.ProcessWebhookPayload(context.Background(), textPayload("222222222222222", "6281111111111", "wamid.B", "halo ke nomor B")); err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(inbox.froms) != 0 {
		t.Fatalf("expected number B to be ignored, ingested %v", inbox.froms)
	}
}

func TestProcessWebhookPayloadAcceptsConfiguredPhoneNumberID(t *testing.T) {
	inbox := &memoryInbox{}
	svc := NewService(config.WhatsAppConfig{
		Enabled:       true,
		PhoneNumberID: "111111111111111",
		VerifyToken:   "verify",
		AppSecret:     "secret",
	}, nil, inbox, nil, nil)

	if err := svc.ProcessWebhookPayload(context.Background(), textPayload("111111111111111", "6281111111111", "wamid.A", "halo ke nomor A")); err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(inbox.froms) != 1 || inbox.froms[0] != "6281111111111" {
		t.Fatalf("expected number A to be ingested, got %v", inbox.froms)
	}
}
