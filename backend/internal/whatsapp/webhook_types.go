package whatsapp

// WebhookPayload is the Meta WhatsApp Cloud API webhook body.
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry is a webhook entry.
type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

// WebhookChange is a webhook change.
type WebhookChange struct {
	Field string       `json:"field"`
	Value WebhookValue `json:"value"`
}

// WebhookValue holds messages and statuses.
type WebhookValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         WebhookMetadata  `json:"metadata"`
	Contacts         []WebhookContact `json:"contacts"`
	Messages         []WebhookMessage `json:"messages"`
	Statuses         []WebhookStatus  `json:"statuses"`
}

// WebhookMetadata contains phone number id metadata.
type WebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// WebhookContact is an inbound contact profile.
type WebhookContact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WaID string `json:"wa_id"`
}

// WebhookMessage is an inbound message.
type WebhookMessage struct {
	From      string             `json:"from"`
	ID        string             `json:"id"`
	Timestamp string             `json:"timestamp"`
	Type      string             `json:"type"`
	Text      *WebhookTextBody   `json:"text"`
}

// WebhookTextBody is a text message body.
type WebhookTextBody struct {
	Body string `json:"body"`
}

// WebhookStatus is an outbound delivery status update.
type WebhookStatus struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	RecipientID string `json:"recipient_id"`
}
