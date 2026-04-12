package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// Config represents the configuration for a Teneo agent
type Config struct {
	// Basic agent info
	AgentID      string   `json:"agent_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Image        string   `json:"image"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	ContactInfo  string   `json:"contact_info"`
	PricingModel string   `json:"pricing_model"`

	// CapabilityDetails provides full capability objects with descriptions.
	// When set, these take precedence over the string-only Capabilities field
	// during deployment and metadata serialization. If empty, Capabilities
	// strings are used with empty descriptions for backward compatibility.
	CapabilityDetails []types.Capability `json:"capability_details,omitempty"`

	// Profile metadata (optional, sent to backend during deploy/update)
	ShortDescription string `json:"short_description,omitempty"` // Brief one-line summary
	TutorialURL      string `json:"tutorial_url,omitempty"`      // YouTube/video tutorial URL

	// Interface configuration
	InterfaceType  string `json:"interface_type"`
	ResponseFormat string `json:"response_format"`

	// Network configuration
	WebSocketURL     string        `json:"websocket_url"`
	ReconnectEnabled bool          `json:"reconnect_enabled"`
	ReconnectDelay   time.Duration `json:"reconnect_delay"`
	MaxReconnects    int           `json:"max_reconnects"`
	MessageTimeout   time.Duration `json:"message_timeout"`
	PingInterval     time.Duration `json:"ping_interval"`
	HandshakeTimeout time.Duration `json:"handshake_timeout"`

	// Health monitoring
	HealthEnabled bool `json:"health_enabled"`
	HealthPort    int  `json:"health_port"`

	// Authentication
	PrivateKey   string `json:"private_key"`
	OwnerAddress string `json:"owner_address"`
	NFTTokenID   string `json:"nft_token_id"`

	// Room configuration
	Room string `json:"room"`

	// Blockchain configuration
	EthereumRPC        string `json:"ethereum_rpc"`
	NFTContractAddress string `json:"nft_contract_address"`

	// Task processing
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`
	TaskTimeout        int `json:"task_timeout"`
	TaskCheckInterval  int `json:"task_check_interval"`

	// Rate limiting
	RateLimitPerMinute int `json:"rate_limit_per_minute"` // 0 = unlimited

	// Redis cache configuration
	RedisEnabled   bool   `json:"redis_enabled"`    // Enable Redis caching
	RedisAddress   string `json:"redis_address"`    // Redis server address (e.g., "localhost:6379")
	RedisUsername  string `json:"redis_username"`   // Redis ACL username (Redis 6+, empty for legacy auth)
	RedisPassword  string `json:"redis_password"`   // Redis password (empty if no password)
	RedisDB        int    `json:"redis_db"`         // Redis database number (0-15)
	RedisKeyPrefix string `json:"redis_key_prefix"` // Prefix for all cache keys
	RedisUseTLS    bool   `json:"redis_use_tls"`    // Enable TLS/SSL (required for managed Redis)

	// Slack alerting configuration
	SlackWebhookURL           string `json:"slack_webhook_url"`            // Slack webhook URL for failure alerts
	SlackAlertThrottleSeconds int    `json:"slack_alert_throttle_seconds"` // Throttle window in seconds (default: 60)

	// Additional HTTP headers sent during WebSocket handshake
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
}

// ResolveCapabilities returns the full Capability objects for this config.
// If CapabilityDetails is set, it takes precedence. Otherwise, Capabilities
// strings are converted to Capability objects with empty descriptions.
func (c *Config) ResolveCapabilities() []types.Capability {
	if len(c.CapabilityDetails) > 0 {
		return c.CapabilityDetails
	}
	return types.CapabilitiesFromStrings(c.Capabilities)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("private key is required")
	}
	// OwnerAddress is derived from private key, so we don't require it to be set
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() error {
	if agentID := os.Getenv("AGENT_ID"); agentID != "" {
		c.AgentID = agentID
	}
	if name := os.Getenv("AGENT_NAME"); name != "" {
		c.Name = name
	}
	if desc := os.Getenv("AGENT_DESCRIPTION"); desc != "" {
		c.Description = desc
	}
	if image := os.Getenv("AGENT_IMAGE"); image != "" {
		c.Image = image
	}
	if version := os.Getenv("AGENT_VERSION"); version != "" {
		c.Version = version
	}
	if caps := os.Getenv("AGENT_CAPABILITIES"); caps != "" {
		c.Capabilities = strings.Split(caps, ",")
	}
	if contact := os.Getenv("AGENT_CONTACT"); contact != "" {
		c.ContactInfo = contact
	}
	if pricing := os.Getenv("AGENT_PRICING"); pricing != "" {
		c.PricingModel = pricing
	}
	if wsURL := os.Getenv("WEBSOCKET_URL"); wsURL != "" {
		c.WebSocketURL = wsURL
	}
	if privateKey := os.Getenv("PRIVATE_KEY"); privateKey != "" {
		c.PrivateKey = privateKey
	}
	if ownerAddr := os.Getenv("OWNER_ADDRESS"); ownerAddr != "" {
		c.OwnerAddress = ownerAddr
	}
	if nftTokenID := os.Getenv("NFT_TOKEN_ID"); nftTokenID != "" {
		c.NFTTokenID = nftTokenID
	}
	if room := os.Getenv("ROOM"); room != "" {
		c.Room = room
	}
	// Also check old env var for backwards compatibility
	if c.Room == "" {
		if roomID := os.Getenv("ROOM_ID"); roomID != "" {
			c.Room = roomID
		}
	}
	if rpc := os.Getenv("ETHEREUM_RPC"); rpc != "" {
		c.EthereumRPC = rpc
	}
	if contract := os.Getenv("NFT_CONTRACT_ADDRESS"); contract != "" {
		c.NFTContractAddress = contract
	}
	if healthPort := os.Getenv("HEALTH_PORT"); healthPort != "" {
		if port, err := strconv.Atoi(healthPort); err == nil {
			c.HealthPort = port
		}
	}
	if rateLimit := os.Getenv("RATE_LIMIT_PER_MINUTE"); rateLimit != "" {
		if limit, err := strconv.Atoi(rateLimit); err == nil {
			c.RateLimitPerMinute = limit
		}
	}
	// Redis configuration
	if redisEnabled := os.Getenv("REDIS_ENABLED"); redisEnabled != "" {
		if enabled, err := strconv.ParseBool(redisEnabled); err == nil {
			c.RedisEnabled = enabled
		}
	}
	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		c.RedisAddress = redisAddr
	}
	// Also check REDIS_URL for convenience (common env var name)
	if c.RedisAddress == "" {
		if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
			c.RedisAddress = redisURL
		}
	}
	if redisUser := os.Getenv("REDIS_USERNAME"); redisUser != "" {
		c.RedisUsername = redisUser
	}
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		c.RedisPassword = redisPass
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if db, err := strconv.Atoi(redisDB); err == nil {
			c.RedisDB = db
		}
	}
	if redisPrefix := os.Getenv("REDIS_KEY_PREFIX"); redisPrefix != "" {
		c.RedisKeyPrefix = redisPrefix
	}
	if redisTLS := os.Getenv("REDIS_USE_TLS"); redisTLS != "" {
		if useTLS, err := strconv.ParseBool(redisTLS); err == nil {
			c.RedisUseTLS = useTLS
		}
	}
	// Slack alerting configuration
	if slackURL := os.Getenv("SLACK_WEBHOOK_URL"); slackURL != "" {
		c.SlackWebhookURL = slackURL
	}
	if throttle := os.Getenv("SLACK_ALERT_THROTTLE_SECONDS"); throttle != "" {
		if secs, err := strconv.Atoi(throttle); err == nil {
			c.SlackAlertThrottleSeconds = secs
		}
	}
	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Name:               "Teneo Agent",
		Description:        "A Teneo network agent",
		Version:            "1.0.0",
		Capabilities:       []string{"general"},
		InterfaceType:      types.InterfaceTypeNaturalLanguage,
		ResponseFormat:     types.ResponseFormatJSON,
		WebSocketURL:       "wss://backend.developer.chatroom.teneo-protocol.ai/ws",
		ReconnectEnabled:   true,
		ReconnectDelay:     5 * time.Second,
		MaxReconnects:      10,
		MessageTimeout:     30 * time.Second,
		PingInterval:       30 * time.Second,
		HandshakeTimeout:   10 * time.Second,
		HealthEnabled:      true,
		HealthPort:         8080,
		EthereumRPC:        "https://peaq.api.onfinality.io/public",
		NFTContractAddress: "0x811FF962AcBe432344AC974c1111b70847195d3C",
		MaxConcurrentTasks: 5,
		TaskTimeout:        30,
		TaskCheckInterval:  10,
		RateLimitPerMinute: 0, // 0 = unlimited
		RedisEnabled:       false,
		RedisAddress:       "localhost:6379",
		RedisUsername:      "", // Empty for legacy auth or default user
		RedisPassword:      "",
		RedisDB:            0,
		RedisKeyPrefix:     "", // Will be set to "teneo:agent:<agent_name>:" if empty
		RedisUseTLS:        false,
	}
}
