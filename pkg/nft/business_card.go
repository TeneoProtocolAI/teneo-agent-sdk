package nft

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BusinessCardManager handles NFT business card operations
type BusinessCardManager struct {
	client            *ethclient.Client
	contract          *AgentBusinessCardV2
	privateKey        *ecdsa.PrivateKey
	fromAddress       common.Address
	contractAddr      common.Address
	foundationService *auth.FoundationSignatureService
}

// NewBusinessCardManager creates a new business card manager
func NewBusinessCardManager(rpcURL, contractAddress, privateKey string) (*BusinessCardManager, error) {
	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// Parse private key
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get address from private key
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)

	// Parse contract address
	contractAddr := common.HexToAddress(contractAddress)

	// Create contract instance
	contract, err := NewAgentBusinessCardV2(contractAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	// Initialize foundation signature service (in production, this would be a remote service)
	foundationService, err := auth.NewFoundationSignatureService(
		"e0e039d10d6cea83c7daedb179b0cfc75e0b0e66abc123def456789abcdef0123", // Foundation private key
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize foundation service: %w", err)
	}

	return &BusinessCardManager{
		client:            client,
		contract:          contract,
		privateKey:        key,
		fromAddress:       fromAddress,
		contractAddr:      contractAddr,
		foundationService: foundationService,
	}, nil
}

// MintAgentCard mints a new agent business card NFT
func (m *BusinessCardManager) MintAgentCard(ctx context.Context, request *types.MintRequest) (*types.BusinessCard, error) {
	log.Printf("🎨 Minting NFT business card for agent: %s", request.Name)

	// Validate request
	if validation := request.Validate(); !validation.IsValid {
		return nil, fmt.Errorf("invalid mint request: %v", validation.Errors)
	}

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(m.privateKey, big.NewInt(3338)) // PEAQ mainnet
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Set gas limit
	auth.GasLimit = uint64(500000)

	// Execute mint transaction
	tx, err := m.contract.MintAgentCard(
		auth,
		request.Name,
		request.Description,
		types.CapabilityNames(request.Capabilities),
		request.ContactInfo,
		request.PricingModel,
		request.InterfaceType,
		request.ResponseFormat,
		request.Version,
		request.SDKVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute mint transaction: %w", err)
	}

	log.Printf("🔄 Transaction sent: %s", tx.Hash().Hex())

	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, m.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status != 1 {
		return nil, fmt.Errorf("transaction failed")
	}

	log.Printf("✅ NFT minted successfully! Block: %d", receipt.BlockNumber.Uint64())

	// Get the minted token ID from the transaction receipt
	tokenID := big.NewInt(0)
	for _, vLog := range receipt.Logs {
		if len(vLog.Topics) > 0 && vLog.Topics[0].Hex() == "0x4d7ad63c7c2d79e6c8b3d7c7f9e8b0a4b2e6c3d1a5f8b9e0c4d7a2b5c8e1f4" {
			tokenID = new(big.Int).SetBytes(vLog.Topics[1][:])
			break
		}
	}

	// If we couldn't extract token ID from logs, get it from the contract
	if tokenID.Cmp(big.NewInt(0)) == 0 {
		tokenID, err = m.contract.OwnerToTokenId(&bind.CallOpts{Context: ctx}, m.fromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get token ID: %w", err)
		}
	}

	// Create business card result
	businessCard := &types.BusinessCard{
		TokenID:      tokenID,
		Owner:        m.fromAddress.Hex(),
		ContractAddr: m.contractAddr.Hex(),
		Metadata: types.AgentMetadata{
			Name:           request.Name,
			Description:    request.Description,
			Capabilities:   request.Capabilities,
			ContactInfo:    request.ContactInfo,
			PricingModel:   request.PricingModel,
			InterfaceType:  request.InterfaceType,
			ResponseFormat: request.ResponseFormat,
			Version:        request.Version,
			SDKVersion:     request.SDKVersion,
			IsActive:       true,
		},
	}

	return businessCard, nil
}

// GetAgentByOwner retrieves an agent's business card by owner address
func (m *BusinessCardManager) GetAgentByOwner(ctx context.Context, ownerAddress string) (*types.BusinessCard, error) {
	log.Printf("📖 Reading NFT business card for owner: %s", ownerAddress)

	owner := common.HexToAddress(ownerAddress)

	// Get token ID for owner
	tokenID, err := m.contract.OwnerToTokenId(&bind.CallOpts{Context: ctx}, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get token ID for owner: %w", err)
	}

	if tokenID.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("no NFT found for owner %s", ownerAddress)
	}

	// Get agent metadata
	metadata, err := m.contract.GetAgentByOwner(&bind.CallOpts{Context: ctx}, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent metadata: %w", err)
	}

	// Create business card
	businessCard := &types.BusinessCard{
		TokenID:      tokenID,
		Owner:        ownerAddress,
		ContractAddr: m.contractAddr.Hex(),
		Metadata: types.AgentMetadata{
			Name:           metadata.Name,
			Description:    metadata.Description,
			Capabilities:   types.CapabilitiesFromStrings(metadata.Capabilities),
			ContactInfo:    metadata.ContactInfo,
			PricingModel:   metadata.PricingModel,
			InterfaceType:  metadata.InterfaceType,
			ResponseFormat: metadata.ResponseFormat,
			CreatedAt:      metadata.CreatedAt,
			IsActive:       metadata.IsActive,
			Version:        metadata.Version,
			SDKVersion:     metadata.SdkVersion,
		},
	}

	return businessCard, nil
}

// UpdateAgentMetadata updates the agent's metadata
func (m *BusinessCardManager) UpdateAgentMetadata(ctx context.Context, description, contactInfo, pricingModel, version string) error {
	log.Printf("✏️ Updating agent metadata...")

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(m.privateKey, big.NewInt(3338))
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.GasLimit = uint64(200000)

	// Execute update transaction
	tx, err := m.contract.UpdateAgentMetadata(
		auth,
		description,
		contactInfo,
		pricingModel,
		version,
	)
	if err != nil {
		return fmt.Errorf("failed to execute update transaction: %w", err)
	}

	log.Printf("🔄 Update transaction sent: %s", tx.Hash().Hex())

	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, m.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("update transaction failed")
	}

	log.Printf("✅ Agent metadata updated successfully!")
	return nil
}

// SetAgentActive sets the agent's active status
func (m *BusinessCardManager) SetAgentActive(ctx context.Context, active bool) error {
	log.Printf("🔄 Setting agent active status to: %v", active)

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(m.privateKey, big.NewInt(3338))
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.GasLimit = uint64(100000)

	// Execute set active transaction
	tx, err := m.contract.SetAgentActive(auth, active)
	if err != nil {
		return fmt.Errorf("failed to execute set active transaction: %w", err)
	}

	log.Printf("🔄 Set active transaction sent: %s", tx.Hash().Hex())

	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, m.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("set active transaction failed")
	}

	log.Printf("✅ Agent active status updated successfully!")
	return nil
}

// GetAgentsByCapability retrieves agents that have a specific capability
func (m *BusinessCardManager) GetAgentsByCapability(ctx context.Context, capability string) ([]*big.Int, error) {
	log.Printf("🔍 Searching for agents with capability: %s", capability)

	tokenIDs, err := m.contract.GetAgentsByCapability(&bind.CallOpts{Context: ctx}, capability)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents by capability: %w", err)
	}

	log.Printf("✅ Found %d agents with capability '%s'", len(tokenIDs), capability)
	return tokenIDs, nil
}

// Close closes the connection to the Ethereum client
func (m *BusinessCardManager) Close() {
	if m.client != nil {
		m.client.Close()
	}
}

// GetContractAddress returns the contract address
func (m *BusinessCardManager) GetContractAddress() string {
	return m.contractAddr.Hex()
}

// GetOwnerAddress returns the owner address
func (m *BusinessCardManager) GetOwnerAddress() string {
	return m.fromAddress.Hex()
}

// hexToBytes converts a hex string to bytes
func hexToBytes(hex string) ([]byte, error) {
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}

	if len(hex)%2 != 0 {
		hex = "0" + hex
	}

	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		var b byte
		if _, err := fmt.Sscanf(hex[i:i+2], "%02x", &b); err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}

	return bytes, nil
}

// SimulateFoundationApproval simulates the foundation approval process
func (m *BusinessCardManager) SimulateFoundationApproval(request *types.MintRequest) (*types.FoundationApprovalResult, error) {
	log.Printf("🔧 Simulating foundation approval process for agent: %s", request.Name)

	// In a real implementation, this would:
	// 1. Send request to foundation backend
	// 2. Foundation validates the request
	// 3. Foundation returns approval with signature

	// For simulation, we'll approve if basic requirements are met
	if len(request.Capabilities) == 0 {
		return &types.FoundationApprovalResult{
			Approved: false,
			Reason:   "No capabilities specified",
		}, nil
	}

	if request.Name == "" {
		return &types.FoundationApprovalResult{
			Approved: false,
			Reason:   "Agent name is required",
		}, nil
	}

	// Simulate approval
	return &types.FoundationApprovalResult{
		Approved:         true,
		ApprovalID:       fmt.Sprintf("approval_%d", time.Now().Unix()),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		FoundationSigner: m.foundationService.GetAddress(),
	}, nil
}
