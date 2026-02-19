package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// PredictionMarketAgent is a commandless agent that receives raw user prompts.
// Commandless agents have no predefined commands — they accept freeform input
// and decide what to do autonomously (e.g. browse markets, place predictions,
// create new markets on a prediction platform).
type PredictionMarketAgent struct{}

func (a *PredictionMarketAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	log.Printf("[PredictionMarketAgent] Received prompt: %q", task)
	// Your agent logic here: parse the prompt, call external APIs, place predictions, etc.
	return fmt.Sprintf("Processed prediction request: %s", task), nil
}

func (a *PredictionMarketAgent) ProcessTaskWithStreaming(ctx context.Context, task string, room string, sender types.MessageSender) error {
	log.Printf("[PredictionMarketAgent] Streaming task in room %s: %q", room, task)
	sender.SendTaskUpdate("Analyzing markets...")
	time.Sleep(500 * time.Millisecond)
	return sender.SendMessage(fmt.Sprintf("Processed prediction request: %s", task))
}

func main() {
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY env var is required")
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "https://dev-rooms-websocket-ai-core-o9fmb.ondigitalocean.app"
	}

	wsURL := os.Getenv("WEBSOCKET_URL")
	if wsURL == "" {
		ws := strings.Replace(backendURL, "https://", "wss://", 1)
		ws = strings.Replace(ws, "http://", "ws://", 1)
		wsURL = ws + "/ws"
	}

	log.Printf("=== Commandless Agent Example ===")
	log.Printf("Backend:   %s", backendURL)
	log.Printf("WebSocket: %s", wsURL)

	// Step 1: Deploy commandless agent
	capabilities, _ := json.Marshal([]map[string]string{
		{"name": "platform-interaction", "description": "Registers and interacts with external platforms on behalf of the user"},
		{"name": "analysis", "description": "Analyzes data and provides insights"},
	})
	categories, _ := json.Marshal([]string{"Automation"})

	deployCfg := deploy.DeployConfig{
		BackendURL:      backendURL,
		PrivateKey:      privateKey,
		AgentID:         "commandless-example-agent",
		AgentName:       "Commandless Example Agent",
		Description:     "Autonomous agent that interacts with external platforms via freeform prompts. No predefined commands — the agent interprets user intent and acts independently.",
		AgentType:       "commandless",
		Capabilities:    capabilities,
		Categories:      categories,
		Commands:        json.RawMessage("[]"),
		NlpFallback:     false,
		MetadataVersion: "2.4.0",
	}

	log.Println("Deploying commandless agent...")
	result, err := deploy.DeployAgent(deployCfg)
	if err != nil {
		log.Fatalf("Deploy failed: %v", err)
	}

	log.Printf("Deploy success! Token ID: %d", result.TokenID)
	if result.TxHash != "" {
		log.Printf("Tx: %s", result.TxHash)
	}

	// Step 2: Connect via WebSocket
	cfg := agent.DefaultConfig()
	cfg.Name = "Commandless Example Agent"
	cfg.Description = "Autonomous agent that interacts with external platforms via freeform prompts"
	cfg.PrivateKey = privateKey
	cfg.WebSocketURL = wsURL
	cfg.Capabilities = []string{"platform-interaction", "analysis"}
	cfg.HealthEnabled = false

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       cfg,
		AgentHandler: &PredictionMarketAgent{},
		TokenID:      result.TokenID,
		BackendURL:   backendURL,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	log.Printf("Agent ready with token ID %d, connecting...", result.TokenID)

	if err := a.Run(); err != nil {
		log.Fatalf("Agent exited with error: %v", err)
	}
}
