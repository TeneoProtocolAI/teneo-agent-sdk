package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Compatibility layer: provides the older helper names expected by main.go
// implemented on top of GetMarketData and direct CoinGecko queries.

// GetCoinGeckoFull fetches the full CoinGecko JSON for a given id or symbol.
// Returns a generic map (same shape as JSON).
func GetCoinGeckoFull(idOrSymbol string) (map[string]interface{}, error) {
	s := strings.TrimSpace(idOrSymbol)
	if s == "" {
		return nil, fmt.Errorf("empty idOrSymbol")
	}

	// try symbol mapping first
	l := strings.ToLower(s)
	if id, ok := cgSymbolToID[l]; ok {
		l = id
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false", l)
	client := &http.Client{Timeout: 12 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko http err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// read body to include in error (but truncate)
		bodyB, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("coingecko status %d: %s", resp.StatusCode, string(bodyB))
	}

	var out map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("coingecko decode err: %w", err)
	}
	return out, nil
}

// FormatCoinGeckoSummary formats a compact human readable summary from the
// full coin JSON (the shape returned by GetCoinGeckoFull).
func FormatCoinGeckoSummary(full map[string]interface{}) string {
	if full == nil {
		return "Coin data: unavailable"
	}

	// helpers to extract safely
	getString := func(keys ...string) string {
		var cur interface{} = full
		for _, k := range keys {
			if m, ok := cur.(map[string]interface{}); ok {
				cur = m[k]
			} else {
				return ""
			}
		}
		if s, ok := cur.(string); ok {
			return s
		}
		return ""
	}
	getFloat := func(keys ...string) float64 {
		var cur interface{} = full
		for _, k := range keys {
			if m, ok := cur.(map[string]interface{}); ok {
				cur = m[k]
			} else {
				return 0
			}
		}
		switch v := cur.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		default:
			return 0
		}
	}

	name := getString("name")
	symbol := getString("symbol")
	if symbol == "" {
		// fallback try top-level id
		symbol = getString("id")
	}
	price := getFloat("market_data", "current_price", "usd")
	change24 := getFloat("market_data", "price_change_percentage_24h")
	vol := getFloat("market_data", "total_volume", "usd")
	mcap := getFloat("market_data", "market_cap", "usd")

	// Build summary
	summary := fmt.Sprintf("%s (%s)\nPrice: $%.6f\n24h: %+0.2f%% • Volume: $%.0f • MarketCap: $%.0f",
		nameOr(symbol, name), strings.ToUpper(symbol), price, change24, vol, mcap)
	return summary
}

func nameOr(sym, name string) string {
	if name != "" {
		return name
	}
	if sym != "" {
		return strings.ToUpper(sym)
	}
	return "unknown"
}

// GetMarketCap returns a human readable market cap string for the symbol.
// Signature matches main.go expectation: returns (string, error)
func GetMarketCap(symbol string) (string, error) {
	if strings.TrimSpace(symbol) == "" {
		return "Usage: marketcap [token]", nil
	}
	// Prefer using our fast GetMarketData cache
	md, err := GetMarketData(symbol)
	if err == nil {
		if md.MarketCapUSD > 0 {
			return fmt.Sprintf("$%.0f", md.MarketCapUSD), nil
		}
		// if not present, fall through to full fetch
	}

	// fallback to full fetch
	full, err := GetCoinGeckoFull(symbol)
	if err != nil {
		return "", err
	}
	mcap := safeGetFloat(full, "market_data", "market_cap", "usd")
	if mcap <= 0 {
		return "Market cap: unavailable", nil
	}
	return fmt.Sprintf("$%.0f", mcap), nil
}

// GetVolume returns 24h volume for the symbol as string (string, error)
func GetVolume(symbol string) (string, error) {
	if strings.TrimSpace(symbol) == "" {
		return "Usage: volume [token]", nil
	}
	md, err := GetMarketData(symbol)
	if err == nil {
		if md.Volume24h > 0 {
			return fmt.Sprintf("$%.0f", md.Volume24h), nil
		}
	}
	full, err := GetCoinGeckoFull(symbol)
	if err != nil {
		return "", err
	}
	vol := safeGetFloat(full, "market_data", "total_volume", "usd")
	if vol <= 0 {
		return "Volume: unavailable", nil
	}
	return fmt.Sprintf("$%.0f", vol), nil
}

// GetCoinPrice returns the current USD price as string (string, error)
func GetCoinPrice(symbol string) (string, error) {
	if strings.TrimSpace(symbol) == "" {
		return "Usage: price [token]", nil
	}
	md, err := GetMarketData(symbol)
	if err == nil && md.PriceUSD > 0 {
		return fmt.Sprintf("$%.6f", md.PriceUSD), nil
	}
	full, err := GetCoinGeckoFull(symbol)
	if err != nil {
		return "", err
	}
	price := safeGetFloat(full, "market_data", "current_price", "usd")
	if price <= 0 {
		return "Price: unavailable", nil
	}
	return fmt.Sprintf("$%.6f", price), nil
}

// GetTrendSnapshot returns a short human-readable trend string for a token.
// Signature: (string, error)
func GetTrendSnapshot(symbol string) (string, error) {
	if strings.TrimSpace(symbol) == "" {
		return "Usage: trend [token]", nil
	}
	// Use quick market data
	md, err := GetMarketData(symbol)
	if err != nil {
		// try full fallback for more fields
		full, err2 := GetCoinGeckoFull(symbol)
		if err2 != nil {
			return "", err
		}
		// attempt to derive change
		change := safeGetFloat(full, "market_data", "price_change_percentage_24h")
		return fmt.Sprintf("Trend snapshot for %s: 24h change %.2f%%", strings.ToUpper(symbol), change), nil
	}
	// compute simple trend description
	change := md.Change24h
	trend := "neutral"
	if change >= 5 {
		trend = "strong bullish momentum"
	} else if change >= 1 {
		trend = "bullish momentum"
	} else if change <= -5 {
		trend = "sharp drop / bearish"
	} else if change <= -1 {
		trend = "bearish momentum"
	}
	return fmt.Sprintf("Trend snapshot for %s: %s (24h %+0.2f%%)", strings.ToUpper(symbol), trend, change), nil
}
