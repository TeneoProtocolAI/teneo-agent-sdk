package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client communicates with an OpenClaw instance via its REST API
type Client struct {
	httpClient *http.Client
	config     *Config
}

// NewClient creates a new OpenClaw API client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}, nil
}

// Chat sends a message to the configured OpenClaw agent and returns the response
func (c *Client) Chat(ctx context.Context, message string) (*ChatResponse, error) {
	return c.ChatWithAgent(ctx, c.config.AgentName, message)
}

// ChatWithAgent sends a message to a specific OpenClaw agent
func (c *Client) ChatWithAgent(ctx context.Context, agentName, message string) (*ChatResponse, error) {
	req := ChatRequest{
		Message:   message,
		AgentName: agentName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("openclaw: failed to marshal request: %w", err)
	}

	url := strings.TrimRight(c.config.BaseURL, "/") + "/api/v1/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openclaw: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openclaw: failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, ErrAuth
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d: %s", ErrUnavailable, resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if chatResp.Error != "" {
		return &chatResp, fmt.Errorf("openclaw: agent error: %s", chatResp.Error)
	}

	return &chatResp, nil
}

// HealthCheck verifies the OpenClaw instance is reachable
func (c *Client) HealthCheck(ctx context.Context) error {
	url := strings.TrimRight(c.config.BaseURL, "/") + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("openclaw: failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: health check returned status %d", ErrUnavailable, resp.StatusCode)
	}

	return nil
}
