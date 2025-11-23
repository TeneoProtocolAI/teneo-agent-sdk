package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// TaskCoordinator manages task execution and coordination
type TaskCoordinator struct {
	agentHandler      types.AgentHandler
	protocolHandler   *ProtocolHandler
	activeTasksMu     sync.RWMutex
	activeTasks       map[string]*TaskExecution
	capabilities      []string
	rateLimitPerMin   int
	rateLimitMu       sync.Mutex
	requestTimestamps []time.Time
}

// TaskExecution represents an active task execution
type TaskExecution struct {
	ID        string
	StartTime time.Time
	Cancel    context.CancelFunc
	Context   context.Context
}

// TaskMessageSender implements the MessageSender interface for streaming tasks
type TaskMessageSender struct {
	taskID          string
	protocolHandler *ProtocolHandler
	room            string
}

// SendMessage sends a message with content (backward compatibility - STRING type)
func (s *TaskMessageSender) SendMessage(content string) error {
	return s.sendStandardizedMessage(types.StandardMessageTypeString, content)
}

// SendTaskUpdate sends a progress update for the current task
func (s *TaskMessageSender) SendTaskUpdate(content string) error {
	updateContent := fmt.Sprintf("🔄 Update: %s", content)
	return s.sendStandardizedMessage(types.StandardMessageTypeString, updateContent)
}

// SendMessageAsJSON sends structured JSON data
func (s *TaskMessageSender) SendMessageAsJSON(content interface{}) error {
	return s.sendStandardizedMessage(types.StandardMessageTypeJSON, content)
}

// SendMessageAsMD sends markdown formatted text
func (s *TaskMessageSender) SendMessageAsMD(content string) error {
	return s.sendStandardizedMessage(types.StandardMessageTypeMD, content)
}

// SendMessageAsArray sends array/list data
func (s *TaskMessageSender) SendMessageAsArray(content []interface{}) error {
	return s.sendStandardizedMessage(types.StandardMessageTypeArray, content)
}

// sendStandardizedMessage sends a message in standardized format
func (s *TaskMessageSender) sendStandardizedMessage(msgType string, content interface{}) error {
	return s.protocolHandler.SendTaskResponseToRoom(s.taskID, content.(string), msgType, true, "", s.room)
}

// NewTaskCoordinator creates a new task coordinator
func NewTaskCoordinator(agentHandler types.AgentHandler, protocolHandler *ProtocolHandler, capabilities []string) *TaskCoordinator {
	coordinator := &TaskCoordinator{
		agentHandler:      agentHandler,
		protocolHandler:   protocolHandler,
		activeTasks:       make(map[string]*TaskExecution),
		capabilities:      capabilities,
		rateLimitPerMin:   0, // Will be set by SetRateLimit
		requestTimestamps: make([]time.Time, 0),
	}

	// Register task handler
	protocolHandler.client.RegisterHandler("task", coordinator.HandleIncomingTask)
	protocolHandler.client.RegisterHandler("message", coordinator.HandleUserMessage)

	return coordinator
}

// SetRateLimit sets the rate limit for task processing (tasks per minute)
// Set to 0 for unlimited
func (t *TaskCoordinator) SetRateLimit(tasksPerMinute int) {
	t.rateLimitMu.Lock()
	defer t.rateLimitMu.Unlock()
	t.rateLimitPerMin = tasksPerMinute
	log.Printf("⚙️ Rate limit set to: %d tasks/minute", tasksPerMinute)
}

// checkRateLimit checks if the rate limit allows processing a new task
// Returns true if task can be processed, false if rate limit exceeded
func (t *TaskCoordinator) checkRateLimit() bool {
	t.rateLimitMu.Lock()
	defer t.rateLimitMu.Unlock()

	// No rate limit (0 = unlimited)
	if t.rateLimitPerMin == 0 {
		return true
	}

	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)

	// Remove timestamps older than 1 minute
	validTimestamps := make([]time.Time, 0)
	for _, ts := range t.requestTimestamps {
		if ts.After(oneMinuteAgo) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	t.requestTimestamps = validTimestamps

	// Check if we've exceeded the limit
	if len(t.requestTimestamps) >= t.rateLimitPerMin {
		return false
	}

	// Add current timestamp
	t.requestTimestamps = append(t.requestTimestamps, now)
	return true
}

// HandleIncomingTask handles incoming tasks from the coordinator
func (t *TaskCoordinator) HandleIncomingTask(msg *types.Message) error {
	log.Printf("📋 Received task from %s: %s", msg.From, msg.Content)

	// Prevent feedback loops
	if t.isResponseMessage(msg.Content) {
		log.Printf("⚠️ Ignoring response message to prevent feedback loop")
		return nil
	}

	// Only handle tasks from coordinator
	if msg.From != "coordinator" {
		log.Printf("⚠️ Ignoring task from non-coordinator: %s", msg.From)
		return nil
	}

	// Extract task ID
	taskID := t.extractTaskID(msg)
	if taskID == "" {
		taskID = fmt.Sprintf("task-%d", time.Now().Unix())
	}

	// Check rate limit
	if !t.checkRateLimit() {
		log.Printf("⚠️ Rate limit exceeded, rejecting task %s", taskID)
		t.protocolHandler.SendTaskResponseToRoom(
			taskID,
			"⚠️ Agent rate limit exceeded. This agent has reached its maximum request capacity. Please try again in a moment.",
			types.StandardMessageTypeString,
			false,
			"rate_limit_exceeded",
			msg.Room,
		)
		return nil
	}

	// Execute task in goroutine
	go t.ExecuteTask(taskID, msg.Content, msg.Room)

	return nil
}

// HandleUserMessage handles direct user messages
func (t *TaskCoordinator) HandleUserMessage(msg *types.Message) error {
	// Skip system messages and self messages
	if msg.From == "system" || msg.From == t.protocolHandler.walletAddr {
		return nil
	}

	log.Printf("💬 Received user message from %s: %s", msg.From, msg.Content)

	// Treat user messages as tasks
	taskID := fmt.Sprintf("user-msg-%d", time.Now().Unix())

	// Check rate limit
	if !t.checkRateLimit() {
		log.Printf("⚠️ Rate limit exceeded, rejecting message from %s", msg.From)
		t.protocolHandler.SendTaskResponseToRoom(
			taskID,
			"⚠️ Agent rate limit exceeded. This agent has reached its maximum request capacity. Please try again in a moment.",
			types.StandardMessageTypeString,
			false,
			"rate_limit_exceeded",
			msg.Room,
		)
		return nil
	}

	go t.ExecuteTask(taskID, msg.Content, msg.Room)

	return nil
}

// ExecuteTask executes a task using the agent handler
func (t *TaskCoordinator) ExecuteTask(taskID, content, room string) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Track active task
	execution := &TaskExecution{
		ID:        taskID,
		StartTime: time.Now(),
		Cancel:    cancel,
		Context:   ctx,
	}

	t.activeTasksMu.Lock()
	t.activeTasks[taskID] = execution
	t.activeTasksMu.Unlock()

	// Clean up when done
	defer func() {
		t.activeTasksMu.Lock()
		delete(t.activeTasks, taskID)
		t.activeTasksMu.Unlock()
	}()

	log.Printf("🔄 Executing task %s: %s", taskID, content)

	// Check if agent supports streaming task handling
	if streamingHandler, ok := t.agentHandler.(types.StreamingTaskHandler); ok {
		log.Printf("📡 Using streaming task handler for task %s", taskID)

		// Create message sender for this task
		messageSender := &TaskMessageSender{
			taskID:          taskID,
			protocolHandler: t.protocolHandler,
			room:            room,
		}

		// Process the task with streaming capability
		if err := streamingHandler.ProcessTaskWithStreaming(ctx, content, room, messageSender); err != nil {
			log.Printf("❌ Streaming task %s failed: %v", taskID, err)
			t.protocolHandler.SendTaskResponseToRoom(taskID, fmt.Sprintf("❌ Error: %v", err), types.StandardMessageTypeString, false, err.Error(), room)
			return
		}

		log.Printf("✅ Streaming task %s completed successfully", taskID)

		// Send final completion message if needed
		// Note: The agent should send its own completion message using the MessageSender

	} else {
		log.Printf("📄 Using standard task handler for task %s", taskID)

		// Process the task using standard method
		result, err := t.agentHandler.ProcessTask(ctx, content)
		if err != nil {
			log.Printf("❌ Task %s failed: %v", taskID, err)
			t.protocolHandler.SendTaskResponseToRoom(taskID, fmt.Sprintf("❌ Error: %v", err), types.StandardMessageTypeString, false, err.Error(), room)
			return
		}

		log.Printf("✅ Task %s completed successfully", taskID)

		// Send response
		if err := t.protocolHandler.SendTaskResponseToRoom(taskID, result, types.StandardMessageTypeString, true, "", room); err != nil {
			log.Printf("❌ Failed to send task response: %v", err)
		}
	}

	// Handle task result if handler supports it (works for both streaming and standard)
	if resultHandler, ok := t.agentHandler.(types.TaskResultHandler); ok {
		// For streaming tasks, we don't have a single result, so we pass the task content
		result := content
		if err := resultHandler.HandleTaskResult(ctx, taskID, result); err != nil {
			log.Printf("⚠️ Failed to handle task result: %v", err)
		}
	}
}

// extractTaskID extracts task ID from message data
func (t *TaskCoordinator) extractTaskID(msg *types.Message) string {
	if msg.Data == nil {
		return ""
	}

	var taskData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &taskData); err != nil {
		return ""
	}

	if id, ok := taskData["task_id"].(string); ok {
		return id
	}

	return ""
}

// isResponseMessage checks if content looks like a response to prevent feedback loops
func (t *TaskCoordinator) isResponseMessage(content string) bool {
	contentLower := strings.ToLower(content)
	responseIndicators := []string{
		"processed",
		"timeline for @",
		"search results for",
		"user profile:",
		"tweet details:",
		"error:",
		"✅",
		"❌",
		"📊",
		"📋",
		"🔍",
	}

	for _, indicator := range responseIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}

	return false
}

// GetActiveTasks returns the list of currently active tasks
func (t *TaskCoordinator) GetActiveTasks() map[string]*TaskExecution {
	t.activeTasksMu.RLock()
	defer t.activeTasksMu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*TaskExecution)
	for k, v := range t.activeTasks {
		result[k] = v
	}

	return result
}

// GetActiveTaskCount returns the number of currently active tasks
func (t *TaskCoordinator) GetActiveTaskCount() int {
	t.activeTasksMu.RLock()
	defer t.activeTasksMu.RUnlock()
	return len(t.activeTasks)
}

// CancelTask cancels a specific task
func (t *TaskCoordinator) CancelTask(taskID string) bool {
	t.activeTasksMu.Lock()
	defer t.activeTasksMu.Unlock()

	if execution, exists := t.activeTasks[taskID]; exists {
		execution.Cancel()
		delete(t.activeTasks, taskID)
		log.Printf("🛑 Cancelled task: %s", taskID)
		return true
	}

	return false
}

// CancelAllTasks cancels all active tasks
func (t *TaskCoordinator) CancelAllTasks() {
	t.activeTasksMu.Lock()
	defer t.activeTasksMu.Unlock()

	for taskID, execution := range t.activeTasks {
		execution.Cancel()
		log.Printf("🛑 Cancelled task: %s", taskID)
	}

	// Clear the map
	t.activeTasks = make(map[string]*TaskExecution)
}

// CanHandleCapability checks if the agent can handle a specific capability
func (t *TaskCoordinator) CanHandleCapability(capability string) bool {
	for _, cap := range t.capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// UpdateCapabilities updates the agent's capabilities
func (t *TaskCoordinator) UpdateCapabilities(capabilities []string) {
	t.capabilities = capabilities
	t.protocolHandler.UpdateCapabilities(capabilities)
}
