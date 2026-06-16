package openclaw

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// OpenClawAgent is a ready-to-use AgentHandler that bridges Teneo tasks to an OpenClaw instance.
// It implements both AgentHandler and StreamingTaskHandler interfaces.
type OpenClawAgent struct {
	client *Client
	config *Config
}

// Compile-time interface checks
var _ types.AgentHandler = (*OpenClawAgent)(nil)
var _ types.StreamingTaskHandler = (*OpenClawAgent)(nil)

// NewOpenClawAgent creates a new OpenClaw agent handler.
// It initializes the HTTP client and validates the configuration.
func NewOpenClawAgent(config *Config) (*OpenClawAgent, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("openclaw agent: %w", err)
	}

	return &OpenClawAgent{
		client: client,
		config: config,
	}, nil
}

// ProcessTask receives a task from the Teneo network and forwards it to OpenClaw.
// Local commands (help, status, health) are handled without calling OpenClaw.
func (a *OpenClawAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	task = strings.TrimSpace(task)
	if task == "" {
		return "No task provided. Send a message and it will be forwarded to OpenClaw.", nil
	}

	taskLower := strings.ToLower(task)

	// Handle local commands
	switch {
	case taskLower == "help" || taskLower == "commands":
		return a.helpMessage(), nil
	case taskLower == "status" || taskLower == "health":
		return a.statusCheck(ctx)
	}

	log.Printf("[openclaw] Forwarding task to OpenClaw: %s", task)

	resp, err := a.client.Chat(ctx, task)
	if err != nil {
		return "", fmt.Errorf("openclaw agent: failed to process task: %w", err)
	}

	return resp.Response, nil
}

// ProcessTaskWithStreaming forwards the task to OpenClaw with progress updates sent back to the Teneo room.
func (a *OpenClawAgent) ProcessTaskWithStreaming(ctx context.Context, task string, room string, sender types.MessageSender) error {
	task = strings.TrimSpace(task)
	if task == "" {
		return sender.SendMessage("No task provided. Send a message and it will be forwarded to OpenClaw.")
	}

	taskLower := strings.ToLower(task)

	// Handle local commands
	switch {
	case taskLower == "help" || taskLower == "commands":
		return sender.SendMessage(a.helpMessage())
	case taskLower == "status" || taskLower == "health":
		result, err := a.statusCheck(ctx)
		if err != nil {
			return sender.SendMessage(fmt.Sprintf("Health check failed: %v", err))
		}
		return sender.SendMessage(result)
	}

	// Send progress update
	agentLabel := a.config.AgentName
	if agentLabel == "" {
		agentLabel = "default agent"
	}
	if err := sender.SendTaskUpdate(fmt.Sprintf("Forwarding to OpenClaw (%s)...", agentLabel)); err != nil {
		log.Printf("[openclaw] Failed to send progress update: %v", err)
	}

	log.Printf("[openclaw] Forwarding task to OpenClaw: %s", task)

	resp, err := a.client.Chat(ctx, task)
	if err != nil {
		return fmt.Errorf("openclaw agent: failed to process task: %w", err)
	}

	return sender.SendMessage(resp.Response)
}

func (a *OpenClawAgent) helpMessage() string {
	agent := a.config.AgentName
	if agent == "" {
		agent = "(default)"
	}
	return fmt.Sprintf(
		"OpenClaw Agent Bridge\n"+
			"=====================\n"+
			"OpenClaw URL: %s\n"+
			"Target Agent: %s\n\n"+
			"Commands:\n"+
			"  help     - Show this message\n"+
			"  status   - Check OpenClaw connectivity\n\n"+
			"All other messages are forwarded to the OpenClaw agent for processing.",
		a.config.BaseURL, agent,
	)
}

func (a *OpenClawAgent) statusCheck(ctx context.Context) (string, error) {
	err := a.client.HealthCheck(ctx)
	if err != nil {
		return fmt.Sprintf("OpenClaw status: UNREACHABLE (%v)", err), nil
	}
	agent := a.config.AgentName
	if agent == "" {
		agent = "(default)"
	}
	return fmt.Sprintf("OpenClaw status: CONNECTED\nURL: %s\nAgent: %s", a.config.BaseURL, agent), nil
}
