package openclaw

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for connecting to an OpenClaw instance
type Config struct {
	// BaseURL is the OpenClaw instance URL (default: http://localhost:3000)
	BaseURL string `json:"base_url"`

	// APIToken is the Bearer token for OpenClaw authentication
	APIToken string `json:"api_token"`

	// AgentName is the target OpenClaw agent to route tasks to
	AgentName string `json:"agent_name"`

	// Timeout is the HTTP request timeout (default: 120s)
	Timeout time.Duration `json:"timeout"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:3000",
		Timeout: 120 * time.Second,
	}
}

// LoadFromEnv populates config fields from environment variables
func (c *Config) LoadFromEnv() {
	if url := os.Getenv("OPENCLAW_URL"); url != "" {
		c.BaseURL = url
	}
	if token := os.Getenv("OPENCLAW_API_TOKEN"); token != "" {
		c.APIToken = token
	}
	if name := os.Getenv("OPENCLAW_AGENT_NAME"); name != "" {
		c.AgentName = name
	}
	if timeoutStr := os.Getenv("OPENCLAW_TIMEOUT"); timeoutStr != "" {
		if secs, err := strconv.Atoi(timeoutStr); err == nil {
			c.Timeout = time.Duration(secs) * time.Second
		}
	}
}

// Validate checks that required configuration fields are set
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("openclaw: base URL is required")
	}
	if c.APIToken == "" {
		return fmt.Errorf("openclaw: API token is required")
	}
	return nil
}
