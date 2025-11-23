# WebSocket Push Message Feature

## Overview

The WebSocket Push Message feature enables agents to proactively send real-time messages to chat rooms without waiting for user requests. This is useful for sending alerts, notifications, and status updates directly from the agent to the UI.

## Features

- **Real-time Communication**: Send messages instantly via WebSocket
- **Multiple Rooms**: Support sending to different chat rooms
- **Flexible Payload**: Send any JSON-serializable data as message payload
- **Connection Management**: Easy connect/disconnect with error handling
- **Thread-safe**: All operations are protected with mutexes

## Installation

The feature is built into the Agent struct and uses the standard `gorilla/websocket` library.

```bash
go get github.com/gorilla/websocket
```

## Quick Start

### 1. Initialize Agent with WebSocket Connection

```go
import "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"

// Create agent config
config := &agent.Config{
    Name:              "MyAgent",
    Version:           "1.0.0",
    OwnerAddress:      "0x1234...",
    PrivateKey:        "0x5678...",
    Capabilities:      []string{"alert", "notification"},
    TaskCheckInterval: 10,
    TaskTimeout:       30,
}

// Create agent
myAgent, err := agent.NewAgent(config, handler)
if err != nil {
    log.Fatal(err)
}
defer myAgent.Close()

// Connect to WebSocket server
wsURL := "ws://localhost:8080/ws/stream"
if err := myAgent.Initialize(wsURL); err != nil {
    log.Printf("Failed to connect: %v", err)
} else {
    log.Println("WebSocket connected!")
}
```

### 2. Send Messages

```go
ctx := context.Background()

// Send alert message
err := myAgent.PushMessage(ctx, "room123", map[string]interface{}{
    "type":    "alert",
    "title":   "Price Update",
    "message": "BTC price jumped 5%",
    "price":   65432.50,
})
if err != nil {
    log.Printf("Failed to send: %v", err)
}

// Send notification
err = myAgent.PushMessage(ctx, "room456", map[string]interface{}{
    "type":    "notification",
    "message": "Agent is online and ready",
    "status":  "active",
})
```

### 3. Close Connection

```go
if err := myAgent.CloseWebSocket(); err != nil {
    log.Printf("Error closing: %v", err)
}
```

## API Reference

### Initialize(wsURL string) error

Connects the agent to a WebSocket server.

**Parameters:**
- `wsURL` (string): WebSocket server URL (e.g., `ws://localhost:8080/ws/stream`)

**Returns:**
- `error`: Connection error if failed

**Example:**
```go
err := agent.Initialize("ws://localhost:8080/ws/stream")
if err != nil {
    // Handle connection error
}
```

### PushMessage(ctx context.Context, roomID string, payload interface{}) error

Sends a message to a chat room via WebSocket.

**Parameters:**
- `ctx` (context.Context): Context for cancellation and timeout
- `roomID` (string): Target room/channel ID
- `payload` (interface{}): Message payload (must be JSON-serializable)

**Returns:**
- `error`: Send error if failed

**Message Structure:**
```json
{
  "room_id": "room123",
  "timestamp": 1700733184,
  "payload": {
    "type": "alert",
    "message": "..."
  }
}
```

**Example:**
```go
err := agent.PushMessage(ctx, "alerts", map[string]interface{}{
    "type": "pump_alert",
    "price": 123.45,
    "change": "+5%",
})
```

### CloseWebSocket() error

Closes the WebSocket connection and releases resources.

**Returns:**
- `error`: Error if connection was not properly initialized

**Example:**
```go
defer agent.CloseWebSocket()
```

### IsWebSocketConnected() bool

Checks if WebSocket is currently connected.

**Returns:**
- `bool`: `true` if connected, `false` otherwise

**Example:**
```go
if agent.IsWebSocketConnected() {
    agent.PushMessage(ctx, "room1", payload)
}
```

## Message Payload Examples

### Pump Alert
```go
map[string]interface{}{
    "type": "pump_alert",
    "token": "SHIB",
    "price": 0.0000145,
    "change_percent": 125.5,
    "volume": 1000000,
    "timestamp": time.Now().Unix(),
}
```

### Notification
```go
map[string]interface{}{
    "type": "notification",
    "level": "info",
    "message": "Agent is now online",
    "details": "Ready to process tasks",
}
```

### Status Update
```go
map[string]interface{}{
    "type": "status",
    "status": "active",
    "tasks_processed": 42,
    "uptime_seconds": 3600,
    "cpu_usage": "45%",
}
```

### Error Report
```go
map[string]interface{}{
    "type": "error",
    "level": "warning",
    "error": "Task processing failed",
    "task_id": "task_123",
    "reason": "Network timeout",
}
```

## Error Handling

### Connection Not Initialized
```go
err := agent.PushMessage(ctx, "room1", payload)
if err != nil && err.Error() == "websocket client not initialized, call Initialize() first" {
    // Initialize connection
    agent.Initialize("ws://localhost:8080/ws/stream")
}
```

### Connection Refused
```go
err := agent.Initialize("ws://invalid:9999")
if err != nil {
    log.Printf("Connection failed: %v", err)
    // Retry or fallback to local processing
}
```

### Send Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := agent.PushMessage(ctx, "room1", payload)
if err == context.DeadlineExceeded {
    log.Println("Send timeout - server not responding")
}
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
)

func main() {
    // Setup config
    config := &agent.Config{
        Name:              "AlertAgent",
        Version:           "1.0.0",
        OwnerAddress:      "0x1234567890123456789012345678901234567890",
        PrivateKey:        "0x1234567890123456789012345678901234567890123456789012345678901234",
        Capabilities:      []string{"alerts", "notifications"},
        TaskCheckInterval: 10,
        TaskTimeout:       30,
    }

    // Create agent
    handler := &MyHandler{}
    myAgent, err := agent.NewAgent(config, handler)
    if err != nil {
        log.Fatal(err)
    }
    defer myAgent.Close()

    // Connect to WebSocket
    if err := myAgent.Initialize("ws://localhost:8080/ws/stream"); err != nil {
        log.Printf("Connection failed: %v\n", err)
        return
    }
    log.Println("Connected to WebSocket server")

    // Send messages in a loop
    ctx := context.Background()
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Check price and send alert if needed
        payload := map[string]interface{}{
            "type":  "price_alert",
            "token": "BTC",
            "price": 65000.0,
        }
        
        if err := myAgent.PushMessage(ctx, "alerts", payload); err != nil {
            log.Printf("Failed to send: %v", err)
        } else {
            log.Println("Alert sent successfully")
        }
    }
}

type MyHandler struct{}

func (h *MyHandler) ProcessTask(ctx context.Context, content string) (string, error) {
    return "processed", nil
}
```

## Testing

Run the unit tests:

```bash
go test ./pkg/agent -v -run TestPushMessage
go test ./pkg/agent -v -run TestInitialize
go test ./pkg/agent -v -run TestWebSocket
```

All tests should pass with 100% success rate.

## Troubleshooting

### WebSocket Connection Refused
- **Cause**: WebSocket server not running or wrong URL
- **Solution**: Verify server is running at the specified URL

### Message Not Sent
- **Cause**: Connection not initialized
- **Solution**: Call `Initialize()` before `PushMessage()`

### Slow Message Sending
- **Cause**: Network latency or large payloads
- **Solution**: Use context with timeout and optimize payload size

### Memory Leak
- **Cause**: Connection not closed
- **Solution**: Always call `defer agent.CloseWebSocket()` or `defer agent.Close()`

## Best Practices

1. **Always Initialize**: Check connection before sending
   ```go
   if !agent.IsWebSocketConnected() {
       agent.Initialize(wsURL)
   }
   ```

2. **Use Context Timeout**: Prevent hanging on network issues
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

3. **Handle Errors**: Don't ignore send errors
   ```go
   if err := agent.PushMessage(ctx, roomID, payload); err != nil {
       log.Printf("Failed to push: %v", err)
       // Implement retry or fallback
   }
   ```

4. **Close on Shutdown**: Properly cleanup resources
   ```go
   defer func() {
       agent.CloseWebSocket()
       agent.Close()
   }()
   ```

5. **Use Small Payloads**: Keep message size reasonable
   ```go
   // Good: minimal necessary data
   payload := map[string]interface{}{
       "price": 100.0,
       "change": "+5%",
   }
   
   // Avoid: large payloads
   payload := map[string]interface{}{
       "large_data": hugeDataStructure,
   }
   ```

## Future Enhancements

- Automatic reconnection with exponential backoff
- Message queue for offline delivery
- Compression for large payloads
- TLS/WSS support for secure connections
- Message acknowledgment/confirmation

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review unit tests for usage examples
3. Open an issue on GitHub
4. Check main SDK documentation
