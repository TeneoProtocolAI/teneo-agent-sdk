package alerting

import (
	"sync"
	"time"
)

// AlertThrottler prevents duplicate alerts within a configurable time window
type AlertThrottler struct {
	window    time.Duration
	lastSent  map[string]time.Time
	mu        sync.Mutex
}

// NewAlertThrottler creates a new throttler with the given window
func NewAlertThrottler(window time.Duration) *AlertThrottler {
	return &AlertThrottler{
		window:   window,
		lastSent: make(map[string]time.Time),
	}
}

// ShouldSend returns true if the alert should be sent (not throttled)
func (t *AlertThrottler) ShouldSend(alert Alert) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := alert.ThrottleKey()
	now := time.Now()

	if last, ok := t.lastSent[key]; ok {
		if now.Sub(last) < t.window {
			return false
		}
	}

	t.lastSent[key] = now
	t.cleanup(now)
	return true
}

// cleanup removes expired entries
func (t *AlertThrottler) cleanup(now time.Time) {
	for key, last := range t.lastSent {
		if now.Sub(last) > t.window*2 {
			delete(t.lastSent, key)
		}
	}
}
