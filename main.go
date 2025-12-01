// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"signalshield/modules"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

type SignalshieldAnalystAgent struct{}

func (a *SignalshieldAnalystAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	log.Printf("Processing task: %s", task)

	task = strings.TrimSpace(task)
	task = strings.TrimPrefix(task, "/")
	parts := strings.Fields(task)
	if len(parts) == 0 {
		return "No command provided. Available commands: scan, monitor, riskcheck, hype, signal, dumpalert, topcalls, sentiment, watch, summary, marketcap, volume, price, gecko, trend, alert, subscribe, unsubscribe, ai", nil
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "scan":
		return modules.RunScan(args)
	case "monitor":
		return "Monitor command (mock): started (use /monitor <keyword>)", nil
	case "riskcheck":
		return modules.RunRiskCheck(args)
	case "hype":
		return modules.RunHype(args)
	case "signal":
		return "Latest signals: 3 new early calls, 1 dump alert (mock).", nil
	case "dumpalert":
		return "Dump alert check: no immediate dump signals detected (mock).", nil
	case "topcalls":
		return modules.RunTopCalls()
	case "sentiment":
		return modules.RunSentiment(args)
	case "watch":
		return modules.RunWatch(args)
	case "summary":
		return modules.RunSummary()
	case "marketcap":
		if len(args) == 0 {
			return "Usage: marketcap [token]", nil
		}
		return modules.GetMarketCap(strings.Join(args, ""))
	case "volume":
		if len(args) == 0 {
			return "Usage: volume [token]", nil
		}
		return modules.GetVolume(strings.Join(args, ""))
	case "price":
		if len(args) == 0 {
			return "Usage: price [token]", nil
		}
		return modules.GetCoinPrice(strings.Join(args, ""))
	case "gecko", "geckosnapshot":
		if len(args) == 0 {
			return "Usage: gecko [id_or_symbol]", nil
		}
		res, err := modules.GetCoinGeckoFull(strings.Join(args, ""))
		if err != nil {
			return "", err
		}
		// FormatCoinGeckoSummary returns string -> must return (string, nil)
		return modules.FormatCoinGeckoSummary(res), nil
	case "trend":
		if len(args) == 0 {
			return "Usage: trend [token]", nil
		}
		// GetTrendSnapshot returns (string, error) so just forward it
		return modules.GetTrendSnapshot(strings.Join(args, ""))
	case "alert":
		return "Alert command (mock): created (use alert [token] [condition])", nil
	case "subscribe":
		return "Subscribe (mock): done", nil
	case "unsubscribe":
		return "Unsubscribe (mock): done", nil
	case "ai":
		// forward natural language instruction to GPT module
		if len(args) == 0 {
			return "Usage: ai [instruction]", nil
		}
		instr := strings.Join(args, " ")
		// IMPORTANT: ForwardToOpenAI in modules now prioritizes GOOGLE_API_KEY (if set)
		if os.Getenv("GOOGLE_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" {
			return "AI backend not configured. Set GOOGLE_API_KEY or OPENAI_API_KEY in .env", nil
		}
		resp, err := modules.ForwardToOpenAI(instr)
		if err != nil {
			return "", err
		}
		return resp, nil
	default:
		return fmt.Sprintf("Unknown command '%s'. Available commands: scan, monitor, riskcheck, hype, signal, dumpalert, topcalls, sentiment, watch, summary, marketcap, volume, price, gecko, trend, alert, subscribe, unsubscribe, ai", cmd), nil
	}
}

func main() {
	// Load .env if available
	_ = godotenv.Load()

	// Basic config & env
	rateLimitStr := os.Getenv("RATE_LIMIT_PER_MINUTE")
	rateLimit := 0
	if rateLimitStr != "" {
		if v, err := strconv.Atoi(rateLimitStr); err == nil {
			rateLimit = v
		}
	}

	pollInterval := 30
	if s := os.Getenv("X_POLL_INTERVAL"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			pollInterval = v
		}
	}
	kols := []string{"Ansem", "GCR", "TheMoonCarl"} // default
	if s := os.Getenv("KOL_LIST"); s != "" {
		parts := strings.Split(s, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) > 0 {
			kols = parts
		}
	}
	xBearer := os.Getenv("X_BEARER_TOKEN")
	source := "mock-x"
	mock := true
	if os.Getenv("MOCK_MODE") == "false" {
		mock = false
	}

	// Teneo agent config
	_ = godotenv.Load()
	config := agent.DefaultConfig()
	config.Name = "SignalShield Analyst"
	config.Description = "SignalShield Analyst monitors KOL early calls + market signals."
	config.Capabilities = []string{"early-call-detection", "risk-mitigation-engine", "sentiment-analysis", "hype-index-scoring", "dump-alert-system", "influencer-tracking", "trend-detection", "anomaly-detection", "multi-chain-token-monitoring", "risk-hype-balancer"}
	config.PrivateKey = os.Getenv("PRIVATE_KEY")
	config.NFTTokenID = os.Getenv("NFT_TOKEN_ID")
	config.OwnerAddress = os.Getenv("OWNER_ADDRESS")
	config.RateLimitPerMinute = rateLimit

	enhancedAgent, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       config,
		AgentHandler: &SignalshieldAnalystAgent{},
	})
	if err != nil {
		log.Fatal("agent.NewEnhancedAgent:", err)
	}

	log.Println("Starting SignalShield Analyst...")
	// run agent in goroutine so we can also start scanner & detection loop
	go enhancedAgent.Run()

	// Create context for scanner & detector
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// detection channel
	detectCh := make(chan modules.Detection, 16)

	// start scanner (xscanner)
	go modules.StartXScanner(ctx, pollInterval, kols, xBearer, source, mock, detectCh)

	// goroutine to handle detections
	go func() {
		for d := range detectCh {
			// Save detection to file (modules.SaveDetection expects Detection)
			if err := modules.SaveDetection("alerts.log", d); err != nil {
				log.Println("Warning: SaveDetection failed:", err)
			}
			// optional: forward text to model pipeline for short summary (non-blocking)
			go func(det modules.Detection) {
				// prefer GOOGLE_API_KEY if set, otherwise OPENAI_API_KEY
				if os.Getenv("GOOGLE_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" {
					return
				}
				res, err := modules.ForwardToOpenAI(det.Text)
				if err != nil {
					log.Println("ForwardToOpenAI err:", err)
					return
				}
				log.Println("[xscanner] GPT summary:", res)
			}(d)
		}
	}()

	// health server (simple)
	httpPort := "8080"
	if p := os.Getenv("HEALTH_PORT"); p != "" {
		httpPort = p
	}
	// Provide a very small health endpoint (so curl http://localhost:8080/health works)
	go func() {
		ln := ":" + httpPort
		log.Printf("HTTP server listening on :%s", httpPort)
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"agent":"%s","status":"healthy","timestamp":"%s","kols":%q,"mock":%v,"pollSec":%d}`, config.Name, time.Now().UTC().Format(time.RFC3339), kols, mock, pollInterval)))
		})
		if err := http.ListenAndServe(ln, nil); err != nil {
			log.Println("health server error:", err)
		}
	}()

	// block forever (agent runs in background)
	select {}
}
