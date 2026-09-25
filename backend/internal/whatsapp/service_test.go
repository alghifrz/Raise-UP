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

type memorySettings struct {
	enabled bool
}

func (m *memorySettings) BotEnabled(context.Context) (bool, error) {
	return m.enabled, nil
}

func (m *memorySettings) SetBotEnabled(_ context.Context, enabled bool) (bool, error) {
	m.enabled = enabled
	return enabled, nil
}

type recordingBot struct {
	phones []string
}

func (b *recordingBot) HandleInbound(_ context.Context, phone, _, _, _ string) error {
	b.phones = append(b.phones, phone)
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
	}, nil, inbox, nil, nil, nil)

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
	}, nil, inbox, nil, nil, nil)

	if err := svc.ProcessWebhookPayload(context.Background(), textPayload("111111111111111", "6281111111111", "wamid.A", "halo ke nomor A")); err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(inbox.froms) != 1 || inbox.froms[0] != "6281111111111" {
		t.Fatalf("expected number A to be ingested, got %v", inbox.froms)
	}
}

func TestProcessWebhookPayloadSkipsBotWhenDisabled(t *testing.T) {
	inbox := &memoryInbox{}
	bot := &recordingBot{}
	svc := NewService(config.WhatsAppConfig{
		Enabled:       true,
		PhoneNumberID: "111111111111111",
	}, nil, inbox, bot, &memorySettings{enabled: false}, nil)

	if err := svc.ProcessWebhookPayload(context.Background(), textPayload("111111111111111", "6281111111111", "wamid.A", "halo")); err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(inbox.froms) != 1 {
		t.Fatalf("expected inbound to be stored, got %v", inbox.froms)
	}
	if len(bot.phones) != 0 {
		t.Fatalf("expected bot to stay quiet, got %v", bot.phones)
	}
}

func TestProcessWebhookPayloadRunsBotWhenEnabled(t *testing.T) {
	inbox := &memoryInbox{}
	bot := &recordingBot{}
	svc := NewService(config.WhatsAppConfig{
		Enabled:       true,
		PhoneNumberID: "111111111111111",
	}, nil, inbox, bot, &memorySettings{enabled: true}, nil)

	if err := svc.ProcessWebhookPayload(context.Background(), textPayload("111111111111111", "6281111111111", "wamid.A", "halo")); err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(bot.phones) != 1 || bot.phones[0] != "6281111111111" {
		t.Fatalf("expected bot reply, got %v", bot.phones)
	}
}

func TestSetBotEnabledPersistsToggle(t *testing.T) {
	settings := &memorySettings{enabled: true}
	svc := NewService(config.WhatsAppConfig{Enabled: true}, nil, nil, nil, settings, nil)

	status, err := svc.SetBotEnabled(context.Background(), false)
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if status.Enabled != true || status.BotEnabled != false {
		t.Fatalf("unexpected status: %+v", status)
	}
	if svc.CurrentStatus(context.Background()).BotEnabled {
		t.Fatal("expected bot to be off")
	}
}
