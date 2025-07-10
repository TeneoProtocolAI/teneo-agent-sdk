package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/teneo/agent-sdk-go/pkg/types"
)

// TaskCoordinator manages task execution and coordination
type TaskCoordinator struct {
	agentHandler    types.AgentHandler
	protocolHandler *ProtocolHandler
	activeTasksMu   sync.RWMutex
	activeTasks     map[string]*TaskExecution
	capabilities    []string
}

// TaskExecution represents an active task execution
type TaskExecution struct {
	ID        string
	StartTime time.Time
	Cancel    context.CancelFunc
	Context   context.Context
}

// NewTaskCoordinator creates a new task coordinator
func NewTaskCoordinator(agentHandler types.AgentHandler, protocolHandler *ProtocolHandler, capabilities []string) *TaskCoordinator {
	coordinator := &TaskCoordinator{
		agentHandler:    agentHandler,
		protocolHandler: protocolHandler,
		activeTasks:     make(map[string]*TaskExecution),
		capabilities:    capabilities,
	}

	// Register task handler
	protocolHandler.client.RegisterHandler("task", coordinator.HandleIncomingTask)
	protocolHandler.client.RegisterHandler("message", coordinator.HandleUserMessage)

	return coordinator
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

	// Execute task in goroutine
	go t.ExecuteTask(taskID, msg.Content)

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
	go t.ExecuteTask(taskID, msg.Content)

	return nil
}

// ExecuteTask executes a task using the agent handler
func (t *TaskCoordinator) ExecuteTask(taskID, content string) {
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

	// Process the task
	result, err := t.agentHandler.ProcessTask(ctx, content)
	if err != nil {
		log.Printf("❌ Task %s failed: %v", taskID, err)
		t.protocolHandler.SendTaskResponse(taskID, fmt.Sprintf("❌ Error: %v", err), false, err.Error())
		return
	}

	log.Printf("✅ Task %s completed successfully", taskID)

	// Send response
	if err := t.protocolHandler.SendTaskResponse(taskID, result, true, ""); err != nil {
		log.Printf("❌ Failed to send task response: %v", err)
	}

	// Handle task result if handler supports it
	if resultHandler, ok := t.agentHandler.(types.TaskResultHandler); ok {
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
