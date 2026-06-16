package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// TestStreamingAgent simulates an agent that streams chunks
type TestStreamingAgent struct{}

func (a *TestStreamingAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return "", nil
}

func (a *TestStreamingAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
	chunks := []string{"Hello", " world", "!"}
	for _, chunk := range chunks {
		if err := sender.SendChunk(chunk); err != nil {
			return err
		}
	}
	return sender.SendStreamEnd()
}

func TestSendChunkSequencing(t *testing.T) {
	agent := &TestStreamingAgent{}
	sender := &streamMockSender{}

	err := agent.ProcessTaskWithStreaming(context.Background(), "test", "room-1", sender)
	if err != nil {
		t.Fatalf("ProcessTaskWithStreaming failed: %v", err)
	}

	if len(sender.chunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(sender.chunks))
	}
	if sender.streamEnded != true {
		t.Error("Expected stream to be ended")
	}
}

func TestStreamMetaMarshal(t *testing.T) {
	meta := types.StreamMeta{Seq: 5, Final: true}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal StreamMeta: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal StreamMeta: %v", err)
	}

	if int(parsed["seq"].(float64)) != 5 {
		t.Errorf("Expected seq=5, got %v", parsed["seq"])
	}
	if parsed["final"] != true {
		t.Errorf("Expected final=true, got %v", parsed["final"])
	}
}

// streamMockSender implements MessageSender for testing streaming
type streamMockSender struct {
	chunks      []string
	streamEnded bool
}

func (s *streamMockSender) SendMessage(content string) error                     { return nil }
func (s *streamMockSender) SendTaskUpdate(content string) error                  { return nil }
func (s *streamMockSender) SendMessageAsJSON(content interface{}) error          { return nil }
func (s *streamMockSender) SendMessageAsMD(content string) error                 { return nil }
func (s *streamMockSender) SendMessageAsArray(content []interface{}) error       { return nil }
func (s *streamMockSender) SendErrorMessage(content string, errorCode string, details map[string]interface{}) error {
	return nil
}
func (s *streamMockSender) TriggerWalletTx(tx types.TxRequest, description string, optional bool) error {
	return nil
}
func (s *streamMockSender) TriggerWalletSignature(req types.SignatureRequest, description string) error {
	return nil
}
func (s *streamMockSender) GetRequesterWalletAddress() string { return "" }

func (s *streamMockSender) SendChunk(content string) error {
	s.chunks = append(s.chunks, content)
	return nil
}

func (s *streamMockSender) SendStreamEnd() error {
	s.streamEnded = true
	return nil
}
