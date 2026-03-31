# Claude Code Integration Prompt for Teneo Agent SDK

Use this prompt with Claude Code to automatically wrap your business logic in the Teneo Agent SDK.

---

## Prompt for Claude Code

```
I need you to help me wrap my existing business logic in the Teneo Agent SDK to create a production-ready agent for the Teneo network.

## About Teneo Agent SDK

The Teneo Agent SDK is a Go framework that handles:
- WebSocket communication with the Teneo network
- Ethereum wallet authentication
- Task routing and management
- Health monitoring endpoints
- Automatic reconnection
- NFT integration

The SDK is pre-configured with production endpoints:
- WebSocket: wss://backend.developer.chatroom.teneo-protocol.ai/ws
- Ethereum RPC: https://peaq.api.onfinality.io/public
- NFT Contract: 0x811FF962AcBe432344AC974c1111b70847195d3C

## SDK Architecture

The SDK requires implementing this simple interface:

```go
type AgentHandler interface {
    ProcessTask(ctx context.Context, task string) (string, error)
}
```

Optional interfaces for advanced functionality:

```go
// Initialize resources when agent starts
type AgentInitializer interface {
    Initialize(ctx context.Context, config interface{}) error
}

// Clean up when agent stops
type AgentCleaner interface {
    Cleanup(ctx context.Context) error
}

// Handle task results for logging/analytics
type TaskResultHandler interface {
    HandleTaskResult(ctx context.Context, taskID, result string) error
}

// Send multiple messages during task processing
type TaskStreamHandler interface {
    ProcessTaskWithStreaming(ctx context.Context, task string, sender types.MessageSender) error
}
```

## Your Task

1. **Analyze my business logic** - Understand what my code does and how it should integrate with the SDK
2. **Create an AgentHandler implementation** - Wrap my logic in the required interface
3. **Add proper error handling** - Ensure all errors are caught and returned appropriately
4. **Implement initialization if needed** - If my code needs setup (database connections, API clients, etc.), implement AgentInitializer
5. **Implement cleanup if needed** - If my code needs cleanup (close connections, save state, etc.), implement AgentCleaner
6. **Add configuration** - Set up proper agent configuration with appropriate name, description, and capabilities
7. **Create a main.go file** - Wire everything together with the enhanced agent
8. **Add .env.example** - Show what environment variables are needed

## Requirements

- **Use production-ready code** - No placeholders, all error handling in place
- **Follow Go best practices** - Proper error wrapping, context handling, structured logging
- **Make it maintainable** - Clear variable names, comments where needed, logical structure
- **Handle edge cases** - Empty inputs, context cancellation, network errors
- **Type safety** - Proper type conversions, validation of inputs
- **Graceful degradation** - If external services fail, return useful error messages

## Agent Metadata (JSON)

When creating the agent metadata JSON for minting, **always reference the example files** in `agent-json-examples/` for the correct format. See `agent-json-examples/README.md` for the full list.

Key rules:
- Use **snake_case** for field names: `agent_id`, `agent_type`, `nlp_fallback`
- `capabilities` must be an array of **objects** with `name` and `description` fields (NOT strings)
- `metadata_version` is **required** (currently `"2.4.0"`)
- `categories` is **required** (1-2 string items)

## Configuration Guidelines

Choose appropriate capabilities based on what my code does. Examples:
- API integration: ["api_calls", "data_retrieval", "external_service"]
- Data processing: ["data_analysis", "transformation", "aggregation"]
- Blockchain: ["smart_contracts", "blockchain_queries", "transactions"]
- AI/ML: ["predictions", "classification", "recommendations"]
- Database: ["data_storage", "queries", "crud_operations"]

## Code Structure

Create this structure:

```
my-teneo-agent/
├── main.go              # Entry point with agent setup
├── agent.go             # AgentHandler implementation
├── config.go            # Configuration and initialization (if needed)
├── .env.example         # Environment variables template
├── go.mod               # Go module file
└── README.md            # Setup and usage instructions
```

## Example Integration Pattern

For a simple API-based service:

```go
type MyServiceAgent struct {
    apiClient  *http.Client
    apiKey     string
    baseURL    string
}

func (a *MyServiceAgent) ProcessTask(ctx context.Context, task string) (string, error) {
    // 1. Parse and validate input
    // 2. Call your business logic
    // 3. Format and return response
    // 4. Handle all errors appropriately
}

func (a *MyServiceAgent) Initialize(ctx context.Context, config interface{}) error {
    // Set up HTTP client, load API keys, etc.
}

func (a *MyServiceAgent) Cleanup(ctx context.Context) error {
    // Close connections, flush buffers, etc.
}
```

## Environment Variables Template

Always create a .env.example with:

```bash
# Required
PRIVATE_KEY=              # Ethereum private key (without 0x prefix)

# Optional - NFT Configuration
NFT_TOKEN_ID=             # Leave empty to auto-mint

# Your business logic variables
API_KEY=                  # Example: Your API key
DATABASE_URL=             # Example: Database connection string
```

## Error Handling Pattern

Always use this pattern:

```go
func (a *MyAgent) ProcessTask(ctx context.Context, task string) (string, error) {
    // Check context first
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

    // Validate input
    if task == "" {
        return "", fmt.Errorf("task cannot be empty")
    }

    // Process with error wrapping
    result, err := a.doBusinessLogic(task)
    if err != nil {
        return "", fmt.Errorf("failed to process task: %w", err)
    }

    return result, nil
}
```

## Main.go Template

Use this structure:

```go
package main

import (
    "log"
    "os"

    "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not found")
    }

    // Create agent handler
    myAgent := NewMyAgent()

    // Configure
    config := agent.DefaultConfig()
    config.Name = "My Agent"
    config.Description = "Description of what this agent does"
    config.Capabilities = []string{"capability1", "capability2"}
    config.PrivateKey = os.Getenv("PRIVATE_KEY")

    // Create enhanced agent
    enhancedAgent, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
        Config:       config,
        AgentHandler: myAgent,
    })
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Run
    log.Println("Starting agent...")
    if err := enhancedAgent.Run(); err != nil {
        log.Fatalf("Agent error: %v", err)
    }
}
```

## What I Need From You

1. Read and understand my business logic code
2. Ask clarifying questions if needed about:
   - What inputs the agent should accept
   - What outputs it should return
   - What external services/APIs it uses
   - What error cases exist
   - What initialization/cleanup is needed
3. Create a complete, production-ready integration following the patterns above
4. Ensure all code is functional, tested, and documented

## My Business Logic

[User will paste their code here]

## Additional Context

[User provides any additional information about their use case, requirements, or constraints]
```

---

## Using This Prompt

1. Copy the entire prompt section above
2. Replace `[User will paste their code here]` with your actual business logic
3. Replace `[Additional Context]` with your specific requirements
4. Paste into Claude Code
5. Review and test the generated integration
