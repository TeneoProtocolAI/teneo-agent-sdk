package openclaw

import "encoding/json"

// ChatRequest represents a request to OpenClaw's /api/v1/chat endpoint
type ChatRequest struct {
	Message   string `json:"message"`
	AgentName string `json:"agent_name,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// ChatResponse represents a response from OpenClaw's /api/v1/chat endpoint
type ChatResponse struct {
	Response  string          `json:"response"`
	AgentName string          `json:"agent_name,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Error     string          `json:"error,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}
