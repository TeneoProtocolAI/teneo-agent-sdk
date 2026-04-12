package network

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/version"
)

// ProtocolHandler handles the Teneo network protocol
type ProtocolHandler struct {
	client                 *NetworkClient
	auth                   *auth.Manager
	agentName              string
	capabilities           []string
	walletAddr             string
	nftTokenID             string
	room                   string
	lastChallenge          string
	lastChallengeSignature string
	registered             chan struct{}
	registeredOnce         sync.Once
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler(client *NetworkClient, authManager *auth.Manager, agentName string, capabilities []string, walletAddr string, nftTokenID string, room string) *ProtocolHandler {
	handler := &ProtocolHandler{
		client:                 client,
		auth:                   authManager,
		agentName:              agentName,
		capabilities:           capabilities,
		walletAddr:             walletAddr,
		nftTokenID:             nftTokenID,
		room:                   room,
		lastChallenge:          "",
		lastChallengeSignature: "",
		registered:             make(chan struct{}),
	}

	// Register message handlers
	handler.registerHandlers()

	return handler
}

// registerHandlers registers all protocol message handlers
func (p *ProtocolHandler) registerHandlers() {
	p.client.RegisterHandler("challenge", p.HandleChallenge)
	p.client.RegisterHandler("auth", p.HandleAuthResponse)
	p.client.RegisterHandler("auth_success", p.HandleAuthSuccess)
	p.client.RegisterHandler("auth_error", p.HandleAuthError)
	p.client.RegisterHandler("registration_success", p.HandleRegistrationSuccess)
	p.client.RegisterHandler("error", p.HandleError)
	p.client.RegisterHandler("pong", p.HandlePong)

	// Add handlers for server acknowledgments/responses
	p.client.RegisterHandler("capabilities", p.HandleCapabilitiesResponse)
	p.client.RegisterHandler("register", p.HandleRegisterResponse)
	p.client.RegisterHandler("agents", p.HandleAgentsResponse)

	// Add task handling
	p.client.RegisterHandler("task", p.HandleTask)
}

// StartAuthentication initiates the authentication process
func (p *ProtocolHandler) StartAuthentication() error {
	log.Println("🔐 Starting authentication process...")
	// Clear any previous authentication state
	p.lastChallenge = ""
	p.lastChallengeSignature = ""
	return p.RequestChallenge()
}

// RequestChallenge requests an authentication challenge from the server
func (p *ProtocolHandler) RequestChallenge() error {
	msg := &types.Message{
		Type:      "request_challenge",
		From:      p.walletAddr,
		Room:      p.room,
		Timestamp: time.Now(),
	}

	log.Println("🔍 Requesting authentication challenge...")
	return p.client.SendMessage(msg)
}

// HandleChallenge handles incoming authentication challenges
func (p *ProtocolHandler) HandleChallenge(msg *types.Message) error {
	log.Printf("🔐 Received challenge from server")

	var challengeData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &challengeData); err != nil {
		return fmt.Errorf("failed to unmarshal challenge data: %w", err)
	}

	challenge, ok := challengeData["challenge"].(string)
	if !ok {
		return fmt.Errorf("invalid challenge format")
	}

	// Store the challenge for later use in registration
	p.lastChallenge = challenge

	return p.Authenticate(challenge)
}

// Authenticate responds to an authentication challenge
func (p *ProtocolHandler) Authenticate(challenge string) error {
	log.Printf("🔐 Signing authentication challenge...")

	// Create the message to sign
	messageToSign := fmt.Sprintf("Teneo authentication challenge: %s", challenge)

	// Sign the message
	signature, err := p.auth.SignMessage(messageToSign)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Store the signature for later use in registration
	p.lastChallengeSignature = signature

	// Create authentication message
	authData := types.AuthMessage{
		Address:    p.walletAddr,
		Message:    messageToSign,
		Signature:  signature,
		UserType:   "agent",
		AgentName:  p.agentName,
		NFTTokenID: p.nftTokenID,
		SDKVersion: version.Version(),
	}

	authDataJson, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	log.Printf("🔑 Authenticating with NFT Token ID: %s", p.nftTokenID)

	msg := &types.Message{
		Type:      "auth",
		From:      p.walletAddr,
		Room:      p.room,
		Data:      authDataJson,
		Timestamp: time.Now(),
	}

	log.Printf("📤 Sending authentication response...")
	return p.client.SendMessage(msg)
}

// HandleAuthResponse handles authentication responses
func (p *ProtocolHandler) HandleAuthResponse(msg *types.Message) error {
	if strings.Contains(msg.Content, "successful") {
		p.client.SetAuthenticated(true)
		log.Printf("✅ Authentication successful! Agent connected to Teneo network")
		return p.SendRegistration()
	} else {
		log.Printf("❌ Authentication failed: %s", msg.Content)
		p.client.SetAuthenticated(false)
	}
	return nil
}

// HandleAuthSuccess handles authentication success messages
func (p *ProtocolHandler) HandleAuthSuccess(msg *types.Message) error {
	log.Printf("✅ Authentication successful! Agent connected to Teneo network")
	p.client.SetAuthenticated(true)
	return p.SendRegistration()
}

// HandleAuthError handles authentication error messages
func (p *ProtocolHandler) HandleAuthError(msg *types.Message) error {
	log.Printf("❌ Authentication failed: %s", msg.Content)
	p.client.SetAuthenticated(false)
	return nil
}

// HandleRegistrationSuccess handles successful agent registration
func (p *ProtocolHandler) HandleRegistrationSuccess(msg *types.Message) error {
	log.Printf("✅ Agent registered successfully with capabilities: %v", p.capabilities)
	p.markRegistered()
	return nil
}

// Registered returns a channel that is closed when the agent has successfully registered.
func (p *ProtocolHandler) Registered() <-chan struct{} {
	return p.registered
}

func (p *ProtocolHandler) markRegistered() {
	p.registeredOnce.Do(func() { close(p.registered) })
}

// HandleError handles error messages from the server
func (p *ProtocolHandler) HandleError(msg *types.Message) error {
	log.Printf("❌ Error from server: %s", msg.Content)
	return nil
}

// HandlePong handles pong responses
func (p *ProtocolHandler) HandlePong(msg *types.Message) error {
	return nil
}

// HandleCapabilitiesResponse handles capabilities responses from the server
func (p *ProtocolHandler) HandleCapabilitiesResponse(msg *types.Message) error {
	log.Printf("📋 Received capabilities response from server: %s", msg.Content)

	// Check if the response indicates success based on content
	if strings.Contains(msg.Content, "updated") || strings.Contains(msg.Content, "successful") {
		log.Printf("✅ Capabilities acknowledged by server")
		return nil
	}

	// Try to parse data if it exists and is not empty
	if len(msg.Data) > 0 {
		var capabilities map[string]interface{}
		if err := json.Unmarshal(msg.Data, &capabilities); err != nil {
			log.Printf("⚠️ Could not parse capabilities data, but response indicates success: %v", err)
			return nil // Don't fail on JSON parse errors if content indicates success
		}

		// Process capabilities if present
		if capData, ok := capabilities["capabilities"].([]interface{}); ok {
			p.UpdateCapabilities(convertInterfaceSliceToStringSlice(capData))
			log.Printf("Updated capabilities: %v", p.capabilities)
		}
	}

	return nil
}

// HandleRegisterResponse handles register responses from the server
func (p *ProtocolHandler) HandleRegisterResponse(msg *types.Message) error {
	log.Printf("📝 Received register response from server: %s", msg.Content)

	// Check if registration was successful based on content message
	if strings.Contains(msg.Content, "successful") || strings.Contains(msg.Content, "Registration successful") {
		log.Printf("✅ Agent registered successfully with server")
		p.markRegistered()
		return nil
	}

	// Try to parse data if it exists
	if len(msg.Data) > 0 {
		var responseData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &responseData); err != nil {
			log.Printf("⚠️ Could not parse registration data: %v", err)
			// If content indicates success, don't fail on JSON parse errors
			if strings.Contains(msg.Content, "successful") {
				return nil
			}
			return fmt.Errorf("failed to unmarshal register response: %w", err)
		}

		// Check for explicit success field
		if success, ok := responseData["success"].(bool); ok && success {
			log.Printf("✅ Agent registered successfully with server")
			p.markRegistered()
			return nil
		}

		// Check if this is actually a user registration confirmation (not for us)
		if userType, ok := responseData["type"].(string); ok && userType == "user" {
			log.Printf("📝 Received user registration confirmation (not for this agent)")
			return nil
		}

		log.Printf("❌ Agent registration may have failed: %v", responseData)
	}

	return nil // Don't fail, just log
}

// HandleAgentsResponse handles agents responses from the server
func (p *ProtocolHandler) HandleAgentsResponse(msg *types.Message) error {
	log.Printf("👥 Received agents response from server: %s", msg.Content)
	var agents []map[string]interface{}
	if err := json.Unmarshal(msg.Data, &agents); err != nil {
		return fmt.Errorf("failed to unmarshal agents response: %w", err)
	}
	log.Printf("Current agents on network: %v", agents)
	// TODO: Implement logic to update local agent list based on this response
	return nil
}

// HandleTask handles incoming task requests from users
func (p *ProtocolHandler) HandleTask(msg *types.Message) error {
	log.Printf("📋 Received task from %s: %s", msg.From, msg.Content)

	var taskData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &taskData); err != nil {
		log.Printf("⚠️ Could not parse task data: %v", err)
		// Use message content as task if data parsing fails
		return p.processTask(msg.From, msg.Content, "", msg.Room)
	}

	taskID, _ := taskData["task_id"].(string)
	taskContent := msg.Content

	if content, ok := taskData["content"].(string); ok && content != "" {
		taskContent = content
	}

	return p.processTask(msg.From, taskContent, taskID, msg.Room)
}

// processTask processes a task and sends a response
func (p *ProtocolHandler) processTask(from, content, taskID, room string) error {
	log.Printf("🔄 Processing task: %s", content)

	// Simple demonstration response - in a real agent this would be more sophisticated
	response := fmt.Sprintf("Hello! I'm %s, a Teneo network agent. I received your message: \"%s\"\n\nI can help with:\n- Text processing\n- Data analysis\n- Conversation\n- Demonstrations\n\nHow can I assist you further?", p.agentName, content)

	// Create response message
	responseData := map[string]interface{}{
		"type":    "task_response",
		"success": true,
	}

	if taskID != "" {
		responseData["task_id"] = taskID
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	msg := &types.Message{
		Type:      "task_response",
		From:      p.walletAddr,
		To:        from,
		Room:      room,
		Content:   response,
		Data:      data,
		Timestamp: time.Now(),
	}

	log.Printf("📤 Sending task response to %s", from)
	return p.client.SendMessage(msg)
}

// SendCapabilities sends agent capabilities to the server
func (p *ProtocolHandler) SendCapabilities() error {
	// Send capabilities in the same format as x-agent (simple JSON, not wrapped in Message)
	capMsg := map[string]interface{}{
		"type":         "capabilities",
		"capabilities": p.capabilities,
		"room":         p.room,
	}

	data, err := json.Marshal(capMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	log.Printf("📋 Sending capabilities: %v", p.capabilities)

	// Send directly via WebSocket using the new SendRawData method
	return p.client.SendRawData(data)
}

// SendPing sends a ping message to the server
func (p *ProtocolHandler) SendPing() error {
	msg := &types.Message{
		Type:      "ping",
		From:      p.walletAddr,
		Content:   "ping",
		Timestamp: time.Now(),
	}

	return p.client.SendMessage(msg)
}

// RegisterAgent registers the agent with the server
func (p *ProtocolHandler) RegisterAgent() error {
	registerData, err := json.Marshal(map[string]interface{}{
		"capabilities": p.capabilities,
		"description":  fmt.Sprintf("%s - Teneo network agent", p.agentName),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal register data: %w", err)
	}

	msg := &types.Message{
		Type:      "register",
		From:      p.walletAddr,
		Room:      p.room,
		Content:   fmt.Sprintf("%s - Teneo network agent", p.agentName),
		Data:      registerData,
		Timestamp: time.Now(),
	}

	log.Printf("📝 Registering agent: %s", p.agentName)
	return p.client.SendMessage(msg)
}

// SendTaskResponseToRoom sends a task response back to the coordinator using a specific room
func (p *ProtocolHandler) SendTaskResponseToRoom(taskID, content string, contentType string, success bool, errorMsg, room string) error {
	// Create response data for the Data field
	responseData := map[string]interface{}{
		"task_id": taskID,
		"success": success,
	}

	if errorMsg != "" {
		responseData["error"] = errorMsg
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Create message with room context fields that client expects
	msg := &types.Message{
		Type:          "task_response",
		From:          p.agentName, // Use agent name instead of wallet
		Room:          room,        // SDK internal field
		DataRoom:      room,        // Client expected field #1
		MessageRoomId: room,        // Client expected field #2
		Content:       content,
		ContentType:   contentType,
		TaskID:        taskID,
		Data:          data,
		Timestamp:     time.Now(),
	}

	// Log for debugging
	log.Printf("🐛 DEBUG: Sending task response with room context - Room: %s, TaskID: %s, Agent: %s",
		room, taskID, p.agentName)

	// Send via WebSocket with room context preserved
	return p.client.SendMessage(msg)
}

// SendStreamingTaskResponseToRoom sends a streaming chunk as a task_response.
// The stream metadata (seq/final) is included in the Data field alongside task_id.
func (p *ProtocolHandler) SendStreamingTaskResponseToRoom(taskID, content, contentType, room string, seq int, final bool) error {
	responseData := map[string]interface{}{
		"task_id": taskID,
		"success": true,
		"stream": map[string]interface{}{
			"seq":   seq,
			"final": final,
		},
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to marshal streaming response data: %w", err)
	}

	msg := &types.Message{
		Type:          "task_response",
		From:          p.agentName,
		Room:          room,
		DataRoom:      room,
		MessageRoomId: room,
		Content:       content,
		ContentType:   contentType,
		TaskID:        taskID,
		Data:          data,
		Timestamp:     time.Now(),
	}

	return p.client.SendMessage(msg)
}

// UpdateCapabilities updates the agent's capabilities
func (p *ProtocolHandler) UpdateCapabilities(capabilities []string) {
	p.capabilities = capabilities
}

// GetCapabilities returns the current capabilities
func (p *ProtocolHandler) GetCapabilities() []string {
	return p.capabilities
}

// GetWalletAddress returns the wallet address
func (p *ProtocolHandler) GetWalletAddress() string {
	return p.walletAddr
}

// SendRegistration sends agent registration with NFT token ID
func (p *ProtocolHandler) SendRegistration() error {
	registrationMsg := &types.RegistrationMessage{
		UserType:          "agent",
		NFTTokenID:        p.nftTokenID,
		WalletAddress:     p.walletAddr,
		Challenge:         p.lastChallenge,
		ChallengeResponse: p.lastChallengeSignature,
		Room:              p.room,
	}

	registrationData, err := json.Marshal(registrationMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal registration data: %w", err)
	}

	// Create message
	msg := &types.Message{
		Type:      "register",
		From:      p.walletAddr,
		Room:      p.room,
		Content:   fmt.Sprintf("Agent registration: %s", p.agentName),
		Data:      registrationData,
		Timestamp: time.Now(),
	}

	log.Printf("📝 Sending agent registration with NFT Token ID: %s", p.nftTokenID)
	return p.client.SendMessage(msg)
}

// convertInterfaceSliceToStringSlice converts a slice of interface{} to a slice of string
func convertInterfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}
