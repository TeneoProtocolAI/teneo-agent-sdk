package modules

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// StartXScanner runs a scanner loop. It sends detections into out channel
// signature:
// ctx context.Context
// intervalSec int
// kols []string
// bearer string (for real mode; if empty, stay in mock mode)
// source string
// mock bool
// out chan<- Detection
func StartXScanner(ctx context.Context, intervalSec int, kols []string, bearer string, source string, mock bool, out chan<- Detection) {
	log.Printf("[xscanner] Starting scanner (mock=%v, interval=%ds, KOLs=%v, source=%s)", mock, intervalSec, kols, source)
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	rand.Seed(time.Now().UnixNano())

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[xscanner] Stopped.")
			return
		case <-ticker.C:
			// produce one mock detection per tick when mock==true
			if mock {
				d := generateMockDetection(kols, source)
				if d.Text != "" {
					out <- d
				}
				continue
			}

			// REAL mode placeholder: not implemented (must use bearer token)
			if bearer == "" {
				log.Println("[xscanner] WARNING: real mode requested but no bearer token provided; skipping")
				continue
			}

			// TODO: implement real fetch using X/Twitter API with rate-limits and parsing
			log.Println("[xscanner] real mode requested but not implemented yet.")
		}
	}
}

func generateMockDetection(kols []string, source string) Detection {
	if len(kols) == 0 {
		return Detection{}
	}
	kol := kols[rand.Intn(len(kols))]
	tokenList := []string{"SOL", "BTC", "DOGE", "PEPE", "BONK", "ABC"}
	token := tokenList[rand.Intn(len(tokenList))]
	msg := fmt.Sprintf("KOL %s mentioned %s", kol, token)
	link := "https://twitter.com/" + strings.ToLower(kol)
	return Detection{
		Text:      msg,
		Link:      link,
		Source:    source,
		Timestamp: TimeNowUTC(),
		KOL:       kol,
		Token:     token,
		Confidence: rand.Float64()*0.6 + 0.4,
	}
}
