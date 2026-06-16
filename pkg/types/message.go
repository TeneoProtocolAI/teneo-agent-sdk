package types

import (
	"encoding/json"
	"errors"
	"time"
)

// Common errors
var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrInvalidConfig           = errors.New("invalid configuration")
	ErrAgentNotFound           = errors.New("agent not found")
	ErrInvalidTask             = errors.New("invalid task")
	ErrTaskTimeout             = errors.New("task timeout")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrNetworkError            = errors.New("network error")
	ErrContractError           = errors.New("contract error")
	ErrSignatureInvalid        = errors.New("invalid signature")
	ErrNFTNotFound             = errors.New("NFT not found")
	ErrAgentAlreadyRegistered  = errors.New("agent already registered")
)

// Message represents a message in the Teneo network
type Message struct {
	ID            string            `json:"id,omitempty"`
	Type          string            `json:"type"`
	From          string            `json:"from,omitempty"`
	To            string            `json:"to,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	Content       string            `json:"content,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Signature     string            `json:"signature,omitempty"`
	TaskID        string            `json:"task_id,omitempty"`
	ReplyTo       string            `json:"reply_to,omitempty"`
	Data          json.RawMessage   `json:"data,omitempty"`
	Room          string            `json:"room,omitempty"`
	DataRoom      string            `json:"dataRoom,omitempty"`      // Client expected field #1
	MessageRoomId string            `json:"messageRoomId,omitempty"` // Client expected field #2
	PublicKey     string            `json:"publicKey,omitempty"`
}

// MessageType constants
const (
	MessageTypeTask         = "task"
	MessageTypeTaskResult   = "task_result"
	MessageTypeTaskResponse = "task_response"
	MessageTypeHeartbeat    = "heartbeat"
	MessageTypeRegistration = "registration"
	MessageTypeAuth         = "auth"
	MessageTypeError        = "error"
	MessageTypeNotification = "notification"
	MessageTypeQuery        = "query"
	MessageTypeResponse     = "response"

	// Additional message types from x-agent
	MessageTypeChallenge        = "challenge"
	MessageTypeRequestChallenge = "request_challenge"
	MessageTypeAuthSuccess      = "auth_success"
	MessageTypeAuthError        = "auth_error"
	MessageTypeRegister         = "register"
	MessageTypeCapabilities     = "capabilities"
	MessageTypePing             = "ping"
	MessageTypePong             = "pong"
	MessageTypeMessage          = "message"
	MessageTypeAgentSelected    = "agent_selected"
	MessageTypeJoin             = "join"
	MessageTypeLeave            = "leave"
	MessageTypeAgents           = "agents"
	MessageTypeRooms            = "rooms"
	MessageTypeNick             = "nick"

	// Agent-initiated message types
	MessageTypeAgentError             = "agent_error"
	MessageTypeTriggerWalletTx        = "trigger_wallet_tx"
	MessageTypeTxResult               = "tx_result"
	MessageTypeTriggerWalletSignature = "trigger_wallet_signature"
	MessageTypeSignatureResult        = "signature_result"
)

// TxStatus is the lifecycle state of a wallet transaction requested via TriggerWalletTx.
// Kept as a named string type so agents can do typed comparisons and so IsTerminal()
// gives one place to check "flow is done".
type TxStatus string

const (
	TxStatusBroadcasted TxStatus = "broadcasted"
	TxStatusConfirmed   TxStatus = "confirmed"
	TxStatusRejected    TxStatus = "rejected"
	TxStatusFailed      TxStatus = "failed"
)

// IsTerminal reports whether the status represents a final state where no further
// updates will arrive. "broadcasted" is not terminal — a "confirmed" follows.
func (s TxStatus) IsTerminal() bool {
	switch s {
	case TxStatusConfirmed, TxStatusRejected, TxStatusFailed:
		return true
	default:
		return false
	}
}

// SignMethod is the off-chain signing method the wallet should use.
type SignMethod string

const (
	// SignMethodTypedDataV4 is EIP-712 typed structured data signing.
	// TypedData field must contain a valid EIP-712 payload (types, domain, primaryType, message).
	SignMethodTypedDataV4 SignMethod = "eth_signTypedData_v4"
	// SignMethodPersonalSign is the legacy personal_sign over a raw UTF-8 string.
	// Use the Message field; TypedData is ignored.
	SignMethodPersonalSign SignMethod = "personal_sign"
)

// SignatureStatus is the lifecycle state of an off-chain signature requested via
// TriggerWalletSignature. Signing is synchronous in the wallet UI so there is no
// "broadcasted" equivalent — the user either signs, rejects, or fails.
type SignatureStatus string

const (
	SignatureStatusSigned   SignatureStatus = "signed"
	SignatureStatusRejected SignatureStatus = "rejected"
	SignatureStatusFailed   SignatureStatus = "failed"
)

// IsTerminal reports whether the status is final. All signature statuses are terminal;
// the method exists for symmetry with TxStatus so generic pipeline code can treat them uniformly.
func (s SignatureStatus) IsTerminal() bool {
	switch s {
	case SignatureStatusSigned, SignatureStatusRejected, SignatureStatusFailed:
		return true
	default:
		return false
	}
}

// AuthMessage represents an authentication message
type AuthMessage struct {
	Type       string `json:"type"`
	Token      string `json:"token"`
	Address    string `json:"address"`
	Signature  string `json:"signature"`
	Message    string `json:"message"`
	UserType   string `json:"userType"`
	AgentName  string `json:"agentName,omitempty"`
	NFTTokenID string `json:"nft_token_id,omitempty"`
	SDKVersion string `json:"sdk_version,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

// ChallengeMessage represents an authentication challenge
type ChallengeMessage struct {
	Challenge string `json:"challenge"`
	Timestamp int64  `json:"timestamp"`
}

// RegistrationMessage represents an agent registration message
type RegistrationMessage struct {
	UserType          string `json:"userType"`
	NFTTokenID        string `json:"nft_token_id"`
	WalletAddress     string `json:"wallet_address"`
	Challenge         string `json:"challenge"`
	ChallengeResponse string `json:"challenge_response"`
	Room              string `json:"room,omitempty"`
}

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	AgentID     string                 `json:"agent_id"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	TasksActive int                    `json:"tasks_active"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NotificationMessage represents a notification message
type NotificationMessage struct {
	Type    string      `json:"type"`
	Title   string      `json:"title"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Urgent  bool        `json:"urgent"`
}

// QueryMessage represents a query message
type QueryMessage struct {
	QueryType  string                 `json:"query_type"`
	Parameters map[string]interface{} `json:"parameters"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// ResponseMessage represents a response message
type ResponseMessage struct {
	QueryID string      `json:"query_id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// TaskMessage represents a task message
type TaskMessage struct {
	TaskID       string            `json:"task_id"`
	Type         string            `json:"type"`
	Description  string            `json:"description"`
	Input        string            `json:"input"`
	Requirements []string          `json:"requirements"`
	Priority     int               `json:"priority"`
	Timeout      int               `json:"timeout"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TaskResultMessage represents a task result message
type TaskResultMessage struct {
	TaskID    string            `json:"task_id"`
	Success   bool              `json:"success"`
	Result    string            `json:"result"`
	Error     string            `json:"error,omitempty"`
	Duration  int64             `json:"duration"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// AgentInfo represents basic agent information
type AgentInfo struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Capabilities []Capability `json:"capabilities"`
	Room         string       `json:"room"`
	Status       string       `json:"status"`
}

// AgentSelectedMessage represents an agent selection message
type AgentSelectedMessage struct {
	AgentID      string       `json:"agent_id"`
	AgentName    string       `json:"agent_name"`
	Capabilities []Capability `json:"capabilities"`
	Reasoning    string   `json:"reasoning"`
	UserRequest  string   `json:"user_request"`
}

// Connection represents a connection to the Teneo network
type Connection struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`
	Address     string    `json:"address"`
}

// NetworkEvent represents an event on the Teneo network
type NetworkEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

// EventType constants
const (
	EventTypeAgentJoined   = "agent_joined"
	EventTypeAgentLeft     = "agent_left"
	EventTypeTaskCreated   = "task_created"
	EventTypeTaskCompleted = "task_completed"
	EventTypeTaskFailed    = "task_failed"
	EventTypeSystemStatus  = "system_status"
	EventTypeNetworkUpdate = "network_update"
)

// TxRequest represents a transaction for the user to sign
type TxRequest struct {
	To      string `json:"to"`
	Value   string `json:"value,omitempty"`
	Data    string `json:"data,omitempty"`
	ChainId int    `json:"chainId"`
}

// AgentErrorData is the payload for agent_error messages
type AgentErrorData struct {
	TaskID    string                 `json:"task_id"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// TriggerWalletTxData is the payload for trigger_wallet_tx messages
type TriggerWalletTxData struct {
	TaskID      string    `json:"task_id"`
	Tx          TxRequest `json:"tx"`
	Description string    `json:"description"`
	Optional    bool      `json:"optional"`
}

// TxResultData is the payload received when user responds to a trigger_wallet_tx request.
// Status is one of: "broadcasted" (tx sent, hash available), "confirmed" (on-chain receipt),
// "rejected" (user declined), or "failed" (error). Use TxStatus* constants.
type TxResultData struct {
	TaskID string   `json:"task_id"`
	TxHash string   `json:"tx_hash,omitempty"`
	Status TxStatus `json:"status"`
	Error  string   `json:"error,omitempty"`
}

// TriggerWalletSignatureData is the payload for trigger_wallet_signature messages.
// The wallet MUST render the TypedData content itself — Description is advisory only
// and not a substitute for showing the user what they're actually signing.
type TriggerWalletSignatureData struct {
	TaskID      string           `json:"task_id"`
	Signature   SignatureRequest `json:"signature"`
	Description string           `json:"description"`
}

// SignatureRequest describes an off-chain signature the agent wants from the user's wallet.
// For SignMethodTypedDataV4: populate TypedData with a valid EIP-712 payload.
// For SignMethodPersonalSign: populate Message with the raw string to sign.
//
// Note: chainId is intentionally not a top-level field. For EIP-712 it belongs inside
// TypedData.domain.chainId (authoritative); personal_sign is chain-agnostic. Having two
// sources of truth for chain caused bugs in prior iterations.
type SignatureRequest struct {
	Method    SignMethod      `json:"method"`
	TypedData json.RawMessage `json:"typed_data,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// SignatureResultData is the payload received when the user responds to a
// trigger_wallet_signature request. Signature is populated only when Status == signed.
type SignatureResultData struct {
	TaskID    string          `json:"task_id"`
	Signature string          `json:"signature,omitempty"`
	Status    SignatureStatus `json:"status"`
	Error     string          `json:"error,omitempty"`
}

// StreamMeta contains streaming metadata for chunked task responses.
// When present in a task_response's Data field, it indicates the response
// is part of a streaming sequence rather than a single atomic response.
type StreamMeta struct {
	Seq   int  `json:"seq"`
	Final bool `json:"final"`
}
