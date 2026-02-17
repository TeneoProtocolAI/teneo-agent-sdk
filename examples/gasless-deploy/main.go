package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft"
	"github.com/joho/godotenv"
)

// MyAgent implements the AgentHandler interface
type MyAgent struct{}

func (a *MyAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return "hello from my gasless agent", nil
}

func main() {
	_ = godotenv.Load()

	// Step 1: Mint the agent NFT (gasless — server pays all fees)
	// On first run: mints a new NFT and returns the token ID
	// On re-runs with same agent_id: detects existing agent, skips minting
	// On JSON changes: auto-updates metadata on IPFS
	result, err := nft.Mint("gasless-agent-metadata.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("==========================================")
	fmt.Printf("  Agent ready — Token ID: %d\n", result.TokenID)
	if result.TxHash != "" {
		fmt.Printf("  Transaction: %s\n", result.TxHash)
	}
	if result.MetadataURI != "" {
		fmt.Printf("  Metadata URI: %s\n", result.MetadataURI)
	}
	fmt.Println("==========================================")

	// Step 2: Start the agent with the minted token ID
	privateKey := os.Getenv("PRIVATE_KEY")
	cfg := agent.DefaultConfig()
	cfg.Name = "Gasless Deploy Example"
	cfg.Description = "Example agent deployed via gasless minting"
	cfg.PrivateKey = privateKey

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       cfg,
		AgentHandler: &MyAgent{},
		TokenID:      result.TokenID,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
