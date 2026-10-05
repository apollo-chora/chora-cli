// Package client provides an HTTP client for the Chora Gateway API.
// All requests include the X-API-Key header. API keys are never logged.
package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	headerAPIKey    = "X-API-Key"
	headerUserAgent = "User-Agent"
	userAgent       = "chora-cli/0.1"
	defaultTimeout  = 30 * time.Second
)

// Client is an HTTP client for the Chora Gateway API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new Client with the given base URL and API key.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Get performs a GET request to the given path (may include query string).
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}
	c.setHeaders(req)
	return c.httpClient.Do(req)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setHeaders(req)
	return c.httpClient.Do(req)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create PUT request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setHeaders(req)
	return c.httpClient.Do(req)
}

// setHeaders adds common headers (API key, user agent).
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set(headerAPIKey, c.apiKey)
	req.Header.Set(headerUserAgent, userAgent)
}
