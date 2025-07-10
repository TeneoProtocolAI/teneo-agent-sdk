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
	Capabilities    []string          `json:"capabilities"`
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

// AgentCapability represents a capability that an agent can perform
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
	Capabilities       []string          `json:"capabilities"`
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
