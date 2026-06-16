package deploy

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// SDKAuthMessagePrefix is the prefix used for signing auth challenges
	SDKAuthMessagePrefix = "Teneo SDK auth: "
)

// Authenticator handles SDK authentication with the backend
type Authenticator struct {
	privateKey *ecdsa.PrivateKey
	address    string
	client     *HTTPClient
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(privateKeyHex string, client *HTTPClient) (*Authenticator, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Derive address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to derive public key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return &Authenticator{
		privateKey: privateKey,
		address:    address,
		client:     client,
	}, nil
}

// Authenticate performs the full challenge-response authentication flow
func (a *Authenticator) Authenticate() (sessionToken string, expiresAt int64, err error) {
	// Step 1: Request challenge
	challengeResp, err := a.client.RequestChallenge(a.address)
	if err != nil {
		return "", 0, fmt.Errorf("failed to request challenge: %w", err)
	}

	// Step 2: Sign challenge
	signature, err := a.SignChallenge(challengeResp.Challenge)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Step 3: Verify signature
	verifyResp, err := a.client.VerifySignature(a.address, challengeResp.Challenge, signature)
	if err != nil {
		return "", 0, fmt.Errorf("failed to verify signature: %w", err)
	}

	return verifyResp.SessionToken, verifyResp.ExpiresAt, nil
}

// SignChallenge signs a challenge with the private key
func (a *Authenticator) SignChallenge(challenge string) (string, error) {
	// Construct message with prefix
	message := SDKAuthMessagePrefix + challenge

	// Hash with Ethereum prefix
	hash := hashMessage([]byte(message))

	// Sign
	signature, err := crypto.Sign(hash, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Adjust v value for Ethereum (27/28 instead of 0/1)
	signature[64] += 27

	return hexutil.Encode(signature), nil
}

// GetAddress returns the wallet address
func (a *Authenticator) GetAddress() string {
	return a.address
}

// hashMessage hashes a message with the Ethereum signed message prefix
func hashMessage(data []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256([]byte(prefix), data)
}
