package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeTaskFailure AlertType = "task_failure"
	AlertTypeAgentCrash  AlertType = "agent_crash"
)

// SlackConfig holds configuration for Slack alerting
type SlackConfig struct {
	WebhookURL      string
	AgentName       string
	AgentWallet     string
	ThrottleSeconds int // 0 = use default (60s)
}

// AlertMetrics tracks alerting-related metrics
type AlertMetrics struct {
	ConsecutiveErrors int
	LastSuccessTime   time.Time
	TotalAlertsSent   int64
	mu                sync.RWMutex
}

// SlackAlerter sends failure alerts to Slack
type SlackAlerter struct {
	webhookURL  string
	agentName   string
	agentWallet string
	throttler   *AlertThrottler
	metrics     *AlertMetrics
	httpClient  *http.Client
}

// NewSlackAlerter creates a new Slack alerter
func NewSlackAlerter(config SlackConfig) *SlackAlerter {
	if config.WebhookURL == "" {
		return nil
	}

	throttleWindow := 60 * time.Second
	if config.ThrottleSeconds > 0 {
		throttleWindow = time.Duration(config.ThrottleSeconds) * time.Second
	}

	return &SlackAlerter{
		webhookURL:  config.WebhookURL,
		agentName:   config.AgentName,
		agentWallet: config.AgentWallet,
		throttler:   NewAlertThrottler(throttleWindow),
		metrics:     &AlertMetrics{},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RecordSuccess records a successful task execution
func (s *SlackAlerter) RecordSuccess() {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()
	s.metrics.ConsecutiveErrors = 0
	s.metrics.LastSuccessTime = time.Now()
}

// RecordFailure increments the consecutive error counter
func (s *SlackAlerter) RecordFailure() {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()
	s.metrics.ConsecutiveErrors++
}

// SendTaskFailure sends a task failure alert to Slack
func (s *SlackAlerter) SendTaskFailure(taskID, errorMsg, stackTrace string) {
	if s == nil {
		return
	}

	s.RecordFailure()

	s.metrics.mu.RLock()
	consecutiveErrors := s.metrics.ConsecutiveErrors
	lastSuccess := s.metrics.LastSuccessTime
	s.metrics.mu.RUnlock()

	alert := Alert{
		Type:              AlertTypeTaskFailure,
		AgentName:         s.agentName,
		AgentWallet:       s.agentWallet,
		TaskID:            taskID,
		ErrorMessage:      errorMsg,
		StackTrace:        truncate(stackTrace, 500),
		ConsecutiveErrors: consecutiveErrors,
		LastSuccessTime:   lastSuccess,
		Timestamp:         time.Now(),
	}

	s.sendAlert(alert)
}

// SendAgentCrash sends an agent crash alert to Slack
// Exit code 1 (force close / signal interrupt) is skipped to avoid noisy alerts on intentional shutdowns
func (s *SlackAlerter) SendAgentCrash(reason string, exitCode int) {
	if s == nil {
		return
	}

	// Skip notification for force close (exit code 1)
	if exitCode == 1 {
		log.Printf("🔕 Skipping Slack alert for force close (exit code 1): %s", reason)
		return
	}

	s.metrics.mu.RLock()
	consecutiveErrors := s.metrics.ConsecutiveErrors
	lastSuccess := s.metrics.LastSuccessTime
	s.metrics.mu.RUnlock()

	alert := Alert{
		Type:              AlertTypeAgentCrash,
		AgentName:         s.agentName,
		AgentWallet:       s.agentWallet,
		ErrorMessage:      reason,
		ExitCode:          exitCode,
		ConsecutiveErrors: consecutiveErrors,
		LastSuccessTime:   lastSuccess,
		Timestamp:         time.Now(),
	}

	// Crash alerts bypass throttling
	s.postToSlack(FormatSlackAlert(alert))
}

func (s *SlackAlerter) sendAlert(alert Alert) {
	if !s.throttler.ShouldSend(alert) {
		log.Printf("🔕 Alert throttled: %s - %s", alert.Type, firstLine(alert.ErrorMessage))
		return
	}

	s.postToSlack(FormatSlackAlert(alert))
}

func (s *SlackAlerter) postToSlack(payload map[string]interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Failed to marshal Slack alert: %v", err)
		return
	}

	resp, err := s.httpClient.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("❌ Failed to send Slack alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Slack webhook returned status %d", resp.StatusCode)
		return
	}

	s.metrics.mu.Lock()
	s.metrics.TotalAlertsSent++
	s.metrics.mu.Unlock()

	log.Printf("📢 Slack alert sent: %s", s.agentName)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}

// Alert represents an alert to be sent
type Alert struct {
	Type              AlertType
	AgentName         string
	AgentWallet       string
	TaskID            string
	ErrorMessage      string
	StackTrace        string
	ExitCode          int
	ConsecutiveErrors int
	LastSuccessTime   time.Time
	HealthStatus      string
	Timestamp         time.Time
}

// ThrottleKey returns a deduplication key for this alert
func (a Alert) ThrottleKey() string {
	return fmt.Sprintf("%s:%s:%s", a.Type, a.AgentName, firstLine(a.ErrorMessage))
}
