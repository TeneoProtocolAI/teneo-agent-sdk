package types

import (
	"context"
	"time"
)

// AgentHandler defines the interface that all Teneo agents must implement
type AgentHandler interface {
	// ProcessTask processes a single task and returns the result
	ProcessTask(ctx context.Context, task string) (string, error)
}

// AgentInitializer is an optional interface for agents that need custom initialization
type AgentInitializer interface {
	Initialize(ctx context.Context, config interface{}) error
}

// TaskProvider is an optional interface for agents that provide their own tasks
type TaskProvider interface {
	GetAvailableTasks(ctx context.Context) ([]Task, error)
}

// TaskResultHandler is an optional interface for agents that need custom result handling
type TaskResultHandler interface {
	HandleTaskResult(ctx context.Context, taskID string, result string) error
}

// AgentCleaner is an optional interface for agents that need custom cleanup
type AgentCleaner interface {
	Cleanup(ctx context.Context) error
}

// TxResultHandler is an optional interface for agents that need to handle
// transaction results from wallet interactions (e.g., approve → swap flows).
// Agents that call TriggerWalletTx should implement this to receive the user's
// tx_result response with status updates (broadcasted, confirmed, rejected, failed).
type TxResultHandler interface {
	HandleTxResult(ctx context.Context, result TxResultData, room string, sender MessageSender) error
}

// SignatureResultHandler is an optional interface for agents that need to handle
// off-chain signature results from wallet interactions (EIP-712 / personal_sign).
// Agents that call TriggerWalletSignature should implement this to receive the
// user's signature_result response. Forwarding the signature to any external
// endpoint is the agent's responsibility, not the SDK's.
type SignatureResultHandler interface {
	HandleSignatureResult(ctx context.Context, result SignatureResultData, room string, sender MessageSender) error
}

// MessageSender interface allows agents to send messages during task execution
type MessageSender interface {
	// SendMessage sends a message with content (backward compatibility - STRING type)
	SendMessage(content string) error
	// SendTaskUpdate sends a progress update for the current task
	SendTaskUpdate(content string) error
	// SendMessageAsJSON sends structured JSON data
	SendMessageAsJSON(content interface{}) error
	// SendMessageAsMD sends markdown formatted text
	SendMessageAsMD(content string) error
	// SendMessageAsArray sends array/list data
	SendMessageAsArray(content []interface{}) error
	// SendErrorMessage sends an error message to the user without triggering a transaction
	SendErrorMessage(content string, errorCode string, details map[string]interface{}) error
	// TriggerWalletTx requests the user to sign a wallet transaction
	TriggerWalletTx(tx TxRequest, description string, optional bool) error
	// TriggerWalletSignature requests an off-chain signature from the user's wallet
	// (EIP-712 typed data or personal_sign). Unlike TriggerWalletTx, there is no
	// "optional" flag — a missing signature cannot be substituted with anything
	// else, so the flow cannot advance without it.
	TriggerWalletSignature(req SignatureRequest, description string) error
	// GetRequesterWalletAddress returns the wallet address of the user who initiated the task.
	// Used for operations that must route funds to the requester (e.g. swap output).
	// Returns empty string if the requester is unknown (e.g. task from coordinator).
	GetRequesterWalletAddress() string
	// SendChunk sends a single streaming chunk to the client.
	// Each chunk is delivered as a task_response with stream metadata (seq/final).
	// Content is appended to the accumulated response on the client side.
	SendChunk(content string) error
	// SendStreamEnd signals the end of a streaming response.
	// Must be called after the last SendChunk to finalize the stream.
	SendStreamEnd() error
}

// StreamingTaskHandler is an optional interface for agents that need to send multiple messages during task execution
type StreamingTaskHandler interface {
	// ProcessTaskWithStreaming processes a task with the ability to send multiple messages
	// The MessageSender can be used to send intermediate results, progress updates, or multiple responses
	// The room parameter provides room context for proper message routing
	ProcessTaskWithStreaming(ctx context.Context, task string, room string, sender MessageSender) error
}

// DefaultAgentHandler provides default implementations for optional interfaces
type DefaultAgentHandler struct{}

// ProcessTask must be implemented by concrete agents
func (d *DefaultAgentHandler) ProcessTask(ctx context.Context, task string) (string, error) {
	return "", ErrNotImplemented
}

// Task represents a task to be processed by an agent
type Task struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Priority    int               `json:"priority"`
	CreatedAt   time.Time         `json:"created_at"`
	Deadline    *time.Time        `json:"deadline,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RequiredCap []string          `json:"required_capabilities,omitempty"`
}

// TaskResult represents the result of a processed task
type TaskResult struct {
	TaskID    string            `json:"task_id"`
	Result    string            `json:"result"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Capabilities    []Capability      `json:"capabilities"`
	IsActive        bool              `json:"is_active"`
	IsOnline        bool              `json:"is_online"`
	TasksProcessed  int64             `json:"tasks_processed"`
	TasksSuccessful int64             `json:"tasks_successful"`
	TasksFailed     int64             `json:"tasks_failed"`
	LastSeen        time.Time         `json:"last_seen"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	NFTTokenID      string            `json:"nft_token_id,omitempty"`
}

// AgentMetrics represents performance metrics for an agent
type AgentMetrics struct {
	AgentID             string        `json:"agent_id"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	SuccessRate         float64       `json:"success_rate"`
	TasksPerHour        float64       `json:"tasks_per_hour"`
	ErrorRate           float64       `json:"error_rate"`
	UptimePercentage    float64       `json:"uptime_percentage"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// Capability represents an agent capability with name and description.
// This is the standard capability format used across the Teneo ecosystem.
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AgentCapability represents a capability that an agent can perform
// Deprecated: Use Capability instead for consistency with the rest of the ecosystem.
type AgentCapability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Required    bool   `json:"required"`
}

// NetworkConfig represents configuration for connecting to the Teneo network
type NetworkConfig struct {
	WebSocketURL      string `json:"websocket_url"`
	HTTPEndpoint      string `json:"http_endpoint"`
	AuthToken         string `json:"auth_token"`
	ReconnectDelay    int    `json:"reconnect_delay"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
}

// AgentConfig represents the configuration for an agent
type AgentConfig struct {
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	Version            string            `json:"version"`
	Capabilities       []Capability      `json:"capabilities"`
	ContactInfo        string            `json:"contact_info"`
	PricingModel       string            `json:"pricing_model"`
	InterfaceType      string            `json:"interface_type"`
	ResponseFormat     string            `json:"response_format"`
	MaxConcurrentTasks int               `json:"max_concurrent_tasks"`
	TaskTimeout        int               `json:"task_timeout"`
	TaskCheckInterval  int               `json:"task_check_interval"`
	LogLevel           string            `json:"log_level"`
	Network            NetworkConfig     `json:"network"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	NamingRules        *AgentNamingRules `json:"naming_rules,omitempty"`
}

// AgentNamingRules defines naming conventions for agents
type AgentNamingRules struct {
	MaxLength        int      `json:"max_length"`
	MinLength        int      `json:"min_length"`
	CaseSensitive    bool     `json:"case_sensitive"`
	AllowNumbers     bool     `json:"allow_numbers"`
	AllowHyphens     bool     `json:"allow_hyphens"`
	AllowUnderscores bool     `json:"allow_underscores"`
	RequiredPrefix   string   `json:"required_prefix,omitempty"`
	RequiredSuffix   string   `json:"required_suffix,omitempty"`
	ReservedNames    []string `json:"reserved_names,omitempty"`
}

// AgentNameValidation represents the result of agent name validation
type AgentNameValidation struct {
	IsValid        bool     `json:"is_valid"`
	NormalizedName string   `json:"normalized_name"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors,omitempty"`
}

// Constants for interface types
const (
	InterfaceTypeNaturalLanguage = "Natural Language"
	InterfaceTypeAPI             = "API"
	InterfaceTypeCLI             = "CLI"
	InterfaceTypeWebSocket       = "WebSocket"
	InterfaceTypeGRPC            = "gRPC"
)

// Constants for response formats
const (
	ResponseFormatJSON       = "JSON"
	ResponseFormatText       = "Text"
	ResponseFormatStructured = "Structured"
	ResponseFormatXML        = "XML"
	ResponseFormatYAML       = "YAML"
)

// Constants for standardized message types
const (
	StandardMessageTypeJSON   = "JSON"
	StandardMessageTypeString = "STRING"
	StandardMessageTypeArray  = "ARRAY"
	StandardMessageTypeMD     = "MD"
)

// StandardizedMessage represents the standardized format for all agent messages
type StandardizedMessage struct {
	ContentType string      `json:"content_type"` // JSON|STRING|ARRAY|MD
	Content     interface{} `json:"content"`      // actual content based on type
}

// Constants for task types
const (
	TaskTypeGeneral        = "general"
	TaskTypeWebScraping    = "web_scraping"
	TaskTypeDataAnalysis   = "data_analysis"
	TaskTypeAIInference    = "ai_inference"
	TaskTypeFileProcessing = "file_processing"
	TaskTypeAPICall        = "api_call"
	TaskTypeCustom         = "custom"
)

// Constants for log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// CapabilitiesFromStrings converts a slice of capability name strings
// to Capability objects. Useful for backward compatibility when migrating
// from []string to []Capability.
func CapabilitiesFromStrings(names []string) []Capability {
	caps := make([]Capability, len(names))
	for i, name := range names {
		caps[i] = Capability{Name: name}
	}
	return caps
}

// CapabilityNames extracts just the name strings from a slice of Capabilities.
func CapabilityNames(caps []Capability) []string {
	names := make([]string, len(caps))
	for i, cap := range caps {
		names[i] = cap.Name
	}
	return names
}

// Common capabilities
var StandardCapabilities = []string{
	"web_scraping",
	"data_analysis",
	"ai_inference",
	"file_processing",
	"api_integration",
	"natural_language_processing",
	"image_processing",
	"database_operations",
	"workflow_automation",
	"monitoring",
}
