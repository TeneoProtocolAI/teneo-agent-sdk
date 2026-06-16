package deploy

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ChainClient handles on-chain operations
type ChainClient struct {
	client          *ethclient.Client
	contractAddress common.Address
	chainID         *big.Int
	privateKey      *ecdsa.PrivateKey
	address         common.Address
}

// MintResult contains the result of a mint operation
type MintResult struct {
	TokenID         uint64 `json:"token_id"`
	TxHash          string `json:"tx_hash,omitempty"`
	AgentID         string `json:"agent_id,omitempty"`
	Status          string `json:"status,omitempty"` // "MINTED", "ALREADY_OWNED", "UPDATE_REQUIRED"
	ContractAddress string `json:"contract_address,omitempty"`
	Message         string `json:"message,omitempty"`
}

// MintStatus constants
const (
	MintStatusMinted         = "MINTED"
	MintStatusAlreadyOwned   = "ALREADY_OWNED"
	MintStatusUpdateRequired = "UPDATE_REQUIRED"
	MintStatusUpdated        = "UPDATED"
)

// NewChainClient creates a new chain client
func NewChainClient(rpcEndpoint, contractAddress, chainIDStr, privateKeyHex string) (*ChainClient, error) {
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
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Parse chain ID
	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}

	// Connect to RPC
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	return &ChainClient{
		client:          client,
		contractAddress: common.HexToAddress(contractAddress),
		chainID:         chainID,
		privateKey:      privateKey,
		address:         address,
	}, nil
}

// Close closes the client connection
func (c *ChainClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// HasAccess checks if the wallet has access (owns an NFT from this contract)
func (c *ChainClient) HasAccess(ctx context.Context) (bool, error) {
	// ABI for hasAccess(address) -> bool
	hasAccessABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"user","type":"address"}],"name":"hasAccess","outputs":[{"name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`))
	if err != nil {
		return false, fmt.Errorf("failed to parse hasAccess ABI: %w", err)
	}

	// Pack call data
	data, err := hasAccessABI.Pack("hasAccess", c.address)
	if err != nil {
		return false, fmt.Errorf("failed to pack hasAccess call: %w", err)
	}

	// Make call
	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("failed to call hasAccess: %w", err)
	}

	// Unpack result
	var hasAccess bool
	if err := hasAccessABI.UnpackIntoInterface(&hasAccess, "hasAccess", result); err != nil {
		return false, fmt.Errorf("failed to unpack hasAccess result: %w", err)
	}

	return hasAccess, nil
}

// GetTokenID gets the token ID owned by the wallet (if any)
func (c *ChainClient) GetTokenID(ctx context.Context) (uint64, error) {
	// ABI for tokenOfOwnerByIndex(address, uint256) -> uint256
	tokenABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"owner","type":"address"},{"name":"index","type":"uint256"}],"name":"tokenOfOwnerByIndex","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`))
	if err != nil {
		return 0, fmt.Errorf("failed to parse tokenOfOwnerByIndex ABI: %w", err)
	}

	// Pack call data (index 0 = first token)
	data, err := tokenABI.Pack("tokenOfOwnerByIndex", c.address, big.NewInt(0))
	if err != nil {
		return 0, fmt.Errorf("failed to pack tokenOfOwnerByIndex call: %w", err)
	}

	// Make call
	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call tokenOfOwnerByIndex: %w", err)
	}

	// Unpack result
	var tokenID *big.Int
	if err := tokenABI.UnpackIntoInterface(&tokenID, "tokenOfOwnerByIndex", result); err != nil {
		return 0, fmt.Errorf("failed to unpack tokenOfOwnerByIndex result: %w", err)
	}

	return tokenID.Uint64(), nil
}

// ExecuteMint executes the on-chain mint transaction
func (c *ChainClient) ExecuteMint(ctx context.Context, signature string, mintPrice *big.Int) (*MintResult, error) {
	// Query mint price from contract if not provided
	if mintPrice == nil {
		var err error
		mintPrice, err = c.GetMintPrice(ctx)
		if err != nil {
			// Fallback to 2 PEAQ if contract call fails
			mintPrice = new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		}
	}

	// Check wallet balance
	balance, err := c.client.BalanceAt(ctx, c.address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance: %w", err)
	}
	if balance.Cmp(mintPrice) < 0 {
		return nil, fmt.Errorf("insufficient balance: have %s wei, need %s wei for mint", balance.String(), mintPrice.String())
	}

	// ABI for mint(address to, bytes signature)
	mintABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"to","type":"address"},{"name":"signature","type":"bytes"}],"name":"mint","outputs":[],"stateMutability":"payable","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"to","type":"address"},{"indexed":true,"name":"tokenId","type":"uint256"}],"name":"Minted","type":"event"}]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mint ABI: %w", err)
	}

	// Decode signature
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature format: %w", err)
	}

	// Pack call data
	data, err := mintABI.Pack("mint", c.address, sigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to pack mint call: %w", err)
	}

	// Get nonce
	nonce, err := c.client.PendingNonceAt(ctx, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas (also validates the tx won't revert)
	estimatedGas, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
		From:  c.address,
		To:    &c.contractAddress,
		Value: mintPrice,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("mint would revert: %w", err)
	}
	gasLimit := estimatedGas * 120 / 100 // 20% safety margin

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		c.contractAddress,
		mintPrice,
		gasLimit,
		gasPrice,
		data,
	)

	// Sign transaction
	signer := types.NewEIP155Signer(c.chainID)
	signedTx, err := types.SignTx(tx, signer, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := c.client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()

	// Wait for receipt with timeout
	receiptCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	receipt, err := c.waitForReceipt(receiptCtx, signedTx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction reverted")
	}

	// Extract token ID from Minted or Transfer event
	tokenID, err := c.ExtractTokenIDFromReceipt(receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract token ID: %w", err)
	}

	return &MintResult{
		TokenID: tokenID,
		TxHash:  txHash,
	}, nil
}

// waitForReceipt polls for transaction receipt
func (c *ChainClient) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			receipt, err := c.client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// Continue polling if receipt not found yet
		}
	}
}

// extractTokenIDFromReceipt extracts the token ID from the Minted event in the receipt
func (c *ChainClient) extractTokenIDFromReceipt(receipt *types.Receipt, contractABI *abi.ABI) (uint64, error) {
	// Find Minted event
	mintedEvent := contractABI.Events["Minted"]
	mintedSig := mintedEvent.ID

	for _, log := range receipt.Logs {
		if len(log.Topics) > 0 && log.Topics[0] == mintedSig {
			// Token ID is the second indexed parameter (Topics[2])
			if len(log.Topics) >= 3 {
				tokenID := new(big.Int).SetBytes(log.Topics[2].Bytes())
				return tokenID.Uint64(), nil
			}
		}
	}

	return 0, fmt.Errorf("Minted event not found in receipt")
}

// GetMintPrice queries the mintPrice from the contract
func (c *ChainClient) GetMintPrice(ctx context.Context) (*big.Int, error) {
	mintPriceABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"name":"mintPrice","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mintPrice ABI: %w", err)
	}

	data, err := mintPriceABI.Pack("mintPrice")
	if err != nil {
		return nil, fmt.Errorf("failed to pack mintPrice call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call mintPrice: %w", err)
	}

	var price *big.Int
	if err := mintPriceABI.UnpackIntoInterface(&price, "mintPrice", result); err != nil {
		return nil, fmt.Errorf("failed to unpack mintPrice result: %w", err)
	}

	return price, nil
}

// GetAddress returns the wallet address
func (c *ChainClient) GetAddress() string {
	return c.address.Hex()
}

// GetTransactionReceipt retrieves the receipt for a transaction
func (c *ChainClient) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	return receipt, nil
}

// ExtractTokenIDFromReceipt extracts the token ID from Minted or Transfer event in the receipt
func (c *ChainClient) ExtractTokenIDFromReceipt(receipt *types.Receipt) (uint64, error) {
	// Try Minted event first (custom event)
	mintABI, err := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"to","type":"address"},{"indexed":true,"name":"tokenId","type":"uint256"}],"name":"Minted","type":"event"}]`))
	if err != nil {
		return 0, fmt.Errorf("failed to parse Minted event ABI: %w", err)
	}

	tokenID, err := c.extractTokenIDFromReceipt(receipt, &mintABI)
	if err == nil {
		return tokenID, nil
	}

	// Fallback to Transfer event (standard ERC721)
	// Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	transferABI, err := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":true,"name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"}]`))
	if err != nil {
		return 0, fmt.Errorf("failed to parse Transfer event ABI: %w", err)
	}

	return c.extractTokenIDFromTransferEvent(receipt, &transferABI)
}

// extractTokenIDFromTransferEvent extracts token ID from standard ERC721 Transfer event
func (c *ChainClient) extractTokenIDFromTransferEvent(receipt *types.Receipt, contractABI *abi.ABI) (uint64, error) {
	transferEvent := contractABI.Events["Transfer"]
	transferSig := transferEvent.ID

	for _, log := range receipt.Logs {
		if len(log.Topics) > 0 && log.Topics[0] == transferSig {
			// Token ID is the fourth indexed parameter (Topics[3])
			if len(log.Topics) >= 4 {
				tokenID := new(big.Int).SetBytes(log.Topics[3].Bytes())
				return tokenID.Uint64(), nil
			}
		}
	}

	return 0, fmt.Errorf("Transfer event not found in receipt")
}
