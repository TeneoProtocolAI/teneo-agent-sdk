package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
)

// Manager handles authentication for Teneo agents
type Manager struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// NewManager creates a new authentication manager
func NewManager(privateKeyHex string) (*Manager, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &Manager{
		privateKey: privateKey,
		address:    address,
	}, nil
}

// GenerateToken generates a JWT token for the given address
func (m *Manager) GenerateToken(address string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"address": address,
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(), // 24 hour expiration
		"iss":     "teneo-agent-sdk",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Use the private key as the signing key (simplified approach)
	// In production, you'd use a proper JWT secret
	signingKey := crypto.Keccak256(crypto.FromECDSA(m.privateKey))

	return token.SignedString(signingKey)
}

// ValidateToken validates a JWT token
func (m *Manager) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	signingKey := crypto.Keccak256(crypto.FromECDSA(m.privateKey))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// SignMessage signs a message with the agent's private key
func (m *Manager) SignMessage(message string) (string, error) {
	hash := accounts.TextHash([]byte(message))
	signature, err := crypto.Sign(hash, m.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Adjust recovery ID for Ethereum compatibility
	signature[64] += 27

	// Use hexutil.Encode to include "0x" prefix, matching x-agent format
	return hexutil.Encode(signature), nil
}

// VerifySignature verifies a signature against a message and address
func (m *Manager) VerifySignature(message, signature, address string) (bool, error) {
	// Decode signature
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Hash the message
	hash := accounts.TextHash([]byte(message))

	// Recover public key from signature
	if len(sig) == 65 {
		// Adjust recovery ID for Ethereum
		sig[64] -= 27
	}

	pubkey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubkey)
	expectedAddr := common.HexToAddress(address)

	return recoveredAddr == expectedAddr, nil
}

// GetAddress returns the Ethereum address associated with this manager
func (m *Manager) GetAddress() string {
	return m.address.Hex()
}

// GenerateNonce generates a random nonce for authentication
func (m *Manager) GenerateNonce() (string, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return hex.EncodeToString(nonce), nil
}

// CreateAuthChallenge creates an authentication challenge
func (m *Manager) CreateAuthChallenge(address string) (*AuthChallenge, error) {
	nonce, err := m.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	challenge := &AuthChallenge{
		Address:   address,
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(), // 5 minute expiration
	}

	return challenge, nil
}

// ValidateAuthChallenge validates an authentication challenge response
func (m *Manager) ValidateAuthChallenge(challenge *AuthChallenge, signature string) (bool, error) {
	// Check if challenge has expired
	if time.Now().Unix() > challenge.ExpiresAt {
		return false, fmt.Errorf("challenge has expired")
	}

	// Create message to verify
	message := fmt.Sprintf("Teneo Agent Authentication\nAddress: %s\nNonce: %s\nTimestamp: %d",
		challenge.Address, challenge.Nonce, challenge.Timestamp)

	// Verify signature
	return m.VerifySignature(message, signature, challenge.Address)
}

// AuthChallenge represents an authentication challenge
type AuthChallenge struct {
	Address   string `json:"address"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	ExpiresAt int64  `json:"expires_at"`
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	Success   bool   `json:"success"`
	Token     string `json:"token,omitempty"`
	Address   string `json:"address,omitempty"`
	Error     string `json:"error,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

// SignatureProvider interface for different signature providers
type SignatureProvider interface {
	SignMessage(message string) (string, error)
	GetAddress() string
}

// FoundationSignatureService simulates the foundation signature service
type FoundationSignatureService struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// NewFoundationSignatureService creates a new foundation signature service
func NewFoundationSignatureService(privateKeyHex string) (*FoundationSignatureService, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &FoundationSignatureService{
		privateKey: privateKey,
		address:    address,
	}, nil
}

// SignMintRequest signs a mint request for NFT creation
func (f *FoundationSignatureService) SignMintRequest(
	userAddress string,
	name string,
	description string,
	capabilities []string,
	contactInfo string,
	pricingModel string,
	interfaceType string,
	responseFormat string,
	version string,
	sdkVersion string,
	nonce uint64,
) (string, error) {
	// Create the message hash (same as in the smart contract)
	hash := crypto.Keccak256(
		[]byte(userAddress),
		[]byte(name),
		[]byte(description),
		[]byte(fmt.Sprintf("%v", capabilities)),
		[]byte(contactInfo),
		[]byte(pricingModel),
		[]byte(interfaceType),
		[]byte(responseFormat),
		[]byte(version),
		[]byte(sdkVersion),
		[]byte(fmt.Sprintf("%d", nonce)),
	)

	// Sign the hash
	signature, err := crypto.Sign(hash, f.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign mint request: %w", err)
	}

	return hex.EncodeToString(signature), nil
}

// GetAddress returns the foundation signer address
func (f *FoundationSignatureService) GetAddress() string {
	return f.address.Hex()
}
