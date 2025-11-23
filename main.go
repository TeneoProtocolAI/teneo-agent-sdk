package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
)

func main() {
	// Create agent config
	// Note: This is a dummy private key for testing only!
	dummyPrivateKey := "0x1234567890123456789012345678901234567890123456789012345678901234"
	
	config := &agent.Config{
		Name:                "DemoAgent",
		Version:             "1.0.0",
		Description:         "A demo agent for testing WebSocket push messages",
		OwnerAddress:        "0x1234567890123456789012345678901234567890",
		Capabilities:        []string{"pump_detection", "alert"},
		ContactInfo:         "demo@example.com",
		PricingModel:        "free",
		InterfaceType:       "chat",
		ResponseFormat:      "json",
		EthereumRPC:         "", // optional
		NFTContractAddress:  "", // optional
		PrivateKey:          dummyPrivateKey,
		TaskCheckInterval:   10,
		TaskTimeout:         30,
	}

	// Create a dummy handler (implement types.AgentHandler interface)
	handler := &demoHandler{}

	// Create agent
	myAgent, err := agent.NewAgent(config, handler)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	defer myAgent.Close()

	// Initialize WebSocket connection
	wsURL := "ws://localhost:8080/ws/stream"
	if err := myAgent.Initialize(wsURL); err != nil {
		fmt.Printf("❌ Failed to connect WebSocket: %v\n", err)
		fmt.Printf("⚠️  Make sure WebSocket server is running at %s\n", wsURL)
	} else {
		fmt.Println("✅ WebSocket connected!")

		// Send test messages
		ctx := context.Background()

		// Test message 1
		err1 := myAgent.PushMessage(ctx, "room123", map[string]interface{}{
			"type":  "pump_alert",
			"price": 123.45,
		})
		if err1 != nil {
			fmt.Printf("❌ PushMessage failed: %v\n", err1)
		} else {
			fmt.Println("✅ Pump alert message sent successfully")
		}

		// Test message 2
		err2 := myAgent.PushMessage(ctx, "room456", map[string]interface{}{
			"type":    "notification",
			"message": "Agent is online and ready",
		})
		if err2 != nil {
			fmt.Printf("❌ PushMessage failed: %v\n", err2)
		} else {
			fmt.Println("✅ Notification message sent successfully")
		}

		// Close WebSocket connection
		if err := myAgent.CloseWebSocket(); err != nil {
			fmt.Printf("⚠️  Error closing websocket: %v\n", err)
		} else {
			fmt.Println("✅ WebSocket closed")
		}
	}
}

// demoHandler implements types.AgentHandler
type demoHandler struct{}

func (d *demoHandler) ProcessTask(ctx context.Context, content string) (string, error) {
	return "processed: " + content, nil
}
