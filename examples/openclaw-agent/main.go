package main

import (
	"log"
	"os"
	"strconv"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/openclaw"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// --- OpenClaw configuration ---
	openclawConfig := openclaw.DefaultConfig()
	openclawConfig.LoadFromEnv()

	handler, err := openclaw.NewOpenClawAgent(openclawConfig)
	if err != nil {
		log.Fatalf("Failed to create OpenClaw agent handler: %v", err)
	}

	// --- SDK configuration ---
	config := agent.DefaultConfig()
	config.Name = "OpenClaw Bridge Agent"
	config.Description = "Bridges Teneo network commands to an OpenClaw instance"
	config.Version = "1.0.0"
	config.Capabilities = []string{
		"openclaw_integration",
		"ai_assistant",
	}
	config.WebSocketURL = "ws://localhost:8080/ws"
	config.HealthEnabled = true
	config.HealthPort = 8091
	config.PrivateKey = os.Getenv("PRIVATE_KEY")
	config.TaskTimeout = 120 // OpenClaw may take longer

	if config.PrivateKey == "" {
		log.Fatalf("PRIVATE_KEY environment variable is required")
	}

	if wsURL := os.Getenv("WEBSOCKET_URL"); wsURL != "" {
		config.WebSocketURL = wsURL
	}

	if port := os.Getenv("HEALTH_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.HealthPort = p
		}
	}

	// --- NFT configuration ---
	var useDeployFlow bool
	var existingTokenID uint64

	if tokenIDStr := os.Getenv("NFT_TOKEN_ID"); tokenIDStr != "" {
		if id, err := strconv.ParseUint(tokenIDStr, 10, 64); err == nil {
			existingTokenID = id
		} else {
			log.Fatalf("Invalid NFT_TOKEN_ID: %s", tokenIDStr)
		}
	} else {
		useDeployFlow = true
		log.Printf("NFT_TOKEN_ID not set, will use deploy flow")
	}

	// --- Create and run agent ---
	enhancedAgentConfig := &agent.EnhancedAgentConfig{
		Config:       config,
		AgentHandler: handler,
		Deploy:       useDeployFlow,
		TokenID:      existingTokenID,
		AgentType:    "command",
		BackendURL:   os.Getenv("BACKEND_URL"),
		RPCEndpoint:  os.Getenv("RPC_ENDPOINT"),
	}

	enhancedAgent, err := agent.NewEnhancedAgent(enhancedAgentConfig)
	if err != nil {
		log.Fatalf("Failed to create enhanced agent: %v", err)
	}

	log.Printf("Starting OpenClaw Bridge Agent")
	log.Printf("  OpenClaw URL: %s", openclawConfig.BaseURL)
	log.Printf("  OpenClaw Agent: %s", openclawConfig.AgentName)
	log.Printf("  WebSocket: %s", config.WebSocketURL)

	if err := enhancedAgent.Run(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	log.Printf("OpenClaw Bridge Agent shutdown complete")
}
