// Package txflow provides a multi-step "user action" pipeline (wallet transactions
// and off-chain signatures) for Teneo agents.
//
// Agents that chain actions — approve → swap, approve → bridge → confirm,
// approve → sign intent → settle, etc. — otherwise each reimplement the same
// per-room state machine: tracking which step is in flight, advancing on
// confirmed/signed, aborting on rejected/failed, GC'ing state when the user
// walks away. txflow centralizes that so agents write pipelines, not mutexes.
//
// A Runner implements both types.TxResultHandler and types.SignatureResultHandler,
// so agents can either embed *Runner directly (simplest) or keep their own
// handlers and delegate via Runner.TryHandleTxResult / TryHandleSignatureResult
// when they also have non-pipeline flows.
package txflow

import (
	"context"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// ActionKind discriminates the union in Step between a wallet tx and an off-chain
// signature. Pipelines are heterogeneous by design — intent-based flows commonly
// interleave both in a single user journey.
type ActionKind int

const (
	ActionKindTx ActionKind = iota + 1
	ActionKindSignature
)

// Step is a single user action in a pipeline. Set Kind and the matching payload
// field (Tx for ActionKindTx, Signature for ActionKindSignature); the other is ignored.
type Step struct {
	Kind        ActionKind
	Tx          *types.TxRequest
	Signature   *types.SignatureRequest
	Description string
	// Optional, only honored for ActionKindTx — if true and the user rejects,
	// the pipeline continues to the next step instead of aborting. Off-chain
	// signatures are never optional (intent flows can't proceed without one).
	Optional bool
	// OnResult fires after the step reaches a terminal status (confirmed for tx,
	// signed for signature). The result argument is *types.TxResultData or
	// *types.SignatureResultData depending on Kind. Use this for per-step
	// bookkeeping or to forward the result to an external endpoint. Returning an
	// error aborts the pipeline with AbortReasonFailed.
	OnResult func(ctx context.Context, result any, sender types.MessageSender) error
}

// AbortReason explains why OnAbort fired.
type AbortReason int

const (
	// AbortReasonRejected: user declined a non-optional step in their wallet.
	AbortReasonRejected AbortReason = iota + 1
	// AbortReasonFailed: the wallet or a step's OnResult hook reported failure.
	AbortReasonFailed
	// AbortReasonTimeout: no terminal status arrived within Timeout (user closed
	// the tab, wallet hung, etc.). Without this, pending state would leak forever.
	AbortReasonTimeout
	// AbortReasonCanceled: Runner.Cancel was called for this room.
	AbortReasonCanceled
)

func (r AbortReason) String() string {
	switch r {
	case AbortReasonRejected:
		return "rejected"
	case AbortReasonFailed:
		return "failed"
	case AbortReasonTimeout:
		return "timeout"
	case AbortReasonCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// Pipeline is a declarative description of a user-action sequence. Room is the
// routing key — only one pipeline may be active per room at a time.
type Pipeline struct {
	Room        string
	Description string
	Steps       []Step
	// Timeout is the max wall-clock window for the entire pipeline. Zero uses the
	// Runner's default (30 minutes). On expiry the pipeline is aborted with
	// AbortReasonTimeout and its state is GC'd.
	Timeout    time.Duration
	OnComplete func(ctx context.Context, sender types.MessageSender) error
	OnAbort    func(ctx context.Context, reason AbortReason, err error, sender types.MessageSender) error
}

// MarshalJSON / UnmarshalJSON helpers for the `any` result passed to OnResult —
// exposed so callers can switch on type without importing reflect.

// TxResultOf returns r cast to *types.TxResultData, or nil if r is not a tx result.
func TxResultOf(r any) *types.TxResultData {
	v, _ := r.(*types.TxResultData)
	return v
}

// SignatureResultOf returns r cast to *types.SignatureResultData, or nil if r is
// not a signature result.
func SignatureResultOf(r any) *types.SignatureResultData {
	v, _ := r.(*types.SignatureResultData)
	return v
}
