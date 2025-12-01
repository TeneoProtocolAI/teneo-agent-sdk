package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Small coin symbol -> coingecko id mapping for common tokens.
// Keys must be unique and lowercase.
var cgSymbolToID = map[string]string{
	"btc":   "bitcoin",
	"eth":   "ethereum",
	"bnb":   "binancecoin",
	"sol":   "solana",
	"matic": "matic-network",
	"ada":   "cardano",
	"doge":  "dogecoin",
	"usdt":  "tether",
	"usdc":  "usd-coin",
	"ltc":   "litecoin",
	"avax":  "avalanche-2",
	"dot":   "polkadot",
	"link":  "chainlink",
	"shib":  "shiba-inu",
	"uni":   "uniswap",
	"ftm":   "fantom",
	"atom":  "cosmos",
	"op":    "optimism",
	"arb":   "arbitrum",
}

// MarketData holds the values we extract from CoinGecko
type MarketData struct {
	ID           string
	Symbol       string
	PriceUSD     float64
	Change24h    float64 // percentage
	Volume24h    float64 // in USD
	MarketCapUSD float64
	RetrievedAt  time.Time
}

// cache entry
type cgCacheEntry struct {
	data      MarketData
	expiresAt time.Time
}

var (
	cgCache    = map[string]cgCacheEntry{}
	cgCacheMu  = sync.Mutex{}
	cacheTTL   = 30 * time.Second
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

// GetMarketData fetches market data for a symbol (e.g., "SOL", "BTC").
func GetMarketData(symbol string) (MarketData, error) {
	sym := strings.ToLower(strings.TrimSpace(symbol))
	// cache check
	cgCacheMu.Lock()
	if e, ok := cgCache[sym]; ok && time.Now().Before(e.expiresAt) {
		cgCacheMu.Unlock()
		return e.data, nil
	}
	cgCacheMu.Unlock()

	id, ok := cgSymbolToID[sym]
	if !ok {
		// try direct id fallback
		id = sym
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false", id)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return MarketData{}, fmt.Errorf("coingecko http err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return MarketData{}, fmt.Errorf("coingecko status %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return MarketData{}, fmt.Errorf("coingecko decode err: %w", err)
	}

	md := MarketData{
		ID:          id,
		Symbol:      sym,
		RetrievedAt: time.Now(),
	}

	if marketData, ok := body["market_data"].(map[string]interface{}); ok {
		if cp, ok := marketData["current_price"].(map[string]interface{}); ok {
			if usd, ok := cp["usd"].(float64); ok {
				md.PriceUSD = usd
			}
		}
		if ch, ok := marketData["price_change_percentage_24h"].(float64); ok {
			md.Change24h = ch
		}
		if vol, ok := marketData["total_volume"].(map[string]interface{}); ok {
			if v, ok := vol["usd"].(float64); ok {
				md.Volume24h = v
			}
		}
		if mc, ok := marketData["market_cap"].(map[string]interface{}); ok {
			if m, ok := mc["usd"].(float64); ok {
				md.MarketCapUSD = m
			}
		}
	}

	// save to cache
	cgCacheMu.Lock()
	cgCache[sym] = cgCacheEntry{
		data:      md,
		expiresAt: time.Now().Add(cacheTTL),
	}
	cgCacheMu.Unlock()

	return md, nil
}

// ComputeHypeScore builds a simple hype score [0..1] using change24h and volume/marketcap
func ComputeHypeScore(m MarketData) float64 {
	score := 0.0
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}
	score += clamp((m.Change24h+10)/40) * 0.6
	if m.MarketCapUSD > 0 {
		r := (m.Volume24h / m.MarketCapUSD)
		score += clamp(r*20) * 0.4
	}
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}
