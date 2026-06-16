package txflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// DefaultPipelineTimeout is the per-pipeline wall-clock budget used when
// Pipeline.Timeout is zero. Chosen to cover "user saw the confirm sheet,
// checked their hardware wallet, and came back" without leaking state forever
// if they abandon the tab.
const DefaultPipelineTimeout = 30 * time.Minute

// ErrRoomBusy is returned from Run when another pipeline is already active in
// the same room. Agents should treat this as "user has a pending flow, wait".
var ErrRoomBusy = errors.New("txflow: pipeline already running for room")

// ErrEmptyPipeline is returned from Run when a pipeline has no steps. Likely a bug.
var ErrEmptyPipeline = errors.New("txflow: pipeline has no steps")

// Runner drives Pipelines. It implements both types.TxResultHandler and
// types.SignatureResultHandler so agents can embed *Runner in their handler
// struct and get pipeline dispatch automatically. Agents with mixed needs
// (both pipelines and direct single-action flows) can instead call the
// TryHandle* methods and fall back to their own logic on miss.
type Runner struct {
	mu             sync.Mutex
	pipelines      map[string]*pipelineState // keyed by room
	defaultTimeout time.Duration
}

type pipelineState struct {
	pipeline    *Pipeline
	currentStep int
	sender      types.MessageSender
	cancelTimer func()
	done        bool
}

// NewRunner returns a Runner with default timeouts.
func NewRunner() *Runner {
	return &Runner{
		pipelines:      make(map[string]*pipelineState),
		defaultTimeout: DefaultPipelineTimeout,
	}
}

// SetDefaultTimeout overrides the pipeline timeout used when Pipeline.Timeout is zero.
func (r *Runner) SetDefaultTimeout(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultTimeout = d
}

// Run registers the pipeline for its room and triggers the first step.
// Returns synchronously after the first TriggerWalletTx / TriggerWalletSignature
// call — remaining steps run via the result callbacks. Returns ErrRoomBusy if
// another pipeline is already active in p.Room.
func (r *Runner) Run(ctx context.Context, sender types.MessageSender, p Pipeline) error {
	if len(p.Steps) == 0 {
		return ErrEmptyPipeline
	}
	if err := validateSteps(p.Steps); err != nil {
		return err
	}
	if p.Room == "" {
		return errors.New("txflow: pipeline.Room is required")
	}

	r.mu.Lock()
	if _, exists := r.pipelines[p.Room]; exists {
		r.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrRoomBusy, p.Room)
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = r.defaultTimeout
	}
	state := &pipelineState{
		pipeline: &p,
		sender:   sender,
	}
	r.pipelines[p.Room] = state

	// Timeout timer. We store cancel fn so a terminal advance can stop it.
	timer := time.AfterFunc(timeout, func() {
		r.abort(p.Room, AbortReasonTimeout, fmt.Errorf("pipeline timed out after %s", timeout))
	})
	state.cancelTimer = func() { timer.Stop() }
	r.mu.Unlock()

	if err := r.triggerStepLocked(state, ctx); err != nil {
		// Trigger failed synchronously — clean up, no OnAbort since user never saw anything.
		r.mu.Lock()
		delete(r.pipelines, p.Room)
		if state.cancelTimer != nil {
			state.cancelTimer()
		}
		r.mu.Unlock()
		return fmt.Errorf("txflow: failed to trigger first step: %w", err)
	}
	return nil
}

// Cancel aborts the pipeline running in the given room, if any. Used when the
// agent knows out-of-band that the flow is no longer relevant (e.g. user asked
// for a different action). No-op if no pipeline is active.
func (r *Runner) Cancel(room string) {
	r.abort(room, AbortReasonCanceled, errors.New("canceled by agent"))
}

// HasPipeline reports whether a pipeline is currently active for the given room.
func (r *Runner) HasPipeline(room string) bool {
	r.mu.Lock()
	_, ok := r.pipelines[room]
	r.mu.Unlock()
	return ok
}

// HandleTxResult implements types.TxResultHandler. Safe to call whether or not
// a pipeline is active for the room — unhandled results are ignored.
func (r *Runner) HandleTxResult(ctx context.Context, result types.TxResultData, room string, sender types.MessageSender) error {
	handled, err := r.TryHandleTxResult(ctx, result, room, sender)
	_ = handled
	return err
}

// TryHandleTxResult advances a pipeline if one is active for the room. Returns
// (true, err) if the result was routed to a pipeline, (false, nil) otherwise.
// Agents with non-pipeline tx flows should call this first and handle misses
// themselves.
func (r *Runner) TryHandleTxResult(ctx context.Context, result types.TxResultData, room string, sender types.MessageSender) (bool, error) {
	r.mu.Lock()
	state, ok := r.pipelines[room]
	if !ok {
		r.mu.Unlock()
		return false, nil
	}
	// Verify the step expects a tx result — otherwise this may be a stray
	// result for a previous flow in the same room (shouldn't happen with the
	// one-pipeline-per-room invariant, but cheap safety).
	step := state.pipeline.Steps[state.currentStep]
	if step.Kind != ActionKindTx {
		r.mu.Unlock()
		log.Printf("txflow: dropping tx_result for room %s — current step is not a tx step", room)
		return true, nil
	}
	r.mu.Unlock()

	if !result.Status.IsTerminal() {
		// "broadcasted" is intermediate — don't advance yet. Agents can still
		// surface intermediate status via their own SendChunk/Update if desired.
		return true, nil
	}

	if result.Status == types.TxStatusRejected {
		if step.Optional {
			return true, r.advance(ctx, room, &result, sender)
		}
		r.abort(room, AbortReasonRejected, fmt.Errorf("user rejected tx at step %d", state.currentStep))
		return true, nil
	}
	if result.Status == types.TxStatusFailed {
		r.abort(room, AbortReasonFailed, fmt.Errorf("tx failed at step %d: %s", state.currentStep, result.Error))
		return true, nil
	}
	// Confirmed.
	return true, r.advance(ctx, room, &result, sender)
}

// HandleSignatureResult implements types.SignatureResultHandler.
func (r *Runner) HandleSignatureResult(ctx context.Context, result types.SignatureResultData, room string, sender types.MessageSender) error {
	handled, err := r.TryHandleSignatureResult(ctx, result, room, sender)
	_ = handled
	return err
}

// TryHandleSignatureResult advances a pipeline waiting on a signature step.
// Returns (true, err) if routed to a pipeline, (false, nil) otherwise.
func (r *Runner) TryHandleSignatureResult(ctx context.Context, result types.SignatureResultData, room string, sender types.MessageSender) (bool, error) {
	r.mu.Lock()
	state, ok := r.pipelines[room]
	if !ok {
		r.mu.Unlock()
		return false, nil
	}
	step := state.pipeline.Steps[state.currentStep]
	if step.Kind != ActionKindSignature {
		r.mu.Unlock()
		log.Printf("txflow: dropping signature_result for room %s — current step is not a signature step", room)
		return true, nil
	}
	r.mu.Unlock()

	switch result.Status {
	case types.SignatureStatusRejected:
		r.abort(room, AbortReasonRejected, fmt.Errorf("user rejected signature at step %d", state.currentStep))
		return true, nil
	case types.SignatureStatusFailed:
		r.abort(room, AbortReasonFailed, fmt.Errorf("signature failed at step %d: %s", state.currentStep, result.Error))
		return true, nil
	case types.SignatureStatusSigned:
		return true, r.advance(ctx, room, &result, sender)
	default:
		log.Printf("txflow: unknown signature status %q for room %s", result.Status, room)
		return true, nil
	}
}

// advance runs the current step's OnResult hook, then either completes the
// pipeline or triggers the next step.
func (r *Runner) advance(ctx context.Context, room string, result any, sender types.MessageSender) error {
	r.mu.Lock()
	state, ok := r.pipelines[room]
	if !ok || state.done {
		r.mu.Unlock()
		return nil
	}
	step := state.pipeline.Steps[state.currentStep]
	r.mu.Unlock()

	if step.OnResult != nil {
		if err := step.OnResult(ctx, result, sender); err != nil {
			r.abort(room, AbortReasonFailed, fmt.Errorf("OnResult hook failed at step %d: %w", state.currentStep, err))
			return nil
		}
	}

	r.mu.Lock()
	state.currentStep++
	atEnd := state.currentStep >= len(state.pipeline.Steps)
	r.mu.Unlock()

	if atEnd {
		r.complete(ctx, room, sender)
		return nil
	}
	return r.triggerStepLocked(state, ctx)
}

// triggerStepLocked triggers the current step. Name has "Locked" suffix because
// it assumes state pointer is stable — callers either hold r.mu or know no one
// else can mutate state (e.g. inside Run, before insertion into the map's
// observable reach). It does NOT take r.mu itself to avoid re-entry.
func (r *Runner) triggerStepLocked(state *pipelineState, ctx context.Context) error {
	_ = ctx // Sender calls already carry their own context via the protocol handler.
	step := state.pipeline.Steps[state.currentStep]
	desc := step.Description
	if desc == "" {
		desc = state.pipeline.Description
	}
	switch step.Kind {
	case ActionKindTx:
		if step.Tx == nil {
			return fmt.Errorf("step %d: Kind=Tx but Tx is nil", state.currentStep)
		}
		return state.sender.TriggerWalletTx(*step.Tx, desc, step.Optional)
	case ActionKindSignature:
		if step.Signature == nil {
			return fmt.Errorf("step %d: Kind=Signature but Signature is nil", state.currentStep)
		}
		return state.sender.TriggerWalletSignature(*step.Signature, desc)
	default:
		return fmt.Errorf("step %d: unknown Kind %d", state.currentStep, step.Kind)
	}
}

func (r *Runner) complete(ctx context.Context, room string, sender types.MessageSender) {
	r.mu.Lock()
	state, ok := r.pipelines[room]
	if !ok || state.done {
		r.mu.Unlock()
		return
	}
	state.done = true
	if state.cancelTimer != nil {
		state.cancelTimer()
	}
	delete(r.pipelines, room)
	onComplete := state.pipeline.OnComplete
	r.mu.Unlock()

	if onComplete != nil {
		if err := onComplete(ctx, sender); err != nil {
			log.Printf("txflow: OnComplete failed for room %s: %v", room, err)
		}
	}
}

func (r *Runner) abort(room string, reason AbortReason, cause error) {
	r.mu.Lock()
	state, ok := r.pipelines[room]
	if !ok || state.done {
		r.mu.Unlock()
		return
	}
	state.done = true
	if state.cancelTimer != nil {
		state.cancelTimer()
	}
	delete(r.pipelines, room)
	onAbort := state.pipeline.OnAbort
	sender := state.sender
	r.mu.Unlock()

	if onAbort != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := onAbort(ctx, reason, cause, sender); err != nil {
			log.Printf("txflow: OnAbort failed for room %s: %v", room, err)
		}
	} else {
		log.Printf("txflow: pipeline aborted in room %s: reason=%s cause=%v", room, reason, cause)
	}
}

func validateSteps(steps []Step) error {
	for i, s := range steps {
		switch s.Kind {
		case ActionKindTx:
			if s.Tx == nil {
				return fmt.Errorf("step %d: Kind=Tx requires Tx != nil", i)
			}
		case ActionKindSignature:
			if s.Signature == nil {
				return fmt.Errorf("step %d: Kind=Signature requires Signature != nil", i)
			}
			if s.Optional {
				return fmt.Errorf("step %d: signature steps cannot be Optional", i)
			}
		default:
			return fmt.Errorf("step %d: Kind must be set (ActionKindTx or ActionKindSignature)", i)
		}
	}
	return nil
}
