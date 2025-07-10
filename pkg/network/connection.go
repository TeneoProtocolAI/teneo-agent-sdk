package network

import (
	"time"
)

// ReconnectionManager handles automatic reconnection logic
type ReconnectionManager struct {
	enabled     bool
	attempts    int
	maxAttempts int
	delay       time.Duration
	backoffFunc func(int) time.Duration
}

// ShouldReconnect returns whether reconnection should be attempted
func (r *ReconnectionManager) ShouldReconnect() bool {
	return r.enabled && r.attempts < r.maxAttempts
}

// NextBackoff calculates the next backoff delay
func (r *ReconnectionManager) NextBackoff() time.Duration {
	if r.backoffFunc != nil {
		return r.backoffFunc(r.attempts)
	}
	return r.delay
}

// Reset resets the reconnection attempts counter
func (r *ReconnectionManager) Reset() {
	r.attempts = 0
}

// GetAttempts returns the current number of attempts
func (r *ReconnectionManager) GetAttempts() int {
	return r.attempts
}

// GetMaxAttempts returns the maximum number of attempts
func (r *ReconnectionManager) GetMaxAttempts() int {
	return r.maxAttempts
}

// IsEnabled returns whether reconnection is enabled
func (r *ReconnectionManager) IsEnabled() bool {
	return r.enabled
}

// SetEnabled enables or disables reconnection
func (r *ReconnectionManager) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// IncrementAttempts increments the attempts counter
func (r *ReconnectionManager) IncrementAttempts() {
	r.attempts++
}
