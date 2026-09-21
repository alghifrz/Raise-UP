package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to Meta WhatsApp Cloud API.
type Client struct {
	httpClient    *http.Client
	apiVersion    string
	phoneNumberID string
	accessToken   string
	baseURL       string
}

// NewClient creates a Graph API client.
func NewClient(apiVersion, phoneNumberID, accessToken string) *Client {
	version := strings.TrimSpace(apiVersion)
	if version == "" {
		version = "v21.0"
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		apiVersion:    version,
		phoneNumberID: strings.TrimSpace(phoneNumberID),
		accessToken:   strings.TrimSpace(accessToken),
		baseURL:       "https://graph.facebook.com",
	}
}

// SendTextResult is the Graph response for a successful send.
type SendTextResult struct {
	MessageID string
}

type graphSendResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error *graphError `json:"error"`
}

type graphError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

// TemplateComponent is a WhatsApp template component payload.
type TemplateComponent struct {
	Type       string              `json:"type"`
	SubType    string              `json:"sub_type,omitempty"`
	Index      *int                `json:"index,omitempty"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateParameter is a template variable.
type TemplateParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// SendText sends a plain text message to an E.164 phone (digits only, no +).
func (c *Client) SendText(ctx context.Context, to, body string) (*SendTextResult, error) {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text": map[string]any{
			"preview_url": false,
			"body":        body,
		},
	}
	return c.send(ctx, payload)
}

// MarkAsRead tells Meta the inbound message was read (blue ticks on the contact's phone).
func (c *Client) MarkAsRead(ctx context.Context, waMessageID string) error {
	waMessageID = strings.TrimSpace(waMessageID)
	if waMessageID == "" {
		return nil
	}
	if c.phoneNumberID == "" || c.accessToken == "" {
		return fmt.Errorf("%w: WhatsApp client is not configured", ErrNotConfigured)
	}

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        waMessageID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal mark-as-read payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/messages", c.baseURL, c.apiVersion, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create mark-as-read request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read mark-as-read response: %w", err)
	}

	var parsed struct {
		Success bool        `json:"success"`
		Error   *graphError `json:"error"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("%w: invalid response (%s)", ErrSendFailed, truncate(string(respBody), 200))
	}
	if parsed.Error != nil {
		return fmt.Errorf("%w: %s", ErrSendFailed, parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: HTTP %d", ErrSendFailed, resp.StatusCode)
	}
	return nil
}

// SendTemplate sends an approved message template.
func (c *Client) SendTemplate(
	ctx context.Context,
	to, templateName, languageCode string,
	components []TemplateComponent,
) (*SendTextResult, error) {
	template := map[string]any{
		"name": templateName,
		"language": map[string]any{
			"code": languageCode,
		},
	}
	if len(components) > 0 {
		template["components"] = components
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "template",
		"template":          template,
	}
	return c.send(ctx, payload)
}

func (c *Client) send(ctx context.Context, payload map[string]any) (*SendTextResult, error) {
	if c.phoneNumberID == "" || c.accessToken == "" {
		return nil, fmt.Errorf("%w: WhatsApp client is not configured", ErrNotConfigured)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/messages", c.baseURL, c.apiVersion, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create whatsapp request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read whatsapp response: %w", err)
	}

	var parsed graphSendResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("%w: invalid response (%s)", ErrSendFailed, truncate(string(respBody), 200))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrSendFailed, parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSendFailed, resp.StatusCode)
	}
	if len(parsed.Messages) == 0 || parsed.Messages[0].ID == "" {
		return nil, fmt.Errorf("%w: missing message id in response", ErrSendFailed)
	}
	return &SendTextResult{MessageID: parsed.Messages[0].ID}, nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "…"
}
