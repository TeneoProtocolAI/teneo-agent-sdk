package alerting

import (
	"fmt"
	"time"
)

// FormatSlackAlert formats an alert as a Slack Block Kit attachment
func FormatSlackAlert(alert Alert) map[string]interface{} {
	color := "#FFA500" // orange for task failures
	emoji := ":large_orange_circle:"
	failureType := "Task Failure"

	if alert.Type == AlertTypeAgentCrash {
		color = "#FF0000" // red for crashes
		emoji = ":red_circle:"
		failureType = "Agent Crash"
	}

	title := fmt.Sprintf("%s %s: %s", emoji, failureType, alert.AgentName)

	// Build fields
	fields := []map[string]interface{}{
		field("Agent", alert.AgentName, true),
		field("Wallet", truncateWallet(alert.AgentWallet), true),
		field("Failure Type", failureType, true),
		field("Consecutive Errors", fmt.Sprintf("%d", alert.ConsecutiveErrors), true),
	}

	if alert.TaskID != "" {
		fields = append(fields, field("Task ID", alert.TaskID, true))
	}

	if alert.Type == AlertTypeAgentCrash {
		fields = append(fields, field("Exit Code", fmt.Sprintf("%d", alert.ExitCode), true))
	}

	lastSuccess := "N/A"
	if !alert.LastSuccessTime.IsZero() {
		lastSuccess = alert.LastSuccessTime.UTC().Format(time.RFC3339)
	}
	fields = append(fields, field("Last Success", lastSuccess, true))

	if alert.HealthStatus != "" {
		fields = append(fields, field("Health Status", alert.HealthStatus, true))
	}

	fields = append(fields, field("Timestamp", alert.Timestamp.UTC().Format(time.RFC3339), true))

	// Build the attachment
	attachment := map[string]interface{}{
		"color":  color,
		"title":  title,
		"fields": fields,
		"text":   fmt.Sprintf("*Error:* %s", alert.ErrorMessage),
	}

	if alert.StackTrace != "" {
		attachment["text"] = fmt.Sprintf("*Error:* %s\n```%s```", alert.ErrorMessage, alert.StackTrace)
	}

	return map[string]interface{}{
		"attachments": []map[string]interface{}{attachment},
	}
}

func field(title, value string, short bool) map[string]interface{} {
	return map[string]interface{}{
		"title": title,
		"value": value,
		"short": short,
	}
}

func truncateWallet(wallet string) string {
	if len(wallet) <= 10 {
		return wallet
	}
	return wallet[:6] + "..." + wallet[len(wallet)-4:]
}
