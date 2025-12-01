package modules

import (
	"strings"
)

// RunHype is the public entry used by the agent to get a hype reply.
func RunHype(args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: hype [token]. Example: hype sol", nil
	}
	token := strings.TrimSpace(args[0])
	if token == "" {
		return "Usage: hype [token]. Example: hype sol", nil
	}

	reply := BuildHypeReply(token)
	return reply, nil
}
