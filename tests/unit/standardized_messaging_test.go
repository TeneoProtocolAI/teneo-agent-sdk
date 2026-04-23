package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// TestMessageSender implements MessageSender for testing
type TestMessageSender struct {
	messages []string
	taskID   string
	room     string
}

// NewTestMessageSender creates a test message sender
func NewTestMessageSender(taskID, room string) *TestMessageSender {
	return &TestMessageSender{
		messages: make([]string, 0),
		taskID:   taskID,
		room:     room,
	}
}

// SendMessage implements backward compatibility (STRING type)
func (t *TestMessageSender) SendMessage(content string) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeString, content)
}

// SendTaskUpdate implements task updates
func (t *TestMessageSender) SendTaskUpdate(content string) error {
	updateContent := fmt.Sprintf("🔄 Update: %s", content)
	return t.sendStandardizedMessage(types.StandardMessageTypeString, updateContent)
}

// SendMessageAsJSON implements JSON message sending
func (t *TestMessageSender) SendMessageAsJSON(content interface{}) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeJSON, content)
}

// SendMessageAsMD implements markdown message sending
func (t *TestMessageSender) SendMessageAsMD(content string) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeMD, content)
}

// SendMessageAsArray implements array message sending
func (t *TestMessageSender) SendMessageAsArray(content []interface{}) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeArray, content)
}

// sendStandardizedMessage handles the core standardized message logic
func (t *TestMessageSender) sendStandardizedMessage(msgType string, content interface{}) error {
	standardizedMsg := types.StandardizedMessage{
		ContentType: msgType,
		Content:     content,
	}

	// marshal the standardized message
	contentJSON, err := json.Marshal(standardizedMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal standardized message: %w", err)
	}

	// simulate sending (store for testing)
	message := fmt.Sprintf("[%s:%s] %s", t.taskID, t.room, string(contentJSON))
	t.messages = append(t.messages, message)

	return nil
}

// GetMessages returns all sent messages
func (t *TestMessageSender) GetMessages() []string {
	return t.messages
}

// SendErrorMessage implements error message sending
func (t *TestMessageSender) SendErrorMessage(content string, errorCode string, details map[string]interface{}) error {
	return t.sendStandardizedMessage(types.StandardMessageTypeString, content)
}

// TriggerWalletTx implements wallet transaction triggering (no-op in tests)
func (t *TestMessageSender) TriggerWalletTx(tx types.TxRequest, description string, optional bool) error {
	return nil
}

// TriggerWalletSignature implements off-chain signature triggering (no-op in tests)
func (t *TestMessageSender) TriggerWalletSignature(req types.SignatureRequest, description string) error {
	return nil
}

// GetRequesterWalletAddress returns empty string for tests
func (t *TestMessageSender) GetRequesterWalletAddress() string {
	return ""
}

// SendChunk implements streaming chunk sending for tests
func (t *TestMessageSender) SendChunk(content string) error {
	return nil
}

// SendStreamEnd implements stream end signaling for tests
func (t *TestMessageSender) SendStreamEnd() error {
	return nil
}

// TestAgent implements StreamingTaskHandler for testing
type TestAgent struct{}

// ProcessTask implements basic agent interface
func (a *TestAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return "Basic task completed", nil
}

// ProcessTaskWithStreaming demonstrates standardized message usage
func (a *TestAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
	// Test 1: String message (backward compatibility)
	if err := sender.SendMessage("Testing standardized message functions"); err != nil {
		return fmt.Errorf("string message test failed: %w", err)
	}

	// Test 2: JSON message
	jsonData := map[string]interface{}{
		"test_suite": "standardized_messaging",
		"timestamp":  time.Now().Format(time.RFC3339),
		"status":     "running",
		"results": map[string]interface{}{
			"total_tests": 4,
			"passed":      0,
		},
	}

	if err := sender.SendMessageAsJSON(jsonData); err != nil {
		return fmt.Errorf("JSON message test failed: %w", err)
	}

	// Test 3: Markdown message
	markdownContent := `# Test Report

## Results
- ✅ **String Messages**: Working
- ✅ **JSON Messages**: Working  
- ⏳ **Markdown Messages**: Testing...

### Next
Array message testing.`

	if err := sender.SendMessageAsMD(markdownContent); err != nil {
		return fmt.Errorf("markdown message test failed: %w", err)
	}

	// Test 4: Array message
	arrayData := []interface{}{
		map[string]interface{}{
			"test":   "string_message",
			"status": "passed",
		},
		map[string]interface{}{
			"test":   "json_message",
			"status": "passed",
		},
		map[string]interface{}{
			"test":   "markdown_message",
			"status": "passed",
		},
		"simple_string_item",
		42,
		true,
	}

	if err := sender.SendMessageAsArray(arrayData); err != nil {
		return fmt.Errorf("array message test failed: %w", err)
	}

	// Final message
	if err := sender.SendMessage("✅ All tests completed successfully!"); err != nil {
		return fmt.Errorf("final message test failed: %w", err)
	}

	return nil
}

func TestStandardizedMessageFunctions(t *testing.T) {
	log.Println("🚀 Testing Standardized Message Functions")

	// Create test components
	taskID := "test-001"
	room := "test-room"
	testSender := NewTestMessageSender(taskID, room)
	testAgent := &TestAgent{}

	// Run test
	ctx := context.Background()
	task := "test task"

	err := testAgent.ProcessTaskWithStreaming(ctx, task, room, testSender)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Validate results
	messages := testSender.GetMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Validate message types
	expectedTypes := []string{"STRING", "JSON", "MD", "ARRAY", "STRING"}
	for i, expectedType := range expectedTypes {
		if i < len(messages) {
			msg := messages[i]
			jsonStart := fmt.Sprintf("[%s:%s] ", taskID, room)
			if len(msg) > len(jsonStart) {
				jsonPart := msg[len(jsonStart):]
				var standardizedMsg types.StandardizedMessage
				if err := json.Unmarshal([]byte(jsonPart), &standardizedMsg); err == nil {
					if standardizedMsg.ContentType != expectedType {
						t.Errorf("Message %d: expected type %s, got %s", i+1, expectedType, standardizedMsg.ContentType)
					}
				} else {
					t.Errorf("Message %d: parse error: %v", i+1, err)
				}
			}
		}
	}
}

func TestMessageSenderStringMessages(t *testing.T) {
	sender := NewTestMessageSender("test", "room")

	err := sender.SendMessage("Test message")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	messages := sender.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Parse and validate
	jsonPart := messages[0][len("[test:room] "):]
	var msg types.StandardizedMessage
	err = json.Unmarshal([]byte(jsonPart), &msg)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	if msg.ContentType != types.StandardMessageTypeString {
		t.Errorf("Expected STRING type, got %s", msg.ContentType)
	}

	if msg.Content != "Test message" {
		t.Errorf("Expected 'Test message', got %v", msg.Content)
	}
}

func TestMessageSenderJSONMessages(t *testing.T) {
	sender := NewTestMessageSender("test", "room")

	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}

	err := sender.SendMessageAsJSON(testData)
	if err != nil {
		t.Fatalf("SendMessageAsJSON failed: %v", err)
	}

	messages := sender.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Parse and validate
	jsonPart := messages[0][len("[test:room] "):]
	var msg types.StandardizedMessage
	err = json.Unmarshal([]byte(jsonPart), &msg)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	if msg.ContentType != types.StandardMessageTypeJSON {
		t.Errorf("Expected JSON type, got %s", msg.ContentType)
	}

	// Validate content structure
	contentMap, ok := msg.Content.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected content to be map[string]interface{}, got %T", msg.Content)
	}

	if contentMap["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", contentMap["key1"])
	}
}

func TestMessageSenderMarkdownMessages(t *testing.T) {
	sender := NewTestMessageSender("test", "room")

	markdown := `# Test Header

This is **bold** and *italic* text.

## List
- Item 1
- Item 2`

	err := sender.SendMessageAsMD(markdown)
	if err != nil {
		t.Fatalf("SendMessageAsMD failed: %v", err)
	}

	messages := sender.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Parse and validate
	jsonPart := messages[0][len("[test:room] "):]
	var msg types.StandardizedMessage
	err = json.Unmarshal([]byte(jsonPart), &msg)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	if msg.ContentType != types.StandardMessageTypeMD {
		t.Errorf("Expected MD type, got %s", msg.ContentType)
	}

	if msg.Content != markdown {
		t.Errorf("Markdown content mismatch")
	}
}

func TestMessageSenderArrayMessages(t *testing.T) {
	sender := NewTestMessageSender("test", "room")

	arrayData := []interface{}{
		"string item",
		42,
		true,
		map[string]interface{}{"key": "value"},
		[]string{"nested", "array"},
	}

	err := sender.SendMessageAsArray(arrayData)
	if err != nil {
		t.Fatalf("SendMessageAsArray failed: %v", err)
	}

	messages := sender.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Parse and validate
	jsonPart := messages[0][len("[test:room] "):]
	var msg types.StandardizedMessage
	err = json.Unmarshal([]byte(jsonPart), &msg)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	if msg.ContentType != types.StandardMessageTypeArray {
		t.Errorf("Expected ARRAY type, got %s", msg.ContentType)
	}

	contentArray, ok := msg.Content.([]interface{})
	if !ok {
		t.Fatalf("Expected content to be []interface{}, got %T", msg.Content)
	}

	if len(contentArray) != 5 {
		t.Errorf("Expected 5 array items, got %d", len(contentArray))
	}

	if contentArray[0] != "string item" {
		t.Errorf("Expected first item to be 'string item', got %v", contentArray[0])
	}
}
