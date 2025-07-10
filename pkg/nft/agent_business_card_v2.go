// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package nft

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// AgentBusinessCardAgentMetadata is an auto generated low-level Go binding around an user-defined struct.
type AgentBusinessCardAgentMetadata struct {
	Name           string
	Description    string
	Capabilities   []string
	ContactInfo    string
	PricingModel   string
	InterfaceType  string
	ResponseFormat string
	CreatedAt      *big.Int
	IsActive       bool
	Version        string
	SdkVersion     string
}

// AgentBusinessCardV2MetaData contains all meta data concerning the AgentBusinessCardV2 contract.
var AgentBusinessCardV2MetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"agentMetadata\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"contactInfo\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"pricingModel\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"interfaceType\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"responseFormat\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"createdAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isActive\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"sdkVersion\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"capabilityToAgents\",\"inputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAgentByOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structAgentBusinessCard.AgentMetadata\",\"components\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"capabilities\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"contactInfo\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"pricingModel\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"interfaceType\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"responseFormat\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"createdAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isActive\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"sdkVersion\",\"type\":\"string\",\"internalType\":\"string\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAgentsByCapability\",\"inputs\":[{\"name\":\"capability\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getApproved\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isApprovedForAll\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mintAgentCard\",\"inputs\":[{\"name\":\"_name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_capabilities\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"_contactInfo\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_pricingModel\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_interfaceType\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_responseFormat\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_sdkVersion\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"name\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ownerOf\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ownerToTokenId\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"safeTransferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"safeTransferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAgentActive\",\"inputs\":[{\"name\":\"_isActive\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setApprovalForAll\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"approved\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSoulboundEnabled\",\"inputs\":[{\"name\":\"_enabled\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"soulboundEnabled\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"symbol\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tokenURI\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateAgentMetadata\",\"inputs\":[{\"name\":\"_description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_contactInfo\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_pricingModel\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_version\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AgentActivated\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"isActive\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AgentRegistered\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AgentUpdated\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Approval\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"approved\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ApprovalForAll\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"approved\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transfer\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"ERC721IncorrectOwner\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721InsufficientApproval\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ERC721InvalidApprover\",\"inputs\":[{\"name\":\"approver\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721InvalidOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721InvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721InvalidReceiver\",\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721InvalidSender\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC721NonexistentToken\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
}

// AgentBusinessCardV2ABI is the input ABI used to generate the binding from.
// Deprecated: Use AgentBusinessCardV2MetaData.ABI instead.
var AgentBusinessCardV2ABI = AgentBusinessCardV2MetaData.ABI

// AgentBusinessCardV2 is an auto generated Go binding around an Ethereum contract.
type AgentBusinessCardV2 struct {
	AgentBusinessCardV2Caller     // Read-only binding to the contract
	AgentBusinessCardV2Transactor // Write-only binding to the contract
	AgentBusinessCardV2Filterer   // Log filterer for contract events
}

// AgentBusinessCardV2Caller is an auto generated read-only Go binding around an Ethereum contract.
type AgentBusinessCardV2Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentBusinessCardV2Transactor is an auto generated write-only Go binding around an Ethereum contract.
type AgentBusinessCardV2Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentBusinessCardV2Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AgentBusinessCardV2Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentBusinessCardV2Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AgentBusinessCardV2Session struct {
	Contract     *AgentBusinessCardV2 // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AgentBusinessCardV2CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AgentBusinessCardV2CallerSession struct {
	Contract *AgentBusinessCardV2Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// AgentBusinessCardV2TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AgentBusinessCardV2TransactorSession struct {
	Contract     *AgentBusinessCardV2Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// AgentBusinessCardV2Raw is an auto generated low-level Go binding around an Ethereum contract.
type AgentBusinessCardV2Raw struct {
	Contract *AgentBusinessCardV2 // Generic contract binding to access the raw methods on
}

// AgentBusinessCardV2CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AgentBusinessCardV2CallerRaw struct {
	Contract *AgentBusinessCardV2Caller // Generic read-only contract binding to access the raw methods on
}

// AgentBusinessCardV2TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AgentBusinessCardV2TransactorRaw struct {
	Contract *AgentBusinessCardV2Transactor // Generic write-only contract binding to access the raw methods on
}

// NewAgentBusinessCardV2 creates a new instance of AgentBusinessCardV2, bound to a specific deployed contract.
func NewAgentBusinessCardV2(address common.Address, backend bind.ContractBackend) (*AgentBusinessCardV2, error) {
	contract, err := bindAgentBusinessCardV2(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2{AgentBusinessCardV2Caller: AgentBusinessCardV2Caller{contract: contract}, AgentBusinessCardV2Transactor: AgentBusinessCardV2Transactor{contract: contract}, AgentBusinessCardV2Filterer: AgentBusinessCardV2Filterer{contract: contract}}, nil
}

// NewAgentBusinessCardV2Caller creates a new read-only instance of AgentBusinessCardV2, bound to a specific deployed contract.
func NewAgentBusinessCardV2Caller(address common.Address, caller bind.ContractCaller) (*AgentBusinessCardV2Caller, error) {
	contract, err := bindAgentBusinessCardV2(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2Caller{contract: contract}, nil
}

// NewAgentBusinessCardV2Transactor creates a new write-only instance of AgentBusinessCardV2, bound to a specific deployed contract.
func NewAgentBusinessCardV2Transactor(address common.Address, transactor bind.ContractTransactor) (*AgentBusinessCardV2Transactor, error) {
	contract, err := bindAgentBusinessCardV2(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2Transactor{contract: contract}, nil
}

// NewAgentBusinessCardV2Filterer creates a new log filterer instance of AgentBusinessCardV2, bound to a specific deployed contract.
func NewAgentBusinessCardV2Filterer(address common.Address, filterer bind.ContractFilterer) (*AgentBusinessCardV2Filterer, error) {
	contract, err := bindAgentBusinessCardV2(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2Filterer{contract: contract}, nil
}

// bindAgentBusinessCardV2 binds a generic wrapper to an already deployed contract.
func bindAgentBusinessCardV2(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AgentBusinessCardV2MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentBusinessCardV2 *AgentBusinessCardV2Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentBusinessCardV2.Contract.AgentBusinessCardV2Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentBusinessCardV2 *AgentBusinessCardV2Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.AgentBusinessCardV2Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentBusinessCardV2 *AgentBusinessCardV2Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.AgentBusinessCardV2Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentBusinessCardV2.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.contract.Transact(opts, method, params...)
}

// AgentMetadata is a free data retrieval call binding the contract method 0x47a95efb.
//
// Solidity: function agentMetadata(uint256 ) view returns(string name, string description, string contactInfo, string pricingModel, string interfaceType, string responseFormat, uint256 createdAt, bool isActive, string version, string sdkVersion)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) AgentMetadata(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Name           string
	Description    string
	ContactInfo    string
	PricingModel   string
	InterfaceType  string
	ResponseFormat string
	CreatedAt      *big.Int
	IsActive       bool
	Version        string
	SdkVersion     string
}, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "agentMetadata", arg0)

	outstruct := new(struct {
		Name           string
		Description    string
		ContactInfo    string
		PricingModel   string
		InterfaceType  string
		ResponseFormat string
		CreatedAt      *big.Int
		IsActive       bool
		Version        string
		SdkVersion     string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Name = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Description = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.ContactInfo = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.PricingModel = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.InterfaceType = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.ResponseFormat = *abi.ConvertType(out[5], new(string)).(*string)
	outstruct.CreatedAt = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.IsActive = *abi.ConvertType(out[7], new(bool)).(*bool)
	outstruct.Version = *abi.ConvertType(out[8], new(string)).(*string)
	outstruct.SdkVersion = *abi.ConvertType(out[9], new(string)).(*string)

	return *outstruct, err

}

// AgentMetadata is a free data retrieval call binding the contract method 0x47a95efb.
//
// Solidity: function agentMetadata(uint256 ) view returns(string name, string description, string contactInfo, string pricingModel, string interfaceType, string responseFormat, uint256 createdAt, bool isActive, string version, string sdkVersion)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) AgentMetadata(arg0 *big.Int) (struct {
	Name           string
	Description    string
	ContactInfo    string
	PricingModel   string
	InterfaceType  string
	ResponseFormat string
	CreatedAt      *big.Int
	IsActive       bool
	Version        string
	SdkVersion     string
}, error) {
	return _AgentBusinessCardV2.Contract.AgentMetadata(&_AgentBusinessCardV2.CallOpts, arg0)
}

// AgentMetadata is a free data retrieval call binding the contract method 0x47a95efb.
//
// Solidity: function agentMetadata(uint256 ) view returns(string name, string description, string contactInfo, string pricingModel, string interfaceType, string responseFormat, uint256 createdAt, bool isActive, string version, string sdkVersion)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) AgentMetadata(arg0 *big.Int) (struct {
	Name           string
	Description    string
	ContactInfo    string
	PricingModel   string
	InterfaceType  string
	ResponseFormat string
	CreatedAt      *big.Int
	IsActive       bool
	Version        string
	SdkVersion     string
}, error) {
	return _AgentBusinessCardV2.Contract.AgentMetadata(&_AgentBusinessCardV2.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.BalanceOf(&_AgentBusinessCardV2.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.BalanceOf(&_AgentBusinessCardV2.CallOpts, owner)
}

// CapabilityToAgents is a free data retrieval call binding the contract method 0x5cb532d0.
//
// Solidity: function capabilityToAgents(string , uint256 ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) CapabilityToAgents(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "capabilityToAgents", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CapabilityToAgents is a free data retrieval call binding the contract method 0x5cb532d0.
//
// Solidity: function capabilityToAgents(string , uint256 ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) CapabilityToAgents(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.CapabilityToAgents(&_AgentBusinessCardV2.CallOpts, arg0, arg1)
}

// CapabilityToAgents is a free data retrieval call binding the contract method 0x5cb532d0.
//
// Solidity: function capabilityToAgents(string , uint256 ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) CapabilityToAgents(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.CapabilityToAgents(&_AgentBusinessCardV2.CallOpts, arg0, arg1)
}

// GetAgentByOwner is a free data retrieval call binding the contract method 0x0381ee79.
//
// Solidity: function getAgentByOwner(address owner) view returns((string,string,string[],string,string,string,string,uint256,bool,string,string))
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) GetAgentByOwner(opts *bind.CallOpts, owner common.Address) (AgentBusinessCardAgentMetadata, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "getAgentByOwner", owner)

	if err != nil {
		return *new(AgentBusinessCardAgentMetadata), err
	}

	out0 := *abi.ConvertType(out[0], new(AgentBusinessCardAgentMetadata)).(*AgentBusinessCardAgentMetadata)

	return out0, err

}

// GetAgentByOwner is a free data retrieval call binding the contract method 0x0381ee79.
//
// Solidity: function getAgentByOwner(address owner) view returns((string,string,string[],string,string,string,string,uint256,bool,string,string))
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) GetAgentByOwner(owner common.Address) (AgentBusinessCardAgentMetadata, error) {
	return _AgentBusinessCardV2.Contract.GetAgentByOwner(&_AgentBusinessCardV2.CallOpts, owner)
}

// GetAgentByOwner is a free data retrieval call binding the contract method 0x0381ee79.
//
// Solidity: function getAgentByOwner(address owner) view returns((string,string,string[],string,string,string,string,uint256,bool,string,string))
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) GetAgentByOwner(owner common.Address) (AgentBusinessCardAgentMetadata, error) {
	return _AgentBusinessCardV2.Contract.GetAgentByOwner(&_AgentBusinessCardV2.CallOpts, owner)
}

// GetAgentsByCapability is a free data retrieval call binding the contract method 0xcc62097a.
//
// Solidity: function getAgentsByCapability(string capability) view returns(uint256[])
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) GetAgentsByCapability(opts *bind.CallOpts, capability string) ([]*big.Int, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "getAgentsByCapability", capability)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetAgentsByCapability is a free data retrieval call binding the contract method 0xcc62097a.
//
// Solidity: function getAgentsByCapability(string capability) view returns(uint256[])
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) GetAgentsByCapability(capability string) ([]*big.Int, error) {
	return _AgentBusinessCardV2.Contract.GetAgentsByCapability(&_AgentBusinessCardV2.CallOpts, capability)
}

// GetAgentsByCapability is a free data retrieval call binding the contract method 0xcc62097a.
//
// Solidity: function getAgentsByCapability(string capability) view returns(uint256[])
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) GetAgentsByCapability(capability string) ([]*big.Int, error) {
	return _AgentBusinessCardV2.Contract.GetAgentsByCapability(&_AgentBusinessCardV2.CallOpts, capability)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AgentBusinessCardV2.Contract.GetApproved(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AgentBusinessCardV2.Contract.GetApproved(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AgentBusinessCardV2.Contract.IsApprovedForAll(&_AgentBusinessCardV2.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AgentBusinessCardV2.Contract.IsApprovedForAll(&_AgentBusinessCardV2.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) Name() (string, error) {
	return _AgentBusinessCardV2.Contract.Name(&_AgentBusinessCardV2.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) Name() (string, error) {
	return _AgentBusinessCardV2.Contract.Name(&_AgentBusinessCardV2.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) Owner() (common.Address, error) {
	return _AgentBusinessCardV2.Contract.Owner(&_AgentBusinessCardV2.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) Owner() (common.Address, error) {
	return _AgentBusinessCardV2.Contract.Owner(&_AgentBusinessCardV2.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AgentBusinessCardV2.Contract.OwnerOf(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AgentBusinessCardV2.Contract.OwnerOf(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// OwnerToTokenId is a free data retrieval call binding the contract method 0x4814f035.
//
// Solidity: function ownerToTokenId(address ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) OwnerToTokenId(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "ownerToTokenId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OwnerToTokenId is a free data retrieval call binding the contract method 0x4814f035.
//
// Solidity: function ownerToTokenId(address ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) OwnerToTokenId(arg0 common.Address) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.OwnerToTokenId(&_AgentBusinessCardV2.CallOpts, arg0)
}

// OwnerToTokenId is a free data retrieval call binding the contract method 0x4814f035.
//
// Solidity: function ownerToTokenId(address ) view returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) OwnerToTokenId(arg0 common.Address) (*big.Int, error) {
	return _AgentBusinessCardV2.Contract.OwnerToTokenId(&_AgentBusinessCardV2.CallOpts, arg0)
}

// SoulboundEnabled is a free data retrieval call binding the contract method 0x298f6a54.
//
// Solidity: function soulboundEnabled() view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) SoulboundEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "soulboundEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SoulboundEnabled is a free data retrieval call binding the contract method 0x298f6a54.
//
// Solidity: function soulboundEnabled() view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SoulboundEnabled() (bool, error) {
	return _AgentBusinessCardV2.Contract.SoulboundEnabled(&_AgentBusinessCardV2.CallOpts)
}

// SoulboundEnabled is a free data retrieval call binding the contract method 0x298f6a54.
//
// Solidity: function soulboundEnabled() view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) SoulboundEnabled() (bool, error) {
	return _AgentBusinessCardV2.Contract.SoulboundEnabled(&_AgentBusinessCardV2.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AgentBusinessCardV2.Contract.SupportsInterface(&_AgentBusinessCardV2.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AgentBusinessCardV2.Contract.SupportsInterface(&_AgentBusinessCardV2.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) Symbol() (string, error) {
	return _AgentBusinessCardV2.Contract.Symbol(&_AgentBusinessCardV2.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) Symbol() (string, error) {
	return _AgentBusinessCardV2.Contract.Symbol(&_AgentBusinessCardV2.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Caller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _AgentBusinessCardV2.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) TokenURI(tokenId *big.Int) (string, error) {
	return _AgentBusinessCardV2.Contract.TokenURI(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AgentBusinessCardV2 *AgentBusinessCardV2CallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AgentBusinessCardV2.Contract.TokenURI(&_AgentBusinessCardV2.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.Approve(&_AgentBusinessCardV2.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.Approve(&_AgentBusinessCardV2.TransactOpts, to, tokenId)
}

// MintAgentCard is a paid mutator transaction binding the contract method 0x978f6e18.
//
// Solidity: function mintAgentCard(string _name, string _description, string[] _capabilities, string _contactInfo, string _pricingModel, string _interfaceType, string _responseFormat, string _version, string _sdkVersion) returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) MintAgentCard(opts *bind.TransactOpts, _name string, _description string, _capabilities []string, _contactInfo string, _pricingModel string, _interfaceType string, _responseFormat string, _version string, _sdkVersion string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "mintAgentCard", _name, _description, _capabilities, _contactInfo, _pricingModel, _interfaceType, _responseFormat, _version, _sdkVersion)
}

// MintAgentCard is a paid mutator transaction binding the contract method 0x978f6e18.
//
// Solidity: function mintAgentCard(string _name, string _description, string[] _capabilities, string _contactInfo, string _pricingModel, string _interfaceType, string _responseFormat, string _version, string _sdkVersion) returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) MintAgentCard(_name string, _description string, _capabilities []string, _contactInfo string, _pricingModel string, _interfaceType string, _responseFormat string, _version string, _sdkVersion string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.MintAgentCard(&_AgentBusinessCardV2.TransactOpts, _name, _description, _capabilities, _contactInfo, _pricingModel, _interfaceType, _responseFormat, _version, _sdkVersion)
}

// MintAgentCard is a paid mutator transaction binding the contract method 0x978f6e18.
//
// Solidity: function mintAgentCard(string _name, string _description, string[] _capabilities, string _contactInfo, string _pricingModel, string _interfaceType, string _responseFormat, string _version, string _sdkVersion) returns(uint256)
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) MintAgentCard(_name string, _description string, _capabilities []string, _contactInfo string, _pricingModel string, _interfaceType string, _responseFormat string, _version string, _sdkVersion string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.MintAgentCard(&_AgentBusinessCardV2.TransactOpts, _name, _description, _capabilities, _contactInfo, _pricingModel, _interfaceType, _responseFormat, _version, _sdkVersion)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) RenounceOwnership() (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.RenounceOwnership(&_AgentBusinessCardV2.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.RenounceOwnership(&_AgentBusinessCardV2.TransactOpts)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SafeTransferFrom(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SafeTransferFrom(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SafeTransferFrom0(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SafeTransferFrom0(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId, data)
}

// SetAgentActive is a paid mutator transaction binding the contract method 0xcf3ea0e6.
//
// Solidity: function setAgentActive(bool _isActive) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) SetAgentActive(opts *bind.TransactOpts, _isActive bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "setAgentActive", _isActive)
}

// SetAgentActive is a paid mutator transaction binding the contract method 0xcf3ea0e6.
//
// Solidity: function setAgentActive(bool _isActive) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SetAgentActive(_isActive bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetAgentActive(&_AgentBusinessCardV2.TransactOpts, _isActive)
}

// SetAgentActive is a paid mutator transaction binding the contract method 0xcf3ea0e6.
//
// Solidity: function setAgentActive(bool _isActive) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) SetAgentActive(_isActive bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetAgentActive(&_AgentBusinessCardV2.TransactOpts, _isActive)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetApprovalForAll(&_AgentBusinessCardV2.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetApprovalForAll(&_AgentBusinessCardV2.TransactOpts, operator, approved)
}

// SetSoulboundEnabled is a paid mutator transaction binding the contract method 0x75e6fa58.
//
// Solidity: function setSoulboundEnabled(bool _enabled) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) SetSoulboundEnabled(opts *bind.TransactOpts, _enabled bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "setSoulboundEnabled", _enabled)
}

// SetSoulboundEnabled is a paid mutator transaction binding the contract method 0x75e6fa58.
//
// Solidity: function setSoulboundEnabled(bool _enabled) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) SetSoulboundEnabled(_enabled bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetSoulboundEnabled(&_AgentBusinessCardV2.TransactOpts, _enabled)
}

// SetSoulboundEnabled is a paid mutator transaction binding the contract method 0x75e6fa58.
//
// Solidity: function setSoulboundEnabled(bool _enabled) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) SetSoulboundEnabled(_enabled bool) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.SetSoulboundEnabled(&_AgentBusinessCardV2.TransactOpts, _enabled)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.TransferFrom(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.TransferFrom(&_AgentBusinessCardV2.TransactOpts, from, to, tokenId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.TransferOwnership(&_AgentBusinessCardV2.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.TransferOwnership(&_AgentBusinessCardV2.TransactOpts, newOwner)
}

// UpdateAgentMetadata is a paid mutator transaction binding the contract method 0x46dcd1d9.
//
// Solidity: function updateAgentMetadata(string _description, string _contactInfo, string _pricingModel, string _version) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Transactor) UpdateAgentMetadata(opts *bind.TransactOpts, _description string, _contactInfo string, _pricingModel string, _version string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.contract.Transact(opts, "updateAgentMetadata", _description, _contactInfo, _pricingModel, _version)
}

// UpdateAgentMetadata is a paid mutator transaction binding the contract method 0x46dcd1d9.
//
// Solidity: function updateAgentMetadata(string _description, string _contactInfo, string _pricingModel, string _version) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2Session) UpdateAgentMetadata(_description string, _contactInfo string, _pricingModel string, _version string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.UpdateAgentMetadata(&_AgentBusinessCardV2.TransactOpts, _description, _contactInfo, _pricingModel, _version)
}

// UpdateAgentMetadata is a paid mutator transaction binding the contract method 0x46dcd1d9.
//
// Solidity: function updateAgentMetadata(string _description, string _contactInfo, string _pricingModel, string _version) returns()
func (_AgentBusinessCardV2 *AgentBusinessCardV2TransactorSession) UpdateAgentMetadata(_description string, _contactInfo string, _pricingModel string, _version string) (*types.Transaction, error) {
	return _AgentBusinessCardV2.Contract.UpdateAgentMetadata(&_AgentBusinessCardV2.TransactOpts, _description, _contactInfo, _pricingModel, _version)
}

// AgentBusinessCardV2AgentActivatedIterator is returned from FilterAgentActivated and is used to iterate over the raw logs and unpacked data for AgentActivated events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentActivatedIterator struct {
	Event *AgentBusinessCardV2AgentActivated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2AgentActivatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2AgentActivated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2AgentActivated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2AgentActivatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2AgentActivatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2AgentActivated represents a AgentActivated event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentActivated struct {
	TokenId  *big.Int
	IsActive bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAgentActivated is a free log retrieval operation binding the contract event 0x9d7dc34d28a140e29572ec36e2be850a194f12dfbf5bbbbb3e6a92ddff13ab09.
//
// Solidity: event AgentActivated(uint256 indexed tokenId, bool isActive)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterAgentActivated(opts *bind.FilterOpts, tokenId []*big.Int) (*AgentBusinessCardV2AgentActivatedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "AgentActivated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2AgentActivatedIterator{contract: _AgentBusinessCardV2.contract, event: "AgentActivated", logs: logs, sub: sub}, nil
}

// WatchAgentActivated is a free log subscription operation binding the contract event 0x9d7dc34d28a140e29572ec36e2be850a194f12dfbf5bbbbb3e6a92ddff13ab09.
//
// Solidity: event AgentActivated(uint256 indexed tokenId, bool isActive)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchAgentActivated(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2AgentActivated, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "AgentActivated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2AgentActivated)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentActivated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAgentActivated is a log parse operation binding the contract event 0x9d7dc34d28a140e29572ec36e2be850a194f12dfbf5bbbbb3e6a92ddff13ab09.
//
// Solidity: event AgentActivated(uint256 indexed tokenId, bool isActive)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseAgentActivated(log types.Log) (*AgentBusinessCardV2AgentActivated, error) {
	event := new(AgentBusinessCardV2AgentActivated)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentActivated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2AgentRegisteredIterator is returned from FilterAgentRegistered and is used to iterate over the raw logs and unpacked data for AgentRegistered events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentRegisteredIterator struct {
	Event *AgentBusinessCardV2AgentRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2AgentRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2AgentRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2AgentRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2AgentRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2AgentRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2AgentRegistered represents a AgentRegistered event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentRegistered struct {
	TokenId *big.Int
	Owner   common.Address
	Name    string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAgentRegistered is a free log retrieval operation binding the contract event 0x0d063c6022bff16d09991a9f91882ffa112f5fb2529136f65eb4c77bbd047e43.
//
// Solidity: event AgentRegistered(uint256 indexed tokenId, address indexed owner, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterAgentRegistered(opts *bind.FilterOpts, tokenId []*big.Int, owner []common.Address) (*AgentBusinessCardV2AgentRegisteredIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "AgentRegistered", tokenIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2AgentRegisteredIterator{contract: _AgentBusinessCardV2.contract, event: "AgentRegistered", logs: logs, sub: sub}, nil
}

// WatchAgentRegistered is a free log subscription operation binding the contract event 0x0d063c6022bff16d09991a9f91882ffa112f5fb2529136f65eb4c77bbd047e43.
//
// Solidity: event AgentRegistered(uint256 indexed tokenId, address indexed owner, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchAgentRegistered(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2AgentRegistered, tokenId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "AgentRegistered", tokenIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2AgentRegistered)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAgentRegistered is a log parse operation binding the contract event 0x0d063c6022bff16d09991a9f91882ffa112f5fb2529136f65eb4c77bbd047e43.
//
// Solidity: event AgentRegistered(uint256 indexed tokenId, address indexed owner, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseAgentRegistered(log types.Log) (*AgentBusinessCardV2AgentRegistered, error) {
	event := new(AgentBusinessCardV2AgentRegistered)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2AgentUpdatedIterator is returned from FilterAgentUpdated and is used to iterate over the raw logs and unpacked data for AgentUpdated events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentUpdatedIterator struct {
	Event *AgentBusinessCardV2AgentUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2AgentUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2AgentUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2AgentUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2AgentUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2AgentUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2AgentUpdated represents a AgentUpdated event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2AgentUpdated struct {
	TokenId *big.Int
	Name    string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAgentUpdated is a free log retrieval operation binding the contract event 0xf0677ec077e6eec8e31d4e8be0cb20b797f3a9dc243e2ddc641532c0ff412036.
//
// Solidity: event AgentUpdated(uint256 indexed tokenId, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterAgentUpdated(opts *bind.FilterOpts, tokenId []*big.Int) (*AgentBusinessCardV2AgentUpdatedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "AgentUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2AgentUpdatedIterator{contract: _AgentBusinessCardV2.contract, event: "AgentUpdated", logs: logs, sub: sub}, nil
}

// WatchAgentUpdated is a free log subscription operation binding the contract event 0xf0677ec077e6eec8e31d4e8be0cb20b797f3a9dc243e2ddc641532c0ff412036.
//
// Solidity: event AgentUpdated(uint256 indexed tokenId, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchAgentUpdated(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2AgentUpdated, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "AgentUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2AgentUpdated)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAgentUpdated is a log parse operation binding the contract event 0xf0677ec077e6eec8e31d4e8be0cb20b797f3a9dc243e2ddc641532c0ff412036.
//
// Solidity: event AgentUpdated(uint256 indexed tokenId, string name)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseAgentUpdated(log types.Log) (*AgentBusinessCardV2AgentUpdated, error) {
	event := new(AgentBusinessCardV2AgentUpdated)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "AgentUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2ApprovalIterator struct {
	Event *AgentBusinessCardV2Approval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2Approval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2Approval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2Approval represents a Approval event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2Approval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*AgentBusinessCardV2ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2ApprovalIterator{contract: _AgentBusinessCardV2.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2Approval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2Approval)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseApproval(log types.Log) (*AgentBusinessCardV2Approval, error) {
	event := new(AgentBusinessCardV2Approval)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2ApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2ApprovalForAllIterator struct {
	Event *AgentBusinessCardV2ApprovalForAll // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2ApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2ApprovalForAll)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2ApprovalForAll)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2ApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2ApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2ApprovalForAll represents a ApprovalForAll event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2ApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*AgentBusinessCardV2ApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2ApprovalForAllIterator{contract: _AgentBusinessCardV2.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2ApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2ApprovalForAll)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseApprovalForAll(log types.Log) (*AgentBusinessCardV2ApprovalForAll, error) {
	event := new(AgentBusinessCardV2ApprovalForAll)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2OwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2OwnershipTransferredIterator struct {
	Event *AgentBusinessCardV2OwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2OwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2OwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2OwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2OwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2OwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2OwnershipTransferred represents a OwnershipTransferred event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2OwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AgentBusinessCardV2OwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2OwnershipTransferredIterator{contract: _AgentBusinessCardV2.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2OwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2OwnershipTransferred)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseOwnershipTransferred(log types.Log) (*AgentBusinessCardV2OwnershipTransferred, error) {
	event := new(AgentBusinessCardV2OwnershipTransferred)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentBusinessCardV2TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2TransferIterator struct {
	Event *AgentBusinessCardV2Transfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentBusinessCardV2TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentBusinessCardV2Transfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentBusinessCardV2Transfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentBusinessCardV2TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentBusinessCardV2TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentBusinessCardV2Transfer represents a Transfer event raised by the AgentBusinessCardV2 contract.
type AgentBusinessCardV2Transfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*AgentBusinessCardV2TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentBusinessCardV2TransferIterator{contract: _AgentBusinessCardV2.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *AgentBusinessCardV2Transfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AgentBusinessCardV2.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentBusinessCardV2Transfer)
				if err := _AgentBusinessCardV2.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AgentBusinessCardV2 *AgentBusinessCardV2Filterer) ParseTransfer(log types.Log) (*AgentBusinessCardV2Transfer, error) {
	event := new(AgentBusinessCardV2Transfer)
	if err := _AgentBusinessCardV2.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
