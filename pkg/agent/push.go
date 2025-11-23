package agent

import (
"context"
"fmt"
"time"

"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/ws"
)

// PushMessage sends a message to the chatroom via WebSocket
func (a *Agent) PushMessage(ctx context.Context, roomID string, payload interface{}) error {
a.mu.RLock()
defer a.mu.RUnlock()

if a.wsClient == nil {
return fmt.Errorf("websocket client not initialized, call Initialize() first")
}

// Create message wrapper
message := map[string]interface{}{
"room_id":   roomID,
"timestamp": time.Now().Unix(),
"payload":   payload,
}

return a.wsClient.Send(message)
}

// Initialize connects the agent to WebSocket server
func (a *Agent) Initialize(wsURL string) error {
a.mu.Lock()
defer a.mu.Unlock()

// Create new WebSocket client
client := ws.NewClient(wsURL)

// Connect to WebSocket
if err := client.Connect(); err != nil {
return fmt.Errorf("failed to connect websocket: %w", err)
}

a.wsClient = client
return nil
}

// CloseWebSocket closes the WebSocket connection
func (a *Agent) CloseWebSocket() error {
a.mu.Lock()
defer a.mu.Unlock()

if a.wsClient != nil {
return a.wsClient.Close()
}
return nil
}

// IsWebSocketConnected checks if WebSocket is connected
func (a *Agent) IsWebSocketConnected() bool {
a.mu.RLock()
defer a.mu.RUnlock()
return a.wsClient != nil
}
