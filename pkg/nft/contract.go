package nft

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TeneoAgentNFT represents the ABI methods we need for minting
// This is a simplified version - in production, you'd use abigen to generate full bindings

// Contract ABI methods
const (
	MethodMint         = "mint"
	MethodNonces       = "nonces"
	MethodTokenURI     = "tokenURI"
	MethodUpdateTokenURI = "updateTokenURI"
	MethodMintPrice    = "mintPrice"
	MethodHasAccess    = "hasAccess"
)

// MintMethod represents the mint function signature
// mint(address to, bytes signature) payable returns (uint256)
type MintMethod struct {
	To        common.Address
	Signature []byte
}

// NonceMethod represents the nonces function signature
// nonces(address) view returns (uint256)
type NonceMethod struct {
	Address common.Address
}

// TokenURIMethod represents the tokenURI function signature
// tokenURI(uint256 tokenId) view returns (string)
type TokenURIMethod struct {
	TokenID *big.Int
}

// UpdateTokenURIMethod represents the updateTokenURI function signature
// updateTokenURI(uint256 tokenId, string newURI)
type UpdateTokenURIMethod struct {
	TokenID *big.Int
	NewURI  string
}

// HasAccessMethod represents the hasAccess function signature
// hasAccess(address user) view returns (bool)
type HasAccessMethod struct {
	User common.Address
}

// GetMintMethodID returns the method ID for the mint function
func GetMintMethodID() []byte {
	// mint(address,bytes)
	return []byte{0x7c, 0x6a, 0xec, 0x61} // First 4 bytes of keccak256("mint(address,bytes)")
}

// GetNoncesMethodID returns the method ID for the nonces function
func GetNoncesMethodID() []byte {
	// nonces(address)
	return []byte{0x7e, 0xcb, 0xe0, 0x00} // First 4 bytes of keccak256("nonces(address)")
}

// ParseABI returns a simplified ABI for the methods we need
func ParseABI() (abi.ABI, error) {
	const abiJSON = `[
		{
			"name": "mint",
			"type": "function",
			"inputs": [
				{"name": "to", "type": "address"},
				{"name": "signature", "type": "bytes"}
			],
			"outputs": [{"name": "tokenID", "type": "uint256"}],
			"stateMutability": "payable"
		},
		{
			"name": "nonces",
			"type": "function",
			"inputs": [
				{"name": "address", "type": "address"}
			],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "mintPrice",
			"type": "function",
			"inputs": [],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "hasAccess",
			"type": "function",
			"inputs": [
				{"name": "user", "type": "address"}
			],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "view"
		},
		{
			"name": "tokenURI",
			"type": "function",
			"inputs": [
				{"name": "tokenId", "type": "uint256"}
			],
			"outputs": [{"name": "", "type": "string"}],
			"stateMutability": "view"
		},
		{
			"name": "updateTokenURI",
			"type": "function",
			"inputs": [
				{"name": "tokenId", "type": "uint256"},
				{"name": "newURI", "type": "string"}
			],
			"outputs": [],
			"stateMutability": "nonpayable"
		}
	]`
	
	return abi.JSON(strings.NewReader(abiJSON))
}

// DefaultMintPrice returns the default mint price (2 PEAQ)
func DefaultMintPrice() *big.Int {
	mintPrice, _ := new(big.Int).SetString("2000000000000000000", 10)
	return mintPrice
}