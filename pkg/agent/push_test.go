package agent

import (
"context"
"testing"
)

// TestInitialize tests the Initialize method
func TestInitialize(t *testing.T) {
tests := []struct {
name    string
wsURL   string
wantErr bool
}{
{
name:    "invalid URL",
wsURL:   "://invalid",
wantErr: true,
},
{
name:    "connection refused (expected for test)",
wsURL:   "ws://localhost:19999", // non-existent port
wantErr: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

err = agent.Initialize(tt.wsURL)
if (err != nil) != tt.wantErr {
t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
}
})
}
}

// TestPushMessageWithoutInit tests that PushMessage errors when not initialized
func TestPushMessageWithoutInit(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

ctx := context.Background()

// PushMessage without Initialize should error
err = agent.PushMessage(ctx, "room1", map[string]string{"msg": "hello"})
if err == nil {
t.Error("PushMessage() without Initialize should error")
}
}

// TestCloseWebSocket tests the CloseWebSocket method
func TestCloseWebSocket(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

// Close without initializing should not error
err = agent.CloseWebSocket()
if err != nil {
t.Errorf("CloseWebSocket() on uninitialized agent should not error: %v", err)
}

// Close again should be safe
err = agent.CloseWebSocket()
if err != nil {
t.Errorf("CloseWebSocket() second call should not error: %v", err)
}
}

// TestIsWebSocketConnected tests the IsWebSocketConnected method
func TestIsWebSocketConnected(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

// Before init, should be false
if agent.IsWebSocketConnected() {
t.Error("IsWebSocketConnected() should be false before Initialize")
}

// Try to initialize (will fail due to no server, but that's OK for this test)
_ = agent.Initialize("ws://localhost:19999")

// After failed init, should still be false
if agent.IsWebSocketConnected() {
t.Error("IsWebSocketConnected() should be false after failed Initialize")
}
}

// TestAgentCreation tests basic agent creation
func TestAgentCreation(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test", "feature1"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

if agent == nil {
t.Error("NewAgent returned nil")
}

// Check capabilities
caps := agent.GetCapabilities()
if len(caps) != 2 {
t.Errorf("GetCapabilities() got %d, want 2", len(caps))
}

// Check address
addr := agent.GetAddress()
if addr != "0x123" {
t.Errorf("GetAddress() got %s, want 0x123", addr)
}
}

// TestPushMessagePayload tests that PushMessage returns correct error when not initialized
func TestPushMessagePayload(t *testing.T) {
config := &Config{
Name:              "TestAgent",
Version:           "1.0.0",
OwnerAddress:      "0x123",
PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
Capabilities:      []string{"test"},
TaskCheckInterval: 10,
TaskTimeout:       30,
}

handler := &mockHandler{}
agent, err := NewAgent(config, handler)
if err != nil {
t.Fatalf("NewAgent failed: %v", err)
}

ctx := context.Background()

// Test that PushMessage with nil wsClient returns error
err = agent.PushMessage(ctx, "room1", map[string]interface{}{
"type": "alert",
"data": "test_data",
})

if err == nil {
t.Error("PushMessage with nil wsClient should error")
}

if err.Error() != "websocket client not initialized, call Initialize() first" {
t.Errorf("PushMessage error message incorrect: %v", err)
}
}

// Mock implementations for testing

type mockHandler struct{}

func (m *mockHandler) ProcessTask(ctx context.Context, content string) (string, error) {
return "mock result", nil
}
