package openclaw

import "errors"

var (
	// ErrUnavailable is returned when the OpenClaw service is not reachable
	ErrUnavailable = errors.New("openclaw: service unavailable")

	// ErrAuth is returned when authentication with OpenClaw fails
	ErrAuth = errors.New("openclaw: authentication failed")

	// ErrTimeout is returned when a request to OpenClaw times out
	ErrTimeout = errors.New("openclaw: request timed out")

	// ErrInvalidResponse is returned when OpenClaw returns an unparseable response
	ErrInvalidResponse = errors.New("openclaw: invalid response")
)
