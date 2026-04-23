package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// MockProtocolHandler simulates the real ProtocolHandler for testing
type MockProtocolHandler struct {
	messages []string
}

func (m *MockProtocolHandler) SendTaskResponseToRoom(taskID, content string, success bool, errorMsg, room string) error {
	message := fmt.Sprintf("TaskID: %s, Room: %s, Success: %t, Content: %s", taskID, room, success, content)
	m.messages = append(m.messages, message)

	// Parse and validate the standardized message format
	var standardizedMsg types.StandardizedMessage
	if err := json.Unmarshal([]byte(content), &standardizedMsg); err != nil {
		return fmt.Errorf("invalid standardized message format: %w", err)
	}

	return nil
}

func (m *MockProtocolHandler) GetMessages() []string {
	return m.messages
}

// TaskMessageSenderTest creates TaskMessageSender with mock protocol handler
type TaskMessageSenderTest struct {
	taskID       string
	room         string
	mockProtocol *MockProtocolHandler
}

func NewTaskMessageSenderTest(taskID, room string) *TaskMessageSenderTest {
	mockProtocol := &MockProtocolHandler{messages: make([]string, 0)}

	return &TaskMessageSenderTest{
		taskID:       taskID,
		room:         room,
		mockProtocol: mockProtocol,
	}
}

// Simulate TaskMessageSender methods with mock protocol
func (t *TaskMessageSenderTest) SendMessage(content string) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeString, content)
}

func (t *TaskMessageSenderTest) SendTaskUpdate(content string) error {
	updateContent := fmt.Sprintf("🔄 Update: %s", content)
	return t.sendStandardizedMessage(types.StandardMessageTypeString, updateContent)
}

func (t *TaskMessageSenderTest) SendMessageAsJSON(content interface{}) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeJSON, content)
}

func (t *TaskMessageSenderTest) SendMessageAsMD(content string) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeMD, content)
}

func (t *TaskMessageSenderTest) SendMessageAsArray(content []interface{}) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeArray, content)
}

func (t *TaskMessageSenderTest) SendErrorMessage(content string, errorCode string, details map[string]interface{}) error {
	errorContent := map[string]interface{}{
		"error":   content,
		"code":    errorCode,
		"details": details,
	}
	return t.sendStandardizedMessage(types.StandardMessageTypeJSON, errorContent)
}

func (t *TaskMessageSenderTest) TriggerWalletTx(tx types.TxRequest, description string, optional bool) error {
	// Mock implementation - just log the request
	return nil
}

func (t *TaskMessageSenderTest) TriggerWalletSignature(req types.SignatureRequest, description string) error {
	// Mock implementation - just log the request
	return nil
}

func (t *TaskMessageSenderTest) sendStandardizedMessage(msgType string, content interface{}) error {
	standardizedMsg := types.StandardizedMessage{
		ContentType: msgType,
		Content:     content,
	}

	contentJSON, err := json.Marshal(standardizedMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal standardized message: %w", err)
	}

	// This simulates what TaskMessageSender.sendStandardizedMessage does
	return t.mockProtocol.SendTaskResponseToRoom(t.taskID, string(contentJSON), true, "", t.room)
}

// IntegrationTestAgent tests with the TaskMessageSender
type IntegrationTestAgent struct{}

func (a *IntegrationTestAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return "Integration test completed", nil
}

func (a *IntegrationTestAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
	// Test the actual methods that would be called
	if err := sender.SendMessage("Integration test started"); err != nil {
		return fmt.Errorf("string message failed: %w", err)
	}

	testData := map[string]interface{}{
		"integration_test": true,
		"coordinator":      "TaskCoordinator",
		"sender":           "TaskMessageSender",
		"room":             room,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	if err := sender.SendMessageAsJSON(testData); err != nil {
		return fmt.Errorf("JSON message failed: %w", err)
	}

	markdown := `# Integration Test

## Components Tested
- ✅ TaskMessageSender
- ✅ ProtocolHandler simulation
- ✅ Standardized message format

All components working correctly.`

	if err := sender.SendMessageAsMD(markdown); err != nil {
		return fmt.Errorf("markdown message failed: %w", err)
	}

	results := []interface{}{
		map[string]interface{}{
			"component": "TaskMessageSender",
			"status":    "working",
			"test_time": time.Now().Unix(),
		},
		map[string]interface{}{
			"component": "ProtocolHandler",
			"status":    "working",
			"test_time": time.Now().Unix(),
		},
		map[string]interface{}{
			"component": "StandardizedFormat",
			"status":    "working",
			"test_time": time.Now().Unix(),
		},
	}

	if err := sender.SendMessageAsArray(results); err != nil {
		return fmt.Errorf("array message failed: %w", err)
	}

	if err := sender.SendTaskUpdate("All integration tests passed"); err != nil {
		return fmt.Errorf("task update failed: %w", err)
	}

	return nil
}

func TestTaskMessageSenderIntegration(t *testing.T) {
	// Create test components
	taskID := "integration-test-001"
	room := "integration-room"
	sender := NewTaskMessageSenderTest(taskID, room)
	agent := &IntegrationTestAgent{}

	// Run integration test
	ctx := context.Background()
	task := "integration test task"

	err := agent.ProcessTaskWithStreaming(ctx, task, room, sender)
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	// Validate results
	messages := sender.mockProtocol.GetMessages()
	expectedCount := 5 // string, json, markdown, array, update
	if len(messages) != expectedCount {
		t.Errorf("Expected %d messages, got %d", expectedCount, len(messages))
	}

	// Validate message format
	for i, msg := range messages {
		if !contains(msg, taskID) {
			t.Errorf("Message %d should contain task ID '%s'", i+1, taskID)
		}

		if !contains(msg, room) {
			t.Errorf("Message %d should contain room '%s'", i+1, room)
		}

		if !contains(msg, "Success: true") {
			t.Errorf("Message %d should indicate success", i+1)
		}
	}
}

func TestStandardizedMessageValidation(t *testing.T) {
	sender := NewTaskMessageSenderTest("test", "room")

	// Test each message type
	tests := []struct {
		name         string
		sendFunc     func() error
		expectedType string
	}{
		{
			name: "string message",
			sendFunc: func() error {
				return sender.SendMessage("test message")
			},
			expectedType: types.StandardMessageTypeString,
		},
		{
			name: "json message",
			sendFunc: func() error {
				return sender.SendMessageAsJSON(map[string]interface{}{"key": "value"})
			},
			expectedType: types.StandardMessageTypeJSON,
		},
		{
			name: "markdown message",
			sendFunc: func() error {
				return sender.SendMessageAsMD("# Test")
			},
			expectedType: types.StandardMessageTypeMD,
		},
		{
			name: "array message",
			sendFunc: func() error {
				return sender.SendMessageAsArray([]interface{}{"item1", "item2"})
			},
			expectedType: types.StandardMessageTypeArray,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous messages
			sender.mockProtocol.messages = []string{}

			err := tt.sendFunc()
			if err != nil {
				t.Fatalf("Send function failed: %v", err)
			}

			messages := sender.mockProtocol.GetMessages()
			if len(messages) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(messages))
			}

			// Extract and validate message content
			msg := messages[0]
			contentStart := "Content: "
			contentIdx := indexOf(msg, contentStart)
			if contentIdx == -1 {
				t.Fatalf("Could not find content in message: %s", msg)
			}

			content := msg[contentIdx+len(contentStart):]

			var standardizedMsg types.StandardizedMessage
			err = json.Unmarshal([]byte(content), &standardizedMsg)
			if err != nil {
				t.Fatalf("Failed to parse standardized message: %v", err)
			}

			if standardizedMsg.ContentType != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, standardizedMsg.ContentType)
			}
		})
	}
}

func TestMessageOrderAndConsistency(t *testing.T) {
	sender := NewTaskMessageSenderTest("test", "room")
	agent := &IntegrationTestAgent{}
	ctx := context.Background()

	// Run multiple times to test consistency
	for i := 0; i < 3; i++ {
		// Clear previous messages
		sender.mockProtocol.messages = []string{}

		err := agent.ProcessTaskWithStreaming(ctx, fmt.Sprintf("test-%d", i), "room", sender)
		if err != nil {
			t.Fatalf("Run %d failed: %v", i, err)
		}

		messages := sender.mockProtocol.GetMessages()
		if len(messages) != 5 {
			t.Errorf("Run %d: expected 5 messages, got %d", i, len(messages))
		}

		// Verify message order by checking types
		expectedTypes := []string{"STRING", "JSON", "MD", "ARRAY", "STRING"}
		for j, expectedType := range expectedTypes {
			if j >= len(messages) {
				continue
			}

			msg := messages[j]
			contentStart := "Content: "
			contentIdx := indexOf(msg, contentStart)
			if contentIdx == -1 {
				continue
			}

			content := msg[contentIdx+len(contentStart):]
			var standardizedMsg types.StandardizedMessage
			if err := json.Unmarshal([]byte(content), &standardizedMsg); err == nil {
				if standardizedMsg.ContentType != expectedType {
					t.Errorf("Run %d, Message %d: expected type %s, got %s", i, j+1, expectedType, standardizedMsg.ContentType)
				}
			}
		}
	}
}

// helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
