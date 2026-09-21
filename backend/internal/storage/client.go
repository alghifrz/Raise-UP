package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxErrorBodyBytes = 1 << 20

// Client talks to the Supabase Storage HTTP API with server-side credentials.
type Client struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

// NewClient creates a Supabase Storage client.
func NewClient(baseURL, serviceKey string) *Client {
	return NewClientWithHTTP(baseURL, serviceKey, &http.Client{Timeout: 30 * time.Second})
}

// NewClientWithHTTP creates a client with an injected HTTP client for tests.
func NewClientWithHTTP(baseURL, serviceKey string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		httpClient: httpClient,
	}
}

// Upload stores an object and returns its public URL.
func (c *Client) Upload(
	ctx context.Context,
	bucket, objectPath, contentType string,
	body io.Reader,
) (string, error) {
	endpoint := c.objectURL(bucket, objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return "", fmt.Errorf("create storage upload request: %w", err)
	}
	c.authorize(req)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "false")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload storage object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", responseError("upload storage object", resp)
	}

	return c.baseURL + "/storage/v1/object/public/" +
		escapePath(bucket) + "/" + escapePath(objectPath), nil
}

// Delete removes an object from a bucket.
func (c *Client) Delete(ctx context.Context, bucket, objectPath string) error {
	payload, err := json.Marshal(map[string][]string{"prefixes": {objectPath}})
	if err != nil {
		return fmt.Errorf("encode storage delete request: %w", err)
	}
	endpoint := c.baseURL + "/storage/v1/object/" + escapePath(bucket)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create storage delete request: %w", err)
	}
	c.authorize(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete storage object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError("delete storage object", resp)
	}
	return nil
}

func (c *Client) objectURL(bucket, objectPath string) string {
	return c.baseURL + "/storage/v1/object/" + escapePath(bucket) + "/" + escapePath(objectPath)
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("apikey", c.serviceKey)
	// New sb_secret keys are opaque and must not be parsed as JWTs. Legacy
	// service_role keys are JWTs and support the traditional Bearer header.
	if !strings.HasPrefix(c.serviceKey, "sb_secret_") {
		req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	}
}

func escapePath(value string) string {
	parts := strings.Split(strings.Trim(value, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func responseError(action string, resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("%s: status %d: %s", action, resp.StatusCode, message)
}
