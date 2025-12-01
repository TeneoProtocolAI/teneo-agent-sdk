package modules

import (
	"strings"
)

// RunRiskCheck is the public entry used by the agent to perform risk checks.
func RunRiskCheck(args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: riskcheck [token]. Example: riskcheck sol", nil
	}
	token := strings.TrimSpace(args[0])
	if token == "" {
		return "Usage: riskcheck [token]. Example: riskcheck sol", nil
	}

	reply := BuildRiskReply(token)
	return reply, nil
}
