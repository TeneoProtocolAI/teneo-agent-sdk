# SignalShield Analyst

SignalShield Analyst is a fast-response crypto market intelligence agent for the Teneo Protocol Chatroom.  
It detects early KOL calls, evaluates hype & sentiment, computes risk scores, and provides real-time market metrics powered by CoinGecko and Google Gemini.

## Features
- Early-call detection (mock/live)
- Hype scoring
- Sentiment analysis
- Risk evaluation
- Trend detection
- Influencer tracking (mock/live)
- Gemini AI reasoning
- Alert system
- Market metrics via CoinGecko (price, volume, marketcap)
- Full Teneo agent lifecycle (auth → register → websocket)

## Requirements
- Go 1.20+
- Google API Key (Generative Language API)
- Private key to sign Teneo challenges
- NFT tokenID registered in Teneo
- Internet connection

## Environment Variables
Create `.env` (never commit):

PRIVATE_KEY=<wallet_priv_key>
NFT_TOKEN_ID=366
OWNER_ADDRESS=0x...
GOOGLE_API_KEY=AIzaSy...
GOOGLE_MODEL=models/gemini-2.5-flash
COINGECKO_BASE_CURRENCY=https://api.coingecko.com/api/v3

MOCK_MODE=true
RATE_LIMIT_PER_MINUTE=30
ENABLE_FORWARD_OPENAI=false

## Running
go mod tidy
go run .
Health check:
curl http://localhost:8081/health

## Supported Commands
@signalshield-analyst hype sol
@signalshield-analyst sentiment eth
@signalshield-analyst riskcheck btc
@signalshield-analyst gecko pepe
@signalshield-analyst ai "explain risks of SOL in 3 bullets"
@signalshield-analyst alert BTC "touch support" 

## Troubleshooting
- API key invalid → re-export env variables
- 404 CoinGecko → symbol not mapped
- Gemini JSON error → malformed request (fixed in current version)

## License
MIT

