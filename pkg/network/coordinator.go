package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/alerting"
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
	alerter           *alerting.SlackAlerter
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
	taskID           string
	protocolHandler  *ProtocolHandler
	room             string
	requesterWallet  string
	streamSeq        int
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

// SendErrorMessage sends an error message to the user without triggering a transaction
func (s *TaskMessageSender) SendErrorMessage(content string, errorCode string, details map[string]interface{}) error {
	errorData := types.AgentErrorData{
		TaskID:    s.taskID,
		ErrorCode: errorCode,
		Details:   details,
	}

	dataBytes, err := json.Marshal(errorData)
	if err != nil {
		return fmt.Errorf("failed to marshal error data: %w", err)
	}

	msg := &types.Message{
		Type:          types.MessageTypeAgentError,
		From:          s.protocolHandler.GetWalletAddress(),
		Room:          s.room,
		DataRoom:      s.room,
		MessageRoomId: s.room,
		Content:       content,
		TaskID:        s.taskID,
		Data:          dataBytes,
		Timestamp:     time.Now(),
	}

	return s.protocolHandler.client.SendMessage(msg)
}

// TriggerWalletTx requests the user to sign a wallet transaction
func (s *TaskMessageSender) TriggerWalletTx(tx types.TxRequest, description string, optional bool) error {
	if tx.To == "" {
		return fmt.Errorf("tx.To is required")
	}
	if tx.ChainId == 0 {
		return fmt.Errorf("tx.ChainId is required")
	}
	if description == "" {
		return fmt.Errorf("description is required")
	}

	txData := types.TriggerWalletTxData{
		TaskID:      s.taskID,
		Tx:          tx,
		Description: description,
		Optional:    optional,
	}

	dataBytes, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("failed to marshal tx data: %w", err)
	}

	msg := &types.Message{
		Type:          types.MessageTypeTriggerWalletTx,
		From:          s.protocolHandler.GetWalletAddress(),
		Room:          s.room,
		DataRoom:      s.room,
		MessageRoomId: s.room,
		Content:       description,
		TaskID:        s.taskID,
		Data:          dataBytes,
		Timestamp:     time.Now(),
	}

	return s.protocolHandler.client.SendMessage(msg)
}

// TriggerWalletSignature requests an off-chain signature (EIP-712 or personal_sign)
// from the user's wallet. Forwarding the returned signature to any external
// endpoint is the agent's responsibility.
func (s *TaskMessageSender) TriggerWalletSignature(req types.SignatureRequest, description string) error {
	if description == "" {
		return fmt.Errorf("description is required")
	}
	switch req.Method {
	case types.SignMethodTypedDataV4:
		if len(req.TypedData) == 0 {
			return fmt.Errorf("typed_data is required for %s", req.Method)
		}
	case types.SignMethodPersonalSign:
		if req.Message == "" {
			return fmt.Errorf("message is required for %s", req.Method)
		}
	case "":
		return fmt.Errorf("signature method is required")
	default:
		return fmt.Errorf("unsupported signature method: %s", req.Method)
	}

	sigData := types.TriggerWalletSignatureData{
		TaskID:      s.taskID,
		Signature:   req,
		Description: description,
	}

	dataBytes, err := json.Marshal(sigData)
	if err != nil {
		return fmt.Errorf("failed to marshal signature data: %w", err)
	}

	msg := &types.Message{
		Type:          types.MessageTypeTriggerWalletSignature,
		From:          s.protocolHandler.GetWalletAddress(),
		Room:          s.room,
		DataRoom:      s.room,
		MessageRoomId: s.room,
		Content:       description,
		TaskID:        s.taskID,
		Data:          dataBytes,
		Timestamp:     time.Now(),
	}

	return s.protocolHandler.client.SendMessage(msg)
}

// sendStandardizedMessage sends a message in standardized format
func (s *TaskMessageSender) sendStandardizedMessage(msgType string, content interface{}) error {
	var contentStr string
	switch v := content.(type) {
	case string:
		contentStr = v
	default:
		marshaled, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal content: %w", err)
		}
		contentStr = string(marshaled)
	}
	return s.protocolHandler.SendTaskResponseToRoom(s.taskID, contentStr, msgType, true, "", s.room)
}

// GetRequesterWalletAddress returns the wallet address of the user who initiated the task
func (s *TaskMessageSender) GetRequesterWalletAddress() string {
	return s.requesterWallet
}

func (s *TaskMessageSender) SendChunk(content string) error {
	err := s.protocolHandler.SendStreamingTaskResponseToRoom(
		s.taskID, content, types.StandardMessageTypeString, s.room, s.streamSeq, false,
	)
	if err != nil {
		return err
	}
	s.streamSeq++
	return nil
}

func (s *TaskMessageSender) SendStreamEnd() error {
	return s.protocolHandler.SendStreamingTaskResponseToRoom(
		s.taskID, "", types.StandardMessageTypeString, s.room, s.streamSeq, true,
	)
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
	protocolHandler.client.RegisterHandler("tx_result", coordinator.HandleTxResultMessage)
	protocolHandler.client.RegisterHandler("signature_result", coordinator.HandleSignatureResultMessage)

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

// SetAlerter sets the Slack alerter for task failure notifications
func (t *TaskCoordinator) SetAlerter(a *alerting.SlackAlerter) {
	t.alerter = a
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
	if msg.Data != nil {
		log.Printf("📋 Task message data (raw): %s", string(msg.Data))
	} else {
		log.Printf("📋 Task message data: (nil)")
	}

	// Prevent feedback loops
	if t.isResponseMessage(msg.Content) {
		log.Printf("⚠️ Ignoring response message to prevent feedback loop")
		return nil
	}

	// Accept tasks from coordinator or from a user's EVM address (when coordinator forwards with from=userWallet)
	if msg.From != "coordinator" {
		if !isEVMAddress(msg.From) || msg.From == t.protocolHandler.GetWalletAddress() {
			log.Printf("⚠️ Ignoring task from non-coordinator: %s", msg.From)
			return nil
		}
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

	// Execute task in goroutine. When msg.From is user's EVM address, use it; else extract from task data.
	requester := t.extractRequesterFromTask(msg)
	go t.ExecuteTask(taskID, msg.Content, msg.Room, requester)

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

	go t.ExecuteTask(taskID, msg.Content, msg.Room, msg.From)

	return nil
}

// HandleTxResultMessage handles incoming tx_result messages from the coordinator.
// These are sent by the user's wallet after signing (or rejecting) a transaction
// that was requested via TriggerWalletTx.
func (t *TaskCoordinator) HandleTxResultMessage(msg *types.Message) error {
	// Parse tx_result data
	var resultData types.TxResultData
	if msg.Data == nil {
		log.Printf("⚠️ Received tx_result with no data payload")
		return nil
	}
	if err := json.Unmarshal(msg.Data, &resultData); err != nil {
		log.Printf("⚠️ Failed to parse tx_result data: %v", err)
		return nil
	}

	log.Printf("📋 Received tx_result for task %s: status=%s tx_hash=%s", resultData.TaskID, resultData.Status, resultData.TxHash)

	// Check if agent implements TxResultHandler
	txResultHandler, ok := t.agentHandler.(types.TxResultHandler)
	if !ok {
		log.Printf("⚠️ Agent does not implement TxResultHandler, ignoring tx_result for task %s", resultData.TaskID)
		return nil
	}

	// Execute in a goroutine to avoid blocking the message processing loop,
	// consistent with HandleIncomingTask's pattern.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		room := msg.Room

		messageSender := &TaskMessageSender{
			taskID:          resultData.TaskID,
			protocolHandler: t.protocolHandler,
			room:            room,
			requesterWallet: msg.From,
		}

		if err := txResultHandler.HandleTxResult(ctx, resultData, room, messageSender); err != nil {
			log.Printf("❌ Agent HandleTxResult failed for task %s: %v", resultData.TaskID, err)
		}
	}()
	return nil
}

// HandleSignatureResultMessage handles incoming signature_result messages from the coordinator.
// These are sent by the user's wallet after signing (or rejecting) an off-chain signature
// that was requested via TriggerWalletSignature.
func (t *TaskCoordinator) HandleSignatureResultMessage(msg *types.Message) error {
	var resultData types.SignatureResultData
	if msg.Data == nil {
		log.Printf("⚠️ Received signature_result with no data payload")
		return nil
	}
	if err := json.Unmarshal(msg.Data, &resultData); err != nil {
		log.Printf("⚠️ Failed to parse signature_result data: %v", err)
		return nil
	}

	log.Printf("📋 Received signature_result for task %s: status=%s", resultData.TaskID, resultData.Status)

	sigResultHandler, ok := t.agentHandler.(types.SignatureResultHandler)
	if !ok {
		log.Printf("⚠️ Agent does not implement SignatureResultHandler, ignoring signature_result for task %s", resultData.TaskID)
		return nil
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		room := msg.Room

		messageSender := &TaskMessageSender{
			taskID:          resultData.TaskID,
			protocolHandler: t.protocolHandler,
			room:            room,
			requesterWallet: msg.From,
		}

		if err := sigResultHandler.HandleSignatureResult(ctx, resultData, room, messageSender); err != nil {
			log.Printf("❌ Agent HandleSignatureResult failed for task %s: %v", resultData.TaskID, err)
		}
	}()
	return nil
}

// ExecuteTask executes a task using the agent handler
func (t *TaskCoordinator) ExecuteTask(taskID, content, room, requesterWallet string) {
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
			requesterWallet: requesterWallet,
		}

		// Process the task with streaming capability
		if err := streamingHandler.ProcessTaskWithStreaming(ctx, content, room, messageSender); err != nil {
			log.Printf("❌ Streaming task %s failed: %v", taskID, err)
			t.protocolHandler.SendTaskResponseToRoom(taskID, fmt.Sprintf("❌ Error: %v", err), types.StandardMessageTypeString, false, err.Error(), room)
			if t.alerter != nil {
				t.alerter.SendTaskFailure(taskID, err.Error(), "")
			}
			return
		}

		log.Printf("✅ Streaming task %s completed successfully", taskID)
		if t.alerter != nil {
			t.alerter.RecordSuccess()
		}

		// Send final completion message if needed
		// Note: The agent should send its own completion message using the MessageSender

	} else {
		log.Printf("📄 Using standard task handler for task %s", taskID)

		// Process the task using standard method
		result, err := t.agentHandler.ProcessTask(ctx, content)
		if err != nil {
			log.Printf("❌ Task %s failed: %v", taskID, err)
			t.protocolHandler.SendTaskResponseToRoom(taskID, fmt.Sprintf("❌ Error: %v", err), types.StandardMessageTypeString, false, err.Error(), room)
			if t.alerter != nil {
				t.alerter.SendTaskFailure(taskID, err.Error(), "")
			}
			return
		}

		log.Printf("✅ Task %s completed successfully", taskID)
		if t.alerter != nil {
			t.alerter.RecordSuccess()
		}

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

// extractRequesterFromTask extracts the requester wallet address from task message data.
// Used when tasks are forwarded from coordinator (msg.From is "coordinator").
func (t *TaskCoordinator) extractRequesterFromTask(msg *types.Message) string {
	// Direct user messages: msg.From is the user's wallet
	if msg.From != "" && msg.From != "system" && msg.From != "coordinator" && msg.From != t.protocolHandler.GetWalletAddress() {
		return strings.TrimSpace(msg.From)
	}

	// Coordinator-forwarded tasks: try to extract from message data
	if msg.Data == nil {
		log.Printf("⚠️ Task has no Data; cannot extract requester wallet")
		return ""
	}

	var taskData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &taskData); err != nil {
		log.Printf("⚠️ Failed to parse task Data: %v", err)
		return ""
	}

	// Top-level EVM address fields (0x + 40 hex chars)
	keys := []string{"payer_wallet", "from", "user_address", "wallet_address", "requester", "sender", "wallet", "address", "user_wallet", "evm_address"}
	for _, key := range keys {
		if v, ok := taskData[key].(string); ok {
			addr := strings.TrimSpace(v)
			if addr != "" && strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) >= 42 {
				return addr
			}
		}
	}

	// Nested: user.address, sender.address, etc.
	if user, ok := taskData["user"].(map[string]interface{}); ok {
		for _, k := range []string{"address", "wallet_address", "wallet"} {
			if v, ok := user[k].(string); ok {
				addr := strings.TrimSpace(v)
				if addr != "" && strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) >= 42 {
					return addr
				}
			}
		}
	}
	if sender, ok := taskData["sender"].(map[string]interface{}); ok {
		for _, k := range []string{"address", "wallet_address", "wallet"} {
			if v, ok := sender[k].(string); ok {
				addr := strings.TrimSpace(v)
				if addr != "" && strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) >= 42 {
					return addr
				}
			}
		}
	}

	log.Printf("⚠️ No requester wallet in task data. Keys present: %v (coordinator must include user's EVM address for swap output routing)", mapKeys(taskData))
	return ""
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func isEVMAddress(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) != 42 || !strings.HasPrefix(strings.ToLower(s), "0x") {
		return false
	}
	for _, c := range s[2:] {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
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
