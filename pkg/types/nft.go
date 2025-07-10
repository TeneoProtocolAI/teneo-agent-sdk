package types

import (
	"math/big"
	"time"
)

// MintRequest represents a request to mint an agent NFT
type MintRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Capabilities   []string `json:"capabilities"`
	ContactInfo    string   `json:"contact_info"`
	PricingModel   string   `json:"pricing_model"`
	InterfaceType  string   `json:"interface_type"`
	ResponseFormat string   `json:"response_format"`
	Version        string   `json:"version"`
	SDKVersion     string   `json:"sdk_version"`
	ImageURI       string   `json:"image_uri,omitempty"`
}

// Validate validates the mint request
func (r *MintRequest) Validate() *ValidationResult {
	var errors []string

	if r.Name == "" {
		errors = append(errors, "name is required")
	}

	if len(r.Name) > 50 {
		errors = append(errors, "name must be 50 characters or less")
	}

	if len(r.Capabilities) == 0 {
		errors = append(errors, "at least one capability is required")
	}

	if len(r.Capabilities) > 20 {
		errors = append(errors, "maximum 20 capabilities allowed")
	}

	if r.Version == "" {
		errors = append(errors, "version is required")
	}

	if r.SDKVersion == "" {
		errors = append(errors, "sdk_version is required")
	}

	// Validate interface type
	validInterfaces := []string{
		InterfaceTypeNaturalLanguage,
		InterfaceTypeAPI,
		InterfaceTypeCLI,
		InterfaceTypeWebSocket,
		InterfaceTypeGRPC,
	}
	
	if r.InterfaceType != "" {
		valid := false
		for _, validType := range validInterfaces {
			if r.InterfaceType == validType {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, "invalid interface type")
		}
	}

	// Validate response format
	validFormats := []string{
		ResponseFormatJSON,
		ResponseFormatText,
		ResponseFormatStructured,
		ResponseFormatXML,
		ResponseFormatYAML,
	}
	
	if r.ResponseFormat != "" {
		valid := false
		for _, validFormat := range validFormats {
			if r.ResponseFormat == validFormat {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, "invalid response format")
		}
	}

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// MintResult represents the result of a mint operation
type MintResult struct {
	Success      bool   `json:"success"`
	TokenID      string `json:"token_id"`
	TxHash       string `json:"tx_hash"`
	BlockNumber  uint64 `json:"block_number"`
	ContractAddr string `json:"contract_address"`
	Error        string `json:"error,omitempty"`
}

// BusinessCard represents an agent's NFT business card
type BusinessCard struct {
	TokenID      *big.Int      `json:"token_id"`
	Owner        string        `json:"owner"`
	ContractAddr string        `json:"contract_address"`
	Metadata     AgentMetadata `json:"metadata"`
}

// AgentMetadata represents the metadata stored in an agent's NFT
type AgentMetadata struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Capabilities   []string   `json:"capabilities"`
	ContactInfo    string     `json:"contact_info"`
	PricingModel   string     `json:"pricing_model"`
	InterfaceType  string     `json:"interface_type"`
	ResponseFormat string     `json:"response_format"`
	CreatedAt      *big.Int   `json:"created_at"`
	IsActive       bool       `json:"is_active"`
	Version        string     `json:"version"`
	SDKVersion     string     `json:"sdk_version"`
	ImageURI       string     `json:"image_uri,omitempty"`
}

// FoundationApprovalResult represents the result of a foundation approval request
type FoundationApprovalResult struct {
	Approved         bool      `json:"approved"`
	ApprovalID       string    `json:"approval_id,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	Signature        string    `json:"signature,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	FoundationSigner string    `json:"foundation_signer,omitempty"`
}

// NFTSearchRequest represents a request to search for NFTs
type NFTSearchRequest struct {
	Capabilities []string `json:"capabilities,omitempty"`
	Name         string   `json:"name,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// NFTSearchResult represents the result of an NFT search
type NFTSearchResult struct {
	Results    []*BusinessCard `json:"results"`
	Total      int             `json:"total"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	HasMore    bool            `json:"has_more"`
}

// ContractInfo represents information about the NFT contract
type ContractInfo struct {
	Address         string `json:"address"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	TotalSupply     string `json:"total_supply"`
	FoundationSigner string `json:"foundation_signer"`
	ChainID         int64  `json:"chain_id"`
	BlockNumber     uint64 `json:"block_number"`
}

// NFTTransferEvent represents an NFT transfer event
type NFTTransferEvent struct {
	From        string `json:"from"`
	To          string `json:"to"`
	TokenID     string `json:"token_id"`
	TxHash      string `json:"tx_hash"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

// AgentRegistrationEvent represents an agent registration event
type AgentRegistrationEvent struct {
	TokenID     string `json:"token_id"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	TxHash      string `json:"tx_hash"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

// AgentUpdateEvent represents an agent update event
type AgentUpdateEvent struct {
	TokenID     string `json:"token_id"`
	Name        string `json:"name"`
	UpdateType  string `json:"update_type"`
	TxHash      string `json:"tx_hash"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

// Constants for update types
const (
	UpdateTypeMetadata = "metadata"
	UpdateTypeActive   = "active"
	UpdateTypeCapabilities = "capabilities"
)

// NFTConfig represents configuration for NFT operations
type NFTConfig struct {
	ContractAddress    string `json:"contract_address"`
	FoundationSigner   string `json:"foundation_signer"`
	ChainID            int64  `json:"chain_id"`
	GasLimit           uint64 `json:"gas_limit"`
	GasPrice           string `json:"gas_price"`
	ConfirmationBlocks int    `json:"confirmation_blocks"`
}

// DefaultNFTConfig returns a default NFT configuration for PEAQ mainnet
func DefaultNFTConfig() *NFTConfig {
	return &NFTConfig{
		ContractAddress:    "0x2257A3993419b295EC062bC59C22c8A4EAA358A1",
		FoundationSigner:   "0xE0E039D10D6CEa83C7DAedb179B0Cfc75e0B0E66",
		ChainID:            3338,
		GasLimit:           500000,
		GasPrice:           "20000000000", // 20 gwei
		ConfirmationBlocks: 3,
	}
} 