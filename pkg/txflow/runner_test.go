package txflow

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// mockSender records TriggerWalletTx / TriggerWalletSignature calls so tests
// can assert ordering. It is concurrency-safe because the Runner may invoke
// it from callback goroutines.
type mockSender struct {
	mu         sync.Mutex
	txCalls    []types.TxRequest
	sigCalls   []types.SignatureRequest
	descs      []string
	triggerErr error
}

func (m *mockSender) SendMessage(string) error                   { return nil }
func (m *mockSender) SendTaskUpdate(string) error                { return nil }
func (m *mockSender) SendMessageAsJSON(interface{}) error        { return nil }
func (m *mockSender) SendMessageAsMD(string) error               { return nil }
func (m *mockSender) SendMessageAsArray([]interface{}) error     { return nil }
func (m *mockSender) SendChunk(string) error                     { return nil }
func (m *mockSender) SendStreamEnd() error                       { return nil }
func (m *mockSender) GetRequesterWalletAddress() string          { return "0xuser" }
func (m *mockSender) SendErrorMessage(string, string, map[string]interface{}) error {
	return nil
}

func (m *mockSender) TriggerWalletTx(tx types.TxRequest, desc string, optional bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.triggerErr != nil {
		return m.triggerErr
	}
	m.txCalls = append(m.txCalls, tx)
	m.descs = append(m.descs, desc)
	return nil
}

func (m *mockSender) TriggerWalletSignature(req types.SignatureRequest, desc string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.triggerErr != nil {
		return m.triggerErr
	}
	m.sigCalls = append(m.sigCalls, req)
	m.descs = append(m.descs, desc)
	return nil
}

func (m *mockSender) snapshotTx() []types.TxRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]types.TxRequest, len(m.txCalls))
	copy(out, m.txCalls)
	return out
}

func (m *mockSender) snapshotSig() []types.SignatureRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]types.SignatureRequest, len(m.sigCalls))
	copy(out, m.sigCalls)
	return out
}

func TestRunner_SingleTxPipeline_Completes(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}

	var completeCount int32
	p := Pipeline{
		Room:        "room1",
		Description: "single-tx",
		Steps: []Step{
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0xabc", ChainId: 1}, Description: "do it"},
		},
		OnComplete: func(ctx context.Context, s types.MessageSender) error {
			atomic.AddInt32(&completeCount, 1)
			return nil
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := sender.snapshotTx(); len(got) != 1 {
		t.Fatalf("expected 1 tx trigger, got %d", len(got))
	}

	// Broadcasted is intermediate — should NOT complete yet.
	r.HandleTxResult(context.Background(), types.TxResultData{
		TaskID: "t1", Status: types.TxStatusBroadcasted, TxHash: "0x1",
	}, "room1", sender)
	if atomic.LoadInt32(&completeCount) != 0 {
		t.Fatal("broadcasted should not complete pipeline")
	}

	r.HandleTxResult(context.Background(), types.TxResultData{
		TaskID: "t1", Status: types.TxStatusConfirmed, TxHash: "0x1",
	}, "room1", sender)
	if atomic.LoadInt32(&completeCount) != 1 {
		t.Fatalf("expected OnComplete=1, got %d", atomic.LoadInt32(&completeCount))
	}
	if r.HasPipeline("room1") {
		t.Fatal("pipeline should be GC'd after completion")
	}
}

func TestRunner_MixedTxThenSignature_RunsInOrder(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}

	var sigResultCaptured atomic.Value // *types.SignatureResultData

	p := Pipeline{
		Room: "intent-room",
		Steps: []Step{
			{
				Kind:        ActionKindTx,
				Tx:          &types.TxRequest{To: "0xapprove", ChainId: 1},
				Description: "approve",
			},
			{
				Kind: ActionKindSignature,
				Signature: &types.SignatureRequest{
					Method:    types.SignMethodTypedDataV4,
					TypedData: json.RawMessage(`{"domain":{}}`),
				},
				Description: "sign intent",
				OnResult: func(ctx context.Context, result any, s types.MessageSender) error {
					sigResultCaptured.Store(SignatureResultOf(result))
					return nil
				},
			},
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Step 1 triggered — no signature yet.
	if len(sender.snapshotTx()) != 1 || len(sender.snapshotSig()) != 0 {
		t.Fatalf("expected 1 tx / 0 sig after Run, got %d/%d", len(sender.snapshotTx()), len(sender.snapshotSig()))
	}

	// Confirm tx — should trigger sig step.
	r.HandleTxResult(context.Background(), types.TxResultData{
		TaskID: "t", Status: types.TxStatusConfirmed, TxHash: "0x1",
	}, "intent-room", sender)
	if len(sender.snapshotSig()) != 1 {
		t.Fatalf("expected 1 sig trigger after tx confirm, got %d", len(sender.snapshotSig()))
	}

	// Deliver signature — pipeline completes, OnResult fires.
	r.HandleSignatureResult(context.Background(), types.SignatureResultData{
		TaskID: "t", Status: types.SignatureStatusSigned, Signature: "0xdeadbeef",
	}, "intent-room", sender)

	got := sigResultCaptured.Load()
	if got == nil {
		t.Fatal("OnResult did not capture signature result")
	}
	sig := got.(*types.SignatureResultData)
	if sig.Signature != "0xdeadbeef" {
		t.Fatalf("unexpected signature: %q", sig.Signature)
	}
	if r.HasPipeline("intent-room") {
		t.Fatal("pipeline should be GC'd after completion")
	}
}

func TestRunner_RejectAbortsPipeline(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}

	var abortReason AbortReason
	var abortErr error
	var abortFired int32

	p := Pipeline{
		Room: "reject-room",
		Steps: []Step{
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x1", ChainId: 1}, Description: "x"},
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x2", ChainId: 1}, Description: "y"},
		},
		OnAbort: func(ctx context.Context, reason AbortReason, err error, s types.MessageSender) error {
			abortReason = reason
			abortErr = err
			atomic.StoreInt32(&abortFired, 1)
			return nil
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	r.HandleTxResult(context.Background(), types.TxResultData{
		Status: types.TxStatusRejected,
	}, "reject-room", sender)

	if atomic.LoadInt32(&abortFired) != 1 {
		t.Fatal("OnAbort did not fire")
	}
	if abortReason != AbortReasonRejected {
		t.Fatalf("expected AbortReasonRejected, got %v", abortReason)
	}
	if abortErr == nil {
		t.Fatal("expected cause error")
	}
	// Second step must not have triggered.
	if len(sender.snapshotTx()) != 1 {
		t.Fatalf("expected only 1 trigger, got %d", len(sender.snapshotTx()))
	}
}

func TestRunner_OptionalStepSurvivesReject(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}
	var completed int32

	p := Pipeline{
		Room: "optional-room",
		Steps: []Step{
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x1", ChainId: 1}, Description: "opt", Optional: true},
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x2", ChainId: 1}, Description: "req"},
		},
		OnComplete: func(ctx context.Context, s types.MessageSender) error {
			atomic.AddInt32(&completed, 1)
			return nil
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Reject the optional step — should advance to step 2.
	r.HandleTxResult(context.Background(), types.TxResultData{
		Status: types.TxStatusRejected,
	}, "optional-room", sender)
	if len(sender.snapshotTx()) != 2 {
		t.Fatalf("expected 2 triggers (advance past optional), got %d", len(sender.snapshotTx()))
	}
	r.HandleTxResult(context.Background(), types.TxResultData{
		Status: types.TxStatusConfirmed,
	}, "optional-room", sender)
	if atomic.LoadInt32(&completed) != 1 {
		t.Fatal("pipeline should complete after required step confirms")
	}
}

func TestRunner_TimeoutAborts(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}

	var abortReason AbortReason
	done := make(chan struct{})

	p := Pipeline{
		Room:    "timeout-room",
		Timeout: 50 * time.Millisecond,
		Steps: []Step{
			{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x1", ChainId: 1}, Description: "stuck"},
		},
		OnAbort: func(ctx context.Context, reason AbortReason, err error, s types.MessageSender) error {
			abortReason = reason
			close(done)
			return nil
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout did not fire")
	}
	if abortReason != AbortReasonTimeout {
		t.Fatalf("expected AbortReasonTimeout, got %v", abortReason)
	}
	if r.HasPipeline("timeout-room") {
		t.Fatal("pipeline state should be GC'd on timeout")
	}
}

func TestRunner_DuplicateRoomRejected(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}

	p := Pipeline{
		Room:  "dup",
		Steps: []Step{{Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x1", ChainId: 1}, Description: "a"}},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	err := r.Run(context.Background(), sender, p)
	if !errors.Is(err, ErrRoomBusy) {
		t.Fatalf("expected ErrRoomBusy, got %v", err)
	}
}

func TestRunner_TryHandleMissReturnsFalse(t *testing.T) {
	r := NewRunner()
	handled, err := r.TryHandleTxResult(context.Background(), types.TxResultData{
		Status: types.TxStatusConfirmed,
	}, "no-such-room", &mockSender{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false for unknown room")
	}
}

func TestRunner_OnResultErrorAbortsPipeline(t *testing.T) {
	r := NewRunner()
	sender := &mockSender{}
	var abortReason AbortReason
	done := make(chan struct{})

	p := Pipeline{
		Room: "hook-fail-room",
		Steps: []Step{
			{
				Kind: ActionKindTx, Tx: &types.TxRequest{To: "0x1", ChainId: 1}, Description: "x",
				OnResult: func(ctx context.Context, result any, s types.MessageSender) error {
					return errors.New("downstream POST failed")
				},
			},
		},
		OnAbort: func(ctx context.Context, reason AbortReason, err error, s types.MessageSender) error {
			abortReason = reason
			close(done)
			return nil
		},
	}
	if err := r.Run(context.Background(), sender, p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	r.HandleTxResult(context.Background(), types.TxResultData{
		Status: types.TxStatusConfirmed,
	}, "hook-fail-room", sender)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("OnAbort did not fire")
	}
	if abortReason != AbortReasonFailed {
		t.Fatalf("expected AbortReasonFailed, got %v", abortReason)
	}
}

func TestValidateSteps_SignatureCannotBeOptional(t *testing.T) {
	err := validateSteps([]Step{
		{Kind: ActionKindSignature, Signature: &types.SignatureRequest{Method: types.SignMethodPersonalSign, Message: "x"}, Optional: true},
	})
	if err == nil {
		t.Fatal("expected validation error for optional signature")
	}
}
