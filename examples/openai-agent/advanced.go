package main

import (
	"log"
	"os"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

// This example shows advanced configuration options for the OpenAI agent
func runAdvancedExample() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Method 2: Advanced - customize all the options
	advancedAgent, err := agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
		// Required
		PrivateKey: os.Getenv("PRIVATE_KEY"),
		OpenAIKey:  os.Getenv("OPENAI_API_KEY"),

		// Optional: Customize agent identity
		Name:        "My Custom AI Assistant",
		Description: "A specialized AI agent for customer support",

		// Optional: OpenAI settings
		Model:       "gpt-5", // or "gpt-3.5-turbo" for faster/cheaper
		Temperature: 0.8,     // Higher = more creative (0.0-2.0)
		MaxTokens:   2000,    // Maximum response length

		// Optional: Custom system prompt to define agent behavior
		SystemPrompt: `You are a professional customer support AI assistant.
Your goal is to help users with their questions in a friendly, clear, and concise manner.
Always be polite, patient, and solution-oriented.`,

		// Optional: Define agent capabilities
		Capabilities: []string{
			"customer_support",
			"technical_assistance",
			"product_information",
			"troubleshooting",
		},

		// Optional: NFT Configuration
		// Set Mint to true to create a new NFT, or provide TokenID to use existing
		Mint:    false,
		TokenID: 12345, // Use your existing NFT token ID

		// Optional: Network configuration

		WebSocketURL: "wss://backend.developer.chatroom.teneo-protocol.ai/ws",
		Room:         "support-room",
	})

	if err != nil {
		log.Fatalf("Failed to create advanced agent: %v", err)
	}

	log.Println("🚀 Starting advanced OpenAI agent with custom configuration...")

	// Run the agent
	if err := advancedAgent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}
