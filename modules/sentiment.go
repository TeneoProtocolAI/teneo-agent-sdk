package modules

import (
	"strings"
)

// RunSentiment provides the public entry used by the agent to return sentiment.
func RunSentiment(args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: sentiment [token]", nil
	}
	token := strings.TrimSpace(args[0])
	if token == "" {
		return "Usage: sentiment [token]", nil
	}

	reply := BuildSentimentReply(token)
	return reply, nil
}
