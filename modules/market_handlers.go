package modules

import (
	"fmt"
	"os"
	"strings"
)

const replyTimeLayout = "2006-01-02 15:04:05 MST"

// BuildHypeReply returns a human-friendly hype summary for a symbol.
func BuildHypeReply(symbol string) string {
	sym := strings.TrimSpace(symbol)
	if sym == "" {
		return "Hype: unknown symbol"
	}
	if strings.ToLower(os.Getenv("MOCK_MODE")) == "true" {
		return fmt.Sprintf("Hype score for $%s: 0.00\nTrend: Trend snapshot for %s (mock): bullish momentum, strong volume spikes\n24h Move: 0.00%%", strings.ToUpper(sym), strings.ToUpper(sym))
	}

	md, err := GetMarketData(sym)
	if err != nil {
		return fmt.Sprintf("Hype score for $%s: (data unavailable). Reason: %v", strings.ToUpper(sym), summarizeErr(err))
	}

	score := ComputeHypeScore(md)

	trend := "neutral"
	if md.Change24h >= 2.0 {
		trend = "bullish"
	} else if md.Change24h <= -2.0 {
		trend = "bearish"
	}

	reply := fmt.Sprintf(
		"Hype score for $%s: %.2f\nTrend: %s (24h change: %.2f%%)\nPrice: $%0.6f • 24h Volume: $%.0f • MarketCap: $%.0f\nData as of: %s",
		strings.ToUpper(sym),
		score,
		strings.Title(trend),
		md.Change24h,
		md.PriceUSD,
		md.Volume24h,
		md.MarketCapUSD,
		md.RetrievedAt.Format(replyTimeLayout),
	)
	return reply
}

// BuildSentimentReply returns a simple sentiment summary for a token.
func BuildSentimentReply(symbol string) string {
	sym := strings.TrimSpace(symbol)
	if sym == "" {
		return "Sentiment: unknown symbol"
	}
	if strings.ToLower(os.Getenv("MOCK_MODE")) == "true" {
		return fmt.Sprintf("Sentiment for $%s:\n👍 0.0%% positive\n👎 0.0%% negative", strings.ToUpper(sym))
	}

	md, err := GetMarketData(sym)
	if err != nil {
		return fmt.Sprintf("Sentiment for $%s: (data unavailable). Reason: %v", strings.ToUpper(sym), summarizeErr(err))
	}

	pos := 0.0
	neg := 0.0
	if md.Change24h > 1.0 {
		pos = 75.0
		neg = 25.0
	} else if md.Change24h < -1.0 {
		pos = 25.0
		neg = 75.0
	} else {
		pos = 50.0
		neg = 50.0
	}

	return fmt.Sprintf("Sentiment for $%s:\n👍 %.1f%% positive\n👎 %.1f%% negative\nPrice: $%.6f (24h: %+0.2f%%)",
		strings.ToUpper(sym), pos, neg, md.PriceUSD, md.Change24h)
}

// BuildRiskReply returns a small risk-check summary.
func BuildRiskReply(symbol string) string {
	sym := strings.TrimSpace(symbol)
	if sym == "" {
		return "Risk: unknown symbol"
	}
	if strings.ToLower(os.Getenv("MOCK_MODE")) == "true" {
		return fmt.Sprintf("Risk check for $%s:\n- RiskScore: 0.30\n- Indicators:\n - Very low market cap", strings.ToUpper(sym))
	}

	md, err := GetMarketData(sym)
	if err != nil {
		return fmt.Sprintf("Risk check for $%s: (data unavailable). Reason: %v", strings.ToUpper(sym), summarizeErr(err))
	}

	score := 0.0
	if md.MarketCapUSD <= 0 {
		score = 0.9
	} else {
		mc := md.MarketCapUSD
		if mc < 1_000_000 {
			score = 0.85
		} else if mc < 10_000_000 {
			score = 0.6
		} else if mc < 100_000_000 {
			score = 0.4
		} else if mc < 1_000_000_000 {
			score = 0.25
		} else {
			score = 0.12
		}
	}
	if md.Change24h > 5 || md.Change24h < -5 {
		score = score + 0.15
	}
	if score > 1 {
		score = 1
	}

	indicators := []string{}
	if md.MarketCapUSD < 10_000_000 {
		indicators = append(indicators, "Very low market cap")
	}
	if md.Volume24h < 10_000 {
		indicators = append(indicators, "Very low volume")
	}
	if md.Change24h > 5 {
		indicators = append(indicators, "Large positive price spike (24h)")
	}
	if md.Change24h < -5 {
		indicators = append(indicators, "Large negative price drop (24h)")
	}
	if len(indicators) == 0 {
		indicators = append(indicators, "No immediate red flags")
	}

	reply := fmt.Sprintf("Risk check for $%s:\n- RiskScore: %.2f\n- Indicators:\n - %s\nPrice: $%.6f • MarketCap: $%.0f • 24h: %+0.2f%%",
		strings.ToUpper(sym),
		score,
		strings.Join(indicators, "\n - "),
		md.PriceUSD,
		md.MarketCapUSD,
		md.Change24h,
	)
	return reply
}

func summarizeErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
