// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newGovernance\",\"type\":\"address\"}],\"name\":\"AcceptGovernance\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newOracleMember\",\"type\":\"address\"}],\"name\":\"AddOracleMember\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"BanValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"withdrawalAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"claimableBalance\",\"type\":\"uint256\"}],\"name\":\"ClaimRewards\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"donationAmount\",\"type\":\"uint256\"}],\"name\":\"EtherReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"initialSmoothingPoolSlot\",\"type\":\"uint64\"}],\"name\":\"InitSmoothingPool\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oracleMemberRemoved\",\"type\":\"address\"}],\"name\":\"RemoveOracleMember\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"newRewardsRoot\",\"type\":\"bytes32\"}],\"name\":\"ReportConsolidated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"withdrawalAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"poolRecipient\",\"type\":\"address\"}],\"name\":\"SetRewardRecipient\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"newRewardsRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oracleMember\",\"type\":\"address\"}],\"name\":\"SubmitReport\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"subscriptionCollateral\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"SubscribeValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newPendingGovernance\",\"type\":\"address\"}],\"name\":\"TransferGovernance\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"UnbanValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"UnsubscribeValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"newCheckpointSlotSize\",\"type\":\"uint64\"}],\"name\":\"UpdateCheckpointSlotSize\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPoolFee\",\"type\":\"uint256\"}],\"name\":\"UpdatePoolFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newPoolFeeRecipient\",\"type\":\"address\"}],\"name\":\"UpdatePoolFeeRecipient\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"newQuorum\",\"type\":\"uint64\"}],\"name\":\"UpdateQuorum\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newSubscriptionCollateral\",\"type\":\"uint256\"}],\"name\":\"UpdateSubscriptionCollateral\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"INITIAL_REPORT_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOracleMember\",\"type\":\"address\"}],\"name\":\"addOracleMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"addressToVotedReportHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"validatorIDArray\",\"type\":\"uint64[]\"}],\"name\":\"banValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"checkpointSlotSize\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"withdrawalAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"accumulatedBalance\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"merkleProof\",\"type\":\"bytes32[]\"}],\"name\":\"claimRewards\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"claimedBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deploymentBlockNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllOracleMembers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracleMember\",\"type\":\"address\"}],\"name\":\"getOracleMemberIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOracleMembersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_slot\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"_rewardsRoot\",\"type\":\"bytes32\"}],\"name\":\"getReportHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"initialSmoothingPoolSlot\",\"type\":\"uint64\"}],\"name\":\"initSmoothingPool\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_governance\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_subscriptionCollateral\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_poolFee\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_poolFeeRecipient\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"_checkpointSlotSize\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"_quorum\",\"type\":\"uint64\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastConsolidatedSlot\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"oracleMembers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poolFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poolFeeRecipient\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorum\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracleMemberAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"oracleMemberIndex\",\"type\":\"uint256\"}],\"name\":\"removeOracleMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"reportHashToReport\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"votes\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"rewardRecipient\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardsRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"}],\"name\":\"setRewardRecipient\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"slotNumber\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"proposedRewardsRoot\",\"type\":\"bytes32\"}],\"name\":\"submitReport\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"subscribeValidator\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"validatorIDArray\",\"type\":\"uint64[]\"}],\"name\":\"subscribeValidators\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subscriptionCollateral\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newPendingGovernance\",\"type\":\"address\"}],\"name\":\"transferGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"validatorIDArray\",\"type\":\"uint64[]\"}],\"name\":\"unbanValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"validatorID\",\"type\":\"uint64\"}],\"name\":\"unsubscribeValidator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"validatorIDArray\",\"type\":\"uint64[]\"}],\"name\":\"unsubscribeValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newCheckpointSlotSize\",\"type\":\"uint64\"}],\"name\":\"updateCheckpointSlotSize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newSubscriptionCollateral\",\"type\":\"uint256\"}],\"name\":\"updateCollateral\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPoolFee\",\"type\":\"uint256\"}],\"name\":\"updatePoolFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newPoolFeeRecipient\",\"type\":\"address\"}],\"name\":\"updatePoolFeeRecipient\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newQuorum\",\"type\":\"uint64\"}],\"name\":\"updateQuorum\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// INITIALREPORTHASH is a free data retrieval call binding the contract method 0x7b5c2f88.
//
// Solidity: function INITIAL_REPORT_HASH() view returns(bytes32)
func (_Contract *ContractCaller) INITIALREPORTHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "INITIAL_REPORT_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// INITIALREPORTHASH is a free data retrieval call binding the contract method 0x7b5c2f88.
//
// Solidity: function INITIAL_REPORT_HASH() view returns(bytes32)
func (_Contract *ContractSession) INITIALREPORTHASH() ([32]byte, error) {
	return _Contract.Contract.INITIALREPORTHASH(&_Contract.CallOpts)
}

// INITIALREPORTHASH is a free data retrieval call binding the contract method 0x7b5c2f88.
//
// Solidity: function INITIAL_REPORT_HASH() view returns(bytes32)
func (_Contract *ContractCallerSession) INITIALREPORTHASH() ([32]byte, error) {
	return _Contract.Contract.INITIALREPORTHASH(&_Contract.CallOpts)
}

// AddressToVotedReportHash is a free data retrieval call binding the contract method 0xc1269650.
//
// Solidity: function addressToVotedReportHash(address ) view returns(bytes32)
func (_Contract *ContractCaller) AddressToVotedReportHash(opts *bind.CallOpts, arg0 common.Address) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "addressToVotedReportHash", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// AddressToVotedReportHash is a free data retrieval call binding the contract method 0xc1269650.
//
// Solidity: function addressToVotedReportHash(address ) view returns(bytes32)
func (_Contract *ContractSession) AddressToVotedReportHash(arg0 common.Address) ([32]byte, error) {
	return _Contract.Contract.AddressToVotedReportHash(&_Contract.CallOpts, arg0)
}

// AddressToVotedReportHash is a free data retrieval call binding the contract method 0xc1269650.
//
// Solidity: function addressToVotedReportHash(address ) view returns(bytes32)
func (_Contract *ContractCallerSession) AddressToVotedReportHash(arg0 common.Address) ([32]byte, error) {
	return _Contract.Contract.AddressToVotedReportHash(&_Contract.CallOpts, arg0)
}

// CheckpointSlotSize is a free data retrieval call binding the contract method 0xc7f75d3f.
//
// Solidity: function checkpointSlotSize() view returns(uint64)
func (_Contract *ContractCaller) CheckpointSlotSize(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "checkpointSlotSize")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// CheckpointSlotSize is a free data retrieval call binding the contract method 0xc7f75d3f.
//
// Solidity: function checkpointSlotSize() view returns(uint64)
func (_Contract *ContractSession) CheckpointSlotSize() (uint64, error) {
	return _Contract.Contract.CheckpointSlotSize(&_Contract.CallOpts)
}

// CheckpointSlotSize is a free data retrieval call binding the contract method 0xc7f75d3f.
//
// Solidity: function checkpointSlotSize() view returns(uint64)
func (_Contract *ContractCallerSession) CheckpointSlotSize() (uint64, error) {
	return _Contract.Contract.CheckpointSlotSize(&_Contract.CallOpts)
}

// ClaimedBalance is a free data retrieval call binding the contract method 0x9886c2a5.
//
// Solidity: function claimedBalance(address ) view returns(uint256)
func (_Contract *ContractCaller) ClaimedBalance(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "claimedBalance", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ClaimedBalance is a free data retrieval call binding the contract method 0x9886c2a5.
//
// Solidity: function claimedBalance(address ) view returns(uint256)
func (_Contract *ContractSession) ClaimedBalance(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.ClaimedBalance(&_Contract.CallOpts, arg0)
}

// ClaimedBalance is a free data retrieval call binding the contract method 0x9886c2a5.
//
// Solidity: function claimedBalance(address ) view returns(uint256)
func (_Contract *ContractCallerSession) ClaimedBalance(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.ClaimedBalance(&_Contract.CallOpts, arg0)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xcf004217.
//
// Solidity: function deploymentBlockNumber() view returns(uint256)
func (_Contract *ContractCaller) DeploymentBlockNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "deploymentBlockNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xcf004217.
//
// Solidity: function deploymentBlockNumber() view returns(uint256)
func (_Contract *ContractSession) DeploymentBlockNumber() (*big.Int, error) {
	return _Contract.Contract.DeploymentBlockNumber(&_Contract.CallOpts)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xcf004217.
//
// Solidity: function deploymentBlockNumber() view returns(uint256)
func (_Contract *ContractCallerSession) DeploymentBlockNumber() (*big.Int, error) {
	return _Contract.Contract.DeploymentBlockNumber(&_Contract.CallOpts)
}

// GetAllOracleMembers is a free data retrieval call binding the contract method 0x2827acf3.
//
// Solidity: function getAllOracleMembers() view returns(address[])
func (_Contract *ContractCaller) GetAllOracleMembers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getAllOracleMembers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetAllOracleMembers is a free data retrieval call binding the contract method 0x2827acf3.
//
// Solidity: function getAllOracleMembers() view returns(address[])
func (_Contract *ContractSession) GetAllOracleMembers() ([]common.Address, error) {
	return _Contract.Contract.GetAllOracleMembers(&_Contract.CallOpts)
}

// GetAllOracleMembers is a free data retrieval call binding the contract method 0x2827acf3.
//
// Solidity: function getAllOracleMembers() view returns(address[])
func (_Contract *ContractCallerSession) GetAllOracleMembers() ([]common.Address, error) {
	return _Contract.Contract.GetAllOracleMembers(&_Contract.CallOpts)
}

// GetOracleMemberIndex is a free data retrieval call binding the contract method 0xe1a80493.
//
// Solidity: function getOracleMemberIndex(address oracleMember) view returns(uint256)
func (_Contract *ContractCaller) GetOracleMemberIndex(opts *bind.CallOpts, oracleMember common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getOracleMemberIndex", oracleMember)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOracleMemberIndex is a free data retrieval call binding the contract method 0xe1a80493.
//
// Solidity: function getOracleMemberIndex(address oracleMember) view returns(uint256)
func (_Contract *ContractSession) GetOracleMemberIndex(oracleMember common.Address) (*big.Int, error) {
	return _Contract.Contract.GetOracleMemberIndex(&_Contract.CallOpts, oracleMember)
}

// GetOracleMemberIndex is a free data retrieval call binding the contract method 0xe1a80493.
//
// Solidity: function getOracleMemberIndex(address oracleMember) view returns(uint256)
func (_Contract *ContractCallerSession) GetOracleMemberIndex(oracleMember common.Address) (*big.Int, error) {
	return _Contract.Contract.GetOracleMemberIndex(&_Contract.CallOpts, oracleMember)
}

// GetOracleMembersCount is a free data retrieval call binding the contract method 0xc37f45bf.
//
// Solidity: function getOracleMembersCount() view returns(uint256)
func (_Contract *ContractCaller) GetOracleMembersCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getOracleMembersCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOracleMembersCount is a free data retrieval call binding the contract method 0xc37f45bf.
//
// Solidity: function getOracleMembersCount() view returns(uint256)
func (_Contract *ContractSession) GetOracleMembersCount() (*big.Int, error) {
	return _Contract.Contract.GetOracleMembersCount(&_Contract.CallOpts)
}

// GetOracleMembersCount is a free data retrieval call binding the contract method 0xc37f45bf.
//
// Solidity: function getOracleMembersCount() view returns(uint256)
func (_Contract *ContractCallerSession) GetOracleMembersCount() (*big.Int, error) {
	return _Contract.Contract.GetOracleMembersCount(&_Contract.CallOpts)
}

// GetReportHash is a free data retrieval call binding the contract method 0x781f5855.
//
// Solidity: function getReportHash(uint64 _slot, bytes32 _rewardsRoot) pure returns(bytes32)
func (_Contract *ContractCaller) GetReportHash(opts *bind.CallOpts, _slot uint64, _rewardsRoot [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getReportHash", _slot, _rewardsRoot)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetReportHash is a free data retrieval call binding the contract method 0x781f5855.
//
// Solidity: function getReportHash(uint64 _slot, bytes32 _rewardsRoot) pure returns(bytes32)
func (_Contract *ContractSession) GetReportHash(_slot uint64, _rewardsRoot [32]byte) ([32]byte, error) {
	return _Contract.Contract.GetReportHash(&_Contract.CallOpts, _slot, _rewardsRoot)
}

// GetReportHash is a free data retrieval call binding the contract method 0x781f5855.
//
// Solidity: function getReportHash(uint64 _slot, bytes32 _rewardsRoot) pure returns(bytes32)
func (_Contract *ContractCallerSession) GetReportHash(_slot uint64, _rewardsRoot [32]byte) ([32]byte, error) {
	return _Contract.Contract.GetReportHash(&_Contract.CallOpts, _slot, _rewardsRoot)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Contract *ContractCaller) Governance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "governance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Contract *ContractSession) Governance() (common.Address, error) {
	return _Contract.Contract.Governance(&_Contract.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Contract *ContractCallerSession) Governance() (common.Address, error) {
	return _Contract.Contract.Governance(&_Contract.CallOpts)
}

// LastConsolidatedSlot is a free data retrieval call binding the contract method 0xa119d687.
//
// Solidity: function lastConsolidatedSlot() view returns(uint64)
func (_Contract *ContractCaller) LastConsolidatedSlot(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "lastConsolidatedSlot")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// LastConsolidatedSlot is a free data retrieval call binding the contract method 0xa119d687.
//
// Solidity: function lastConsolidatedSlot() view returns(uint64)
func (_Contract *ContractSession) LastConsolidatedSlot() (uint64, error) {
	return _Contract.Contract.LastConsolidatedSlot(&_Contract.CallOpts)
}

// LastConsolidatedSlot is a free data retrieval call binding the contract method 0xa119d687.
//
// Solidity: function lastConsolidatedSlot() view returns(uint64)
func (_Contract *ContractCallerSession) LastConsolidatedSlot() (uint64, error) {
	return _Contract.Contract.LastConsolidatedSlot(&_Contract.CallOpts)
}

// OracleMembers is a free data retrieval call binding the contract method 0x37215ebd.
//
// Solidity: function oracleMembers(uint256 ) view returns(address)
func (_Contract *ContractCaller) OracleMembers(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "oracleMembers", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OracleMembers is a free data retrieval call binding the contract method 0x37215ebd.
//
// Solidity: function oracleMembers(uint256 ) view returns(address)
func (_Contract *ContractSession) OracleMembers(arg0 *big.Int) (common.Address, error) {
	return _Contract.Contract.OracleMembers(&_Contract.CallOpts, arg0)
}

// OracleMembers is a free data retrieval call binding the contract method 0x37215ebd.
//
// Solidity: function oracleMembers(uint256 ) view returns(address)
func (_Contract *ContractCallerSession) OracleMembers(arg0 *big.Int) (common.Address, error) {
	return _Contract.Contract.OracleMembers(&_Contract.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCallerSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Contract *ContractCaller) PendingGovernance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "pendingGovernance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Contract *ContractSession) PendingGovernance() (common.Address, error) {
	return _Contract.Contract.PendingGovernance(&_Contract.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Contract *ContractCallerSession) PendingGovernance() (common.Address, error) {
	return _Contract.Contract.PendingGovernance(&_Contract.CallOpts)
}

// PoolFee is a free data retrieval call binding the contract method 0x089fe6aa.
//
// Solidity: function poolFee() view returns(uint256)
func (_Contract *ContractCaller) PoolFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "poolFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PoolFee is a free data retrieval call binding the contract method 0x089fe6aa.
//
// Solidity: function poolFee() view returns(uint256)
func (_Contract *ContractSession) PoolFee() (*big.Int, error) {
	return _Contract.Contract.PoolFee(&_Contract.CallOpts)
}

// PoolFee is a free data retrieval call binding the contract method 0x089fe6aa.
//
// Solidity: function poolFee() view returns(uint256)
func (_Contract *ContractCallerSession) PoolFee() (*big.Int, error) {
	return _Contract.Contract.PoolFee(&_Contract.CallOpts)
}

// PoolFeeRecipient is a free data retrieval call binding the contract method 0x75f678a0.
//
// Solidity: function poolFeeRecipient() view returns(address)
func (_Contract *ContractCaller) PoolFeeRecipient(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "poolFeeRecipient")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PoolFeeRecipient is a free data retrieval call binding the contract method 0x75f678a0.
//
// Solidity: function poolFeeRecipient() view returns(address)
func (_Contract *ContractSession) PoolFeeRecipient() (common.Address, error) {
	return _Contract.Contract.PoolFeeRecipient(&_Contract.CallOpts)
}

// PoolFeeRecipient is a free data retrieval call binding the contract method 0x75f678a0.
//
// Solidity: function poolFeeRecipient() view returns(address)
func (_Contract *ContractCallerSession) PoolFeeRecipient() (common.Address, error) {
	return _Contract.Contract.PoolFeeRecipient(&_Contract.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint64)
func (_Contract *ContractCaller) Quorum(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "quorum")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint64)
func (_Contract *ContractSession) Quorum() (uint64, error) {
	return _Contract.Contract.Quorum(&_Contract.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint64)
func (_Contract *ContractCallerSession) Quorum() (uint64, error) {
	return _Contract.Contract.Quorum(&_Contract.CallOpts)
}

// ReportHashToReport is a free data retrieval call binding the contract method 0x03ef4aff.
//
// Solidity: function reportHashToReport(bytes32 ) view returns(uint64 slot, uint64 votes)
func (_Contract *ContractCaller) ReportHashToReport(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Slot  uint64
	Votes uint64
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "reportHashToReport", arg0)

	outstruct := new(struct {
		Slot  uint64
		Votes uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Slot = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.Votes = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// ReportHashToReport is a free data retrieval call binding the contract method 0x03ef4aff.
//
// Solidity: function reportHashToReport(bytes32 ) view returns(uint64 slot, uint64 votes)
func (_Contract *ContractSession) ReportHashToReport(arg0 [32]byte) (struct {
	Slot  uint64
	Votes uint64
}, error) {
	return _Contract.Contract.ReportHashToReport(&_Contract.CallOpts, arg0)
}

// ReportHashToReport is a free data retrieval call binding the contract method 0x03ef4aff.
//
// Solidity: function reportHashToReport(bytes32 ) view returns(uint64 slot, uint64 votes)
func (_Contract *ContractCallerSession) ReportHashToReport(arg0 [32]byte) (struct {
	Slot  uint64
	Votes uint64
}, error) {
	return _Contract.Contract.ReportHashToReport(&_Contract.CallOpts, arg0)
}

// RewardRecipient is a free data retrieval call binding the contract method 0xf372c0c9.
//
// Solidity: function rewardRecipient(address ) view returns(address)
func (_Contract *ContractCaller) RewardRecipient(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "rewardRecipient", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardRecipient is a free data retrieval call binding the contract method 0xf372c0c9.
//
// Solidity: function rewardRecipient(address ) view returns(address)
func (_Contract *ContractSession) RewardRecipient(arg0 common.Address) (common.Address, error) {
	return _Contract.Contract.RewardRecipient(&_Contract.CallOpts, arg0)
}

// RewardRecipient is a free data retrieval call binding the contract method 0xf372c0c9.
//
// Solidity: function rewardRecipient(address ) view returns(address)
func (_Contract *ContractCallerSession) RewardRecipient(arg0 common.Address) (common.Address, error) {
	return _Contract.Contract.RewardRecipient(&_Contract.CallOpts, arg0)
}

// RewardsRoot is a free data retrieval call binding the contract method 0x217863b7.
//
// Solidity: function rewardsRoot() view returns(bytes32)
func (_Contract *ContractCaller) RewardsRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "rewardsRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RewardsRoot is a free data retrieval call binding the contract method 0x217863b7.
//
// Solidity: function rewardsRoot() view returns(bytes32)
func (_Contract *ContractSession) RewardsRoot() ([32]byte, error) {
	return _Contract.Contract.RewardsRoot(&_Contract.CallOpts)
}

// RewardsRoot is a free data retrieval call binding the contract method 0x217863b7.
//
// Solidity: function rewardsRoot() view returns(bytes32)
func (_Contract *ContractCallerSession) RewardsRoot() ([32]byte, error) {
	return _Contract.Contract.RewardsRoot(&_Contract.CallOpts)
}

// SubscriptionCollateral is a free data retrieval call binding the contract method 0xf93558e3.
//
// Solidity: function subscriptionCollateral() view returns(uint256)
func (_Contract *ContractCaller) SubscriptionCollateral(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "subscriptionCollateral")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubscriptionCollateral is a free data retrieval call binding the contract method 0xf93558e3.
//
// Solidity: function subscriptionCollateral() view returns(uint256)
func (_Contract *ContractSession) SubscriptionCollateral() (*big.Int, error) {
	return _Contract.Contract.SubscriptionCollateral(&_Contract.CallOpts)
}

// SubscriptionCollateral is a free data retrieval call binding the contract method 0xf93558e3.
//
// Solidity: function subscriptionCollateral() view returns(uint256)
func (_Contract *ContractCallerSession) SubscriptionCollateral() (*big.Int, error) {
	return _Contract.Contract.SubscriptionCollateral(&_Contract.CallOpts)
}

// AcceptGovernance is a paid mutator transaction binding the contract method 0x238efcbc.
//
// Solidity: function acceptGovernance() returns()
func (_Contract *ContractTransactor) AcceptGovernance(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "acceptGovernance")
}

// AcceptGovernance is a paid mutator transaction binding the contract method 0x238efcbc.
//
// Solidity: function acceptGovernance() returns()
func (_Contract *ContractSession) AcceptGovernance() (*types.Transaction, error) {
	return _Contract.Contract.AcceptGovernance(&_Contract.TransactOpts)
}

// AcceptGovernance is a paid mutator transaction binding the contract method 0x238efcbc.
//
// Solidity: function acceptGovernance() returns()
func (_Contract *ContractTransactorSession) AcceptGovernance() (*types.Transaction, error) {
	return _Contract.Contract.AcceptGovernance(&_Contract.TransactOpts)
}

// AddOracleMember is a paid mutator transaction binding the contract method 0xb164e437.
//
// Solidity: function addOracleMember(address newOracleMember) returns()
func (_Contract *ContractTransactor) AddOracleMember(opts *bind.TransactOpts, newOracleMember common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "addOracleMember", newOracleMember)
}

// AddOracleMember is a paid mutator transaction binding the contract method 0xb164e437.
//
// Solidity: function addOracleMember(address newOracleMember) returns()
func (_Contract *ContractSession) AddOracleMember(newOracleMember common.Address) (*types.Transaction, error) {
	return _Contract.Contract.AddOracleMember(&_Contract.TransactOpts, newOracleMember)
}

// AddOracleMember is a paid mutator transaction binding the contract method 0xb164e437.
//
// Solidity: function addOracleMember(address newOracleMember) returns()
func (_Contract *ContractTransactorSession) AddOracleMember(newOracleMember common.Address) (*types.Transaction, error) {
	return _Contract.Contract.AddOracleMember(&_Contract.TransactOpts, newOracleMember)
}

// BanValidators is a paid mutator transaction binding the contract method 0x17a6863f.
//
// Solidity: function banValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactor) BanValidators(opts *bind.TransactOpts, validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "banValidators", validatorIDArray)
}

// BanValidators is a paid mutator transaction binding the contract method 0x17a6863f.
//
// Solidity: function banValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractSession) BanValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.BanValidators(&_Contract.TransactOpts, validatorIDArray)
}

// BanValidators is a paid mutator transaction binding the contract method 0x17a6863f.
//
// Solidity: function banValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactorSession) BanValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.BanValidators(&_Contract.TransactOpts, validatorIDArray)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0xd64bc331.
//
// Solidity: function claimRewards(address withdrawalAddress, uint256 accumulatedBalance, bytes32[] merkleProof) returns()
func (_Contract *ContractTransactor) ClaimRewards(opts *bind.TransactOpts, withdrawalAddress common.Address, accumulatedBalance *big.Int, merkleProof [][32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimRewards", withdrawalAddress, accumulatedBalance, merkleProof)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0xd64bc331.
//
// Solidity: function claimRewards(address withdrawalAddress, uint256 accumulatedBalance, bytes32[] merkleProof) returns()
func (_Contract *ContractSession) ClaimRewards(withdrawalAddress common.Address, accumulatedBalance *big.Int, merkleProof [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.ClaimRewards(&_Contract.TransactOpts, withdrawalAddress, accumulatedBalance, merkleProof)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0xd64bc331.
//
// Solidity: function claimRewards(address withdrawalAddress, uint256 accumulatedBalance, bytes32[] merkleProof) returns()
func (_Contract *ContractTransactorSession) ClaimRewards(withdrawalAddress common.Address, accumulatedBalance *big.Int, merkleProof [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.ClaimRewards(&_Contract.TransactOpts, withdrawalAddress, accumulatedBalance, merkleProof)
}

// InitSmoothingPool is a paid mutator transaction binding the contract method 0x3964bdca.
//
// Solidity: function initSmoothingPool(uint64 initialSmoothingPoolSlot) returns()
func (_Contract *ContractTransactor) InitSmoothingPool(opts *bind.TransactOpts, initialSmoothingPoolSlot uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initSmoothingPool", initialSmoothingPoolSlot)
}

// InitSmoothingPool is a paid mutator transaction binding the contract method 0x3964bdca.
//
// Solidity: function initSmoothingPool(uint64 initialSmoothingPoolSlot) returns()
func (_Contract *ContractSession) InitSmoothingPool(initialSmoothingPoolSlot uint64) (*types.Transaction, error) {
	return _Contract.Contract.InitSmoothingPool(&_Contract.TransactOpts, initialSmoothingPoolSlot)
}

// InitSmoothingPool is a paid mutator transaction binding the contract method 0x3964bdca.
//
// Solidity: function initSmoothingPool(uint64 initialSmoothingPoolSlot) returns()
func (_Contract *ContractTransactorSession) InitSmoothingPool(initialSmoothingPoolSlot uint64) (*types.Transaction, error) {
	return _Contract.Contract.InitSmoothingPool(&_Contract.TransactOpts, initialSmoothingPoolSlot)
}

// Initialize is a paid mutator transaction binding the contract method 0x151e4d3d.
//
// Solidity: function initialize(address _governance, uint256 _subscriptionCollateral, uint256 _poolFee, address _poolFeeRecipient, uint64 _checkpointSlotSize, uint64 _quorum) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _governance common.Address, _subscriptionCollateral *big.Int, _poolFee *big.Int, _poolFeeRecipient common.Address, _checkpointSlotSize uint64, _quorum uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _governance, _subscriptionCollateral, _poolFee, _poolFeeRecipient, _checkpointSlotSize, _quorum)
}

// Initialize is a paid mutator transaction binding the contract method 0x151e4d3d.
//
// Solidity: function initialize(address _governance, uint256 _subscriptionCollateral, uint256 _poolFee, address _poolFeeRecipient, uint64 _checkpointSlotSize, uint64 _quorum) returns()
func (_Contract *ContractSession) Initialize(_governance common.Address, _subscriptionCollateral *big.Int, _poolFee *big.Int, _poolFeeRecipient common.Address, _checkpointSlotSize uint64, _quorum uint64) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _governance, _subscriptionCollateral, _poolFee, _poolFeeRecipient, _checkpointSlotSize, _quorum)
}

// Initialize is a paid mutator transaction binding the contract method 0x151e4d3d.
//
// Solidity: function initialize(address _governance, uint256 _subscriptionCollateral, uint256 _poolFee, address _poolFeeRecipient, uint64 _checkpointSlotSize, uint64 _quorum) returns()
func (_Contract *ContractTransactorSession) Initialize(_governance common.Address, _subscriptionCollateral *big.Int, _poolFee *big.Int, _poolFeeRecipient common.Address, _checkpointSlotSize uint64, _quorum uint64) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _governance, _subscriptionCollateral, _poolFee, _poolFeeRecipient, _checkpointSlotSize, _quorum)
}

// RemoveOracleMember is a paid mutator transaction binding the contract method 0x53985e5a.
//
// Solidity: function removeOracleMember(address oracleMemberAddress, uint256 oracleMemberIndex) returns()
func (_Contract *ContractTransactor) RemoveOracleMember(opts *bind.TransactOpts, oracleMemberAddress common.Address, oracleMemberIndex *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "removeOracleMember", oracleMemberAddress, oracleMemberIndex)
}

// RemoveOracleMember is a paid mutator transaction binding the contract method 0x53985e5a.
//
// Solidity: function removeOracleMember(address oracleMemberAddress, uint256 oracleMemberIndex) returns()
func (_Contract *ContractSession) RemoveOracleMember(oracleMemberAddress common.Address, oracleMemberIndex *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RemoveOracleMember(&_Contract.TransactOpts, oracleMemberAddress, oracleMemberIndex)
}

// RemoveOracleMember is a paid mutator transaction binding the contract method 0x53985e5a.
//
// Solidity: function removeOracleMember(address oracleMemberAddress, uint256 oracleMemberIndex) returns()
func (_Contract *ContractTransactorSession) RemoveOracleMember(oracleMemberAddress common.Address, oracleMemberIndex *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RemoveOracleMember(&_Contract.TransactOpts, oracleMemberAddress, oracleMemberIndex)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// SetRewardRecipient is a paid mutator transaction binding the contract method 0xe521136f.
//
// Solidity: function setRewardRecipient(address rewardAddress) returns()
func (_Contract *ContractTransactor) SetRewardRecipient(opts *bind.TransactOpts, rewardAddress common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setRewardRecipient", rewardAddress)
}

// SetRewardRecipient is a paid mutator transaction binding the contract method 0xe521136f.
//
// Solidity: function setRewardRecipient(address rewardAddress) returns()
func (_Contract *ContractSession) SetRewardRecipient(rewardAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetRewardRecipient(&_Contract.TransactOpts, rewardAddress)
}

// SetRewardRecipient is a paid mutator transaction binding the contract method 0xe521136f.
//
// Solidity: function setRewardRecipient(address rewardAddress) returns()
func (_Contract *ContractTransactorSession) SetRewardRecipient(rewardAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetRewardRecipient(&_Contract.TransactOpts, rewardAddress)
}

// SubmitReport is a paid mutator transaction binding the contract method 0xb539f38b.
//
// Solidity: function submitReport(uint64 slotNumber, bytes32 proposedRewardsRoot) returns()
func (_Contract *ContractTransactor) SubmitReport(opts *bind.TransactOpts, slotNumber uint64, proposedRewardsRoot [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "submitReport", slotNumber, proposedRewardsRoot)
}

// SubmitReport is a paid mutator transaction binding the contract method 0xb539f38b.
//
// Solidity: function submitReport(uint64 slotNumber, bytes32 proposedRewardsRoot) returns()
func (_Contract *ContractSession) SubmitReport(slotNumber uint64, proposedRewardsRoot [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitReport(&_Contract.TransactOpts, slotNumber, proposedRewardsRoot)
}

// SubmitReport is a paid mutator transaction binding the contract method 0xb539f38b.
//
// Solidity: function submitReport(uint64 slotNumber, bytes32 proposedRewardsRoot) returns()
func (_Contract *ContractTransactorSession) SubmitReport(slotNumber uint64, proposedRewardsRoot [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitReport(&_Contract.TransactOpts, slotNumber, proposedRewardsRoot)
}

// SubscribeValidator is a paid mutator transaction binding the contract method 0xb9cd552e.
//
// Solidity: function subscribeValidator(uint64 validatorID) payable returns()
func (_Contract *ContractTransactor) SubscribeValidator(opts *bind.TransactOpts, validatorID uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "subscribeValidator", validatorID)
}

// SubscribeValidator is a paid mutator transaction binding the contract method 0xb9cd552e.
//
// Solidity: function subscribeValidator(uint64 validatorID) payable returns()
func (_Contract *ContractSession) SubscribeValidator(validatorID uint64) (*types.Transaction, error) {
	return _Contract.Contract.SubscribeValidator(&_Contract.TransactOpts, validatorID)
}

// SubscribeValidator is a paid mutator transaction binding the contract method 0xb9cd552e.
//
// Solidity: function subscribeValidator(uint64 validatorID) payable returns()
func (_Contract *ContractTransactorSession) SubscribeValidator(validatorID uint64) (*types.Transaction, error) {
	return _Contract.Contract.SubscribeValidator(&_Contract.TransactOpts, validatorID)
}

// SubscribeValidators is a paid mutator transaction binding the contract method 0x5bc37534.
//
// Solidity: function subscribeValidators(uint64[] validatorIDArray) payable returns()
func (_Contract *ContractTransactor) SubscribeValidators(opts *bind.TransactOpts, validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "subscribeValidators", validatorIDArray)
}

// SubscribeValidators is a paid mutator transaction binding the contract method 0x5bc37534.
//
// Solidity: function subscribeValidators(uint64[] validatorIDArray) payable returns()
func (_Contract *ContractSession) SubscribeValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.SubscribeValidators(&_Contract.TransactOpts, validatorIDArray)
}

// SubscribeValidators is a paid mutator transaction binding the contract method 0x5bc37534.
//
// Solidity: function subscribeValidators(uint64[] validatorIDArray) payable returns()
func (_Contract *ContractTransactorSession) SubscribeValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.SubscribeValidators(&_Contract.TransactOpts, validatorIDArray)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newPendingGovernance) returns()
func (_Contract *ContractTransactor) TransferGovernance(opts *bind.TransactOpts, newPendingGovernance common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferGovernance", newPendingGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newPendingGovernance) returns()
func (_Contract *ContractSession) TransferGovernance(newPendingGovernance common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferGovernance(&_Contract.TransactOpts, newPendingGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newPendingGovernance) returns()
func (_Contract *ContractTransactorSession) TransferGovernance(newPendingGovernance common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferGovernance(&_Contract.TransactOpts, newPendingGovernance)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// UnbanValidators is a paid mutator transaction binding the contract method 0x0008900a.
//
// Solidity: function unbanValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactor) UnbanValidators(opts *bind.TransactOpts, validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unbanValidators", validatorIDArray)
}

// UnbanValidators is a paid mutator transaction binding the contract method 0x0008900a.
//
// Solidity: function unbanValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractSession) UnbanValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnbanValidators(&_Contract.TransactOpts, validatorIDArray)
}

// UnbanValidators is a paid mutator transaction binding the contract method 0x0008900a.
//
// Solidity: function unbanValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactorSession) UnbanValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnbanValidators(&_Contract.TransactOpts, validatorIDArray)
}

// UnsubscribeValidator is a paid mutator transaction binding the contract method 0xc1542c52.
//
// Solidity: function unsubscribeValidator(uint64 validatorID) returns()
func (_Contract *ContractTransactor) UnsubscribeValidator(opts *bind.TransactOpts, validatorID uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unsubscribeValidator", validatorID)
}

// UnsubscribeValidator is a paid mutator transaction binding the contract method 0xc1542c52.
//
// Solidity: function unsubscribeValidator(uint64 validatorID) returns()
func (_Contract *ContractSession) UnsubscribeValidator(validatorID uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnsubscribeValidator(&_Contract.TransactOpts, validatorID)
}

// UnsubscribeValidator is a paid mutator transaction binding the contract method 0xc1542c52.
//
// Solidity: function unsubscribeValidator(uint64 validatorID) returns()
func (_Contract *ContractTransactorSession) UnsubscribeValidator(validatorID uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnsubscribeValidator(&_Contract.TransactOpts, validatorID)
}

// UnsubscribeValidators is a paid mutator transaction binding the contract method 0x3ae15892.
//
// Solidity: function unsubscribeValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactor) UnsubscribeValidators(opts *bind.TransactOpts, validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unsubscribeValidators", validatorIDArray)
}

// UnsubscribeValidators is a paid mutator transaction binding the contract method 0x3ae15892.
//
// Solidity: function unsubscribeValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractSession) UnsubscribeValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnsubscribeValidators(&_Contract.TransactOpts, validatorIDArray)
}

// UnsubscribeValidators is a paid mutator transaction binding the contract method 0x3ae15892.
//
// Solidity: function unsubscribeValidators(uint64[] validatorIDArray) returns()
func (_Contract *ContractTransactorSession) UnsubscribeValidators(validatorIDArray []uint64) (*types.Transaction, error) {
	return _Contract.Contract.UnsubscribeValidators(&_Contract.TransactOpts, validatorIDArray)
}

// UpdateCheckpointSlotSize is a paid mutator transaction binding the contract method 0x38d092b9.
//
// Solidity: function updateCheckpointSlotSize(uint64 newCheckpointSlotSize) returns()
func (_Contract *ContractTransactor) UpdateCheckpointSlotSize(opts *bind.TransactOpts, newCheckpointSlotSize uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateCheckpointSlotSize", newCheckpointSlotSize)
}

// UpdateCheckpointSlotSize is a paid mutator transaction binding the contract method 0x38d092b9.
//
// Solidity: function updateCheckpointSlotSize(uint64 newCheckpointSlotSize) returns()
func (_Contract *ContractSession) UpdateCheckpointSlotSize(newCheckpointSlotSize uint64) (*types.Transaction, error) {
	return _Contract.Contract.UpdateCheckpointSlotSize(&_Contract.TransactOpts, newCheckpointSlotSize)
}

// UpdateCheckpointSlotSize is a paid mutator transaction binding the contract method 0x38d092b9.
//
// Solidity: function updateCheckpointSlotSize(uint64 newCheckpointSlotSize) returns()
func (_Contract *ContractTransactorSession) UpdateCheckpointSlotSize(newCheckpointSlotSize uint64) (*types.Transaction, error) {
	return _Contract.Contract.UpdateCheckpointSlotSize(&_Contract.TransactOpts, newCheckpointSlotSize)
}

// UpdateCollateral is a paid mutator transaction binding the contract method 0x6721de26.
//
// Solidity: function updateCollateral(uint256 newSubscriptionCollateral) returns()
func (_Contract *ContractTransactor) UpdateCollateral(opts *bind.TransactOpts, newSubscriptionCollateral *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateCollateral", newSubscriptionCollateral)
}

// UpdateCollateral is a paid mutator transaction binding the contract method 0x6721de26.
//
// Solidity: function updateCollateral(uint256 newSubscriptionCollateral) returns()
func (_Contract *ContractSession) UpdateCollateral(newSubscriptionCollateral *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateCollateral(&_Contract.TransactOpts, newSubscriptionCollateral)
}

// UpdateCollateral is a paid mutator transaction binding the contract method 0x6721de26.
//
// Solidity: function updateCollateral(uint256 newSubscriptionCollateral) returns()
func (_Contract *ContractTransactorSession) UpdateCollateral(newSubscriptionCollateral *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateCollateral(&_Contract.TransactOpts, newSubscriptionCollateral)
}

// UpdatePoolFee is a paid mutator transaction binding the contract method 0xc2cac04b.
//
// Solidity: function updatePoolFee(uint256 newPoolFee) returns()
func (_Contract *ContractTransactor) UpdatePoolFee(opts *bind.TransactOpts, newPoolFee *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updatePoolFee", newPoolFee)
}

// UpdatePoolFee is a paid mutator transaction binding the contract method 0xc2cac04b.
//
// Solidity: function updatePoolFee(uint256 newPoolFee) returns()
func (_Contract *ContractSession) UpdatePoolFee(newPoolFee *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdatePoolFee(&_Contract.TransactOpts, newPoolFee)
}

// UpdatePoolFee is a paid mutator transaction binding the contract method 0xc2cac04b.
//
// Solidity: function updatePoolFee(uint256 newPoolFee) returns()
func (_Contract *ContractTransactorSession) UpdatePoolFee(newPoolFee *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdatePoolFee(&_Contract.TransactOpts, newPoolFee)
}

// UpdatePoolFeeRecipient is a paid mutator transaction binding the contract method 0x6cf9dfad.
//
// Solidity: function updatePoolFeeRecipient(address newPoolFeeRecipient) returns()
func (_Contract *ContractTransactor) UpdatePoolFeeRecipient(opts *bind.TransactOpts, newPoolFeeRecipient common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updatePoolFeeRecipient", newPoolFeeRecipient)
}

// UpdatePoolFeeRecipient is a paid mutator transaction binding the contract method 0x6cf9dfad.
//
// Solidity: function updatePoolFeeRecipient(address newPoolFeeRecipient) returns()
func (_Contract *ContractSession) UpdatePoolFeeRecipient(newPoolFeeRecipient common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdatePoolFeeRecipient(&_Contract.TransactOpts, newPoolFeeRecipient)
}

// UpdatePoolFeeRecipient is a paid mutator transaction binding the contract method 0x6cf9dfad.
//
// Solidity: function updatePoolFeeRecipient(address newPoolFeeRecipient) returns()
func (_Contract *ContractTransactorSession) UpdatePoolFeeRecipient(newPoolFeeRecipient common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdatePoolFeeRecipient(&_Contract.TransactOpts, newPoolFeeRecipient)
}

// UpdateQuorum is a paid mutator transaction binding the contract method 0x29218b61.
//
// Solidity: function updateQuorum(uint64 newQuorum) returns()
func (_Contract *ContractTransactor) UpdateQuorum(opts *bind.TransactOpts, newQuorum uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateQuorum", newQuorum)
}

// UpdateQuorum is a paid mutator transaction binding the contract method 0x29218b61.
//
// Solidity: function updateQuorum(uint64 newQuorum) returns()
func (_Contract *ContractSession) UpdateQuorum(newQuorum uint64) (*types.Transaction, error) {
	return _Contract.Contract.UpdateQuorum(&_Contract.TransactOpts, newQuorum)
}

// UpdateQuorum is a paid mutator transaction binding the contract method 0x29218b61.
//
// Solidity: function updateQuorum(uint64 newQuorum) returns()
func (_Contract *ContractTransactorSession) UpdateQuorum(newQuorum uint64) (*types.Transaction, error) {
	return _Contract.Contract.UpdateQuorum(&_Contract.TransactOpts, newQuorum)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Contract *ContractTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Contract.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Contract *ContractSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Contract.Contract.Fallback(&_Contract.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Contract *ContractTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Contract.Contract.Fallback(&_Contract.TransactOpts, calldata)
}

// ContractAcceptGovernanceIterator is returned from FilterAcceptGovernance and is used to iterate over the raw logs and unpacked data for AcceptGovernance events raised by the Contract contract.
type ContractAcceptGovernanceIterator struct {
	Event *ContractAcceptGovernance // Event containing the contract specifics and raw log

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
func (it *ContractAcceptGovernanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAcceptGovernance)
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
		it.Event = new(ContractAcceptGovernance)
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
func (it *ContractAcceptGovernanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAcceptGovernanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAcceptGovernance represents a AcceptGovernance event raised by the Contract contract.
type ContractAcceptGovernance struct {
	NewGovernance common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAcceptGovernance is a free log retrieval operation binding the contract event 0x0e5e627abed15db8c4841ff7db9a3fb94e105b243564c206bf485362210eee07.
//
// Solidity: event AcceptGovernance(address newGovernance)
func (_Contract *ContractFilterer) FilterAcceptGovernance(opts *bind.FilterOpts) (*ContractAcceptGovernanceIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AcceptGovernance")
	if err != nil {
		return nil, err
	}
	return &ContractAcceptGovernanceIterator{contract: _Contract.contract, event: "AcceptGovernance", logs: logs, sub: sub}, nil
}

// WatchAcceptGovernance is a free log subscription operation binding the contract event 0x0e5e627abed15db8c4841ff7db9a3fb94e105b243564c206bf485362210eee07.
//
// Solidity: event AcceptGovernance(address newGovernance)
func (_Contract *ContractFilterer) WatchAcceptGovernance(opts *bind.WatchOpts, sink chan<- *ContractAcceptGovernance) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AcceptGovernance")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAcceptGovernance)
				if err := _Contract.contract.UnpackLog(event, "AcceptGovernance", log); err != nil {
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

// ParseAcceptGovernance is a log parse operation binding the contract event 0x0e5e627abed15db8c4841ff7db9a3fb94e105b243564c206bf485362210eee07.
//
// Solidity: event AcceptGovernance(address newGovernance)
func (_Contract *ContractFilterer) ParseAcceptGovernance(log types.Log) (*ContractAcceptGovernance, error) {
	event := new(ContractAcceptGovernance)
	if err := _Contract.contract.UnpackLog(event, "AcceptGovernance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractAddOracleMemberIterator is returned from FilterAddOracleMember and is used to iterate over the raw logs and unpacked data for AddOracleMember events raised by the Contract contract.
type ContractAddOracleMemberIterator struct {
	Event *ContractAddOracleMember // Event containing the contract specifics and raw log

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
func (it *ContractAddOracleMemberIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAddOracleMember)
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
		it.Event = new(ContractAddOracleMember)
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
func (it *ContractAddOracleMemberIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAddOracleMemberIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAddOracleMember represents a AddOracleMember event raised by the Contract contract.
type ContractAddOracleMember struct {
	NewOracleMember common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterAddOracleMember is a free log retrieval operation binding the contract event 0x82ebad05b594f3bb43fed0280ee782c47f15549310ffb9de21ad790a03dbab18.
//
// Solidity: event AddOracleMember(address newOracleMember)
func (_Contract *ContractFilterer) FilterAddOracleMember(opts *bind.FilterOpts) (*ContractAddOracleMemberIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AddOracleMember")
	if err != nil {
		return nil, err
	}
	return &ContractAddOracleMemberIterator{contract: _Contract.contract, event: "AddOracleMember", logs: logs, sub: sub}, nil
}

// WatchAddOracleMember is a free log subscription operation binding the contract event 0x82ebad05b594f3bb43fed0280ee782c47f15549310ffb9de21ad790a03dbab18.
//
// Solidity: event AddOracleMember(address newOracleMember)
func (_Contract *ContractFilterer) WatchAddOracleMember(opts *bind.WatchOpts, sink chan<- *ContractAddOracleMember) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AddOracleMember")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAddOracleMember)
				if err := _Contract.contract.UnpackLog(event, "AddOracleMember", log); err != nil {
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

// ParseAddOracleMember is a log parse operation binding the contract event 0x82ebad05b594f3bb43fed0280ee782c47f15549310ffb9de21ad790a03dbab18.
//
// Solidity: event AddOracleMember(address newOracleMember)
func (_Contract *ContractFilterer) ParseAddOracleMember(log types.Log) (*ContractAddOracleMember, error) {
	event := new(ContractAddOracleMember)
	if err := _Contract.contract.UnpackLog(event, "AddOracleMember", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractBanValidatorIterator is returned from FilterBanValidator and is used to iterate over the raw logs and unpacked data for BanValidator events raised by the Contract contract.
type ContractBanValidatorIterator struct {
	Event *ContractBanValidator // Event containing the contract specifics and raw log

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
func (it *ContractBanValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractBanValidator)
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
		it.Event = new(ContractBanValidator)
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
func (it *ContractBanValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractBanValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractBanValidator represents a BanValidator event raised by the Contract contract.
type ContractBanValidator struct {
	ValidatorID uint64
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBanValidator is a free log retrieval operation binding the contract event 0xb6978e4c15338f096fdd02b00c667cb63eeedd4df4f2243213e9371dbc5d1ca7.
//
// Solidity: event BanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) FilterBanValidator(opts *bind.FilterOpts) (*ContractBanValidatorIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "BanValidator")
	if err != nil {
		return nil, err
	}
	return &ContractBanValidatorIterator{contract: _Contract.contract, event: "BanValidator", logs: logs, sub: sub}, nil
}

// WatchBanValidator is a free log subscription operation binding the contract event 0xb6978e4c15338f096fdd02b00c667cb63eeedd4df4f2243213e9371dbc5d1ca7.
//
// Solidity: event BanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) WatchBanValidator(opts *bind.WatchOpts, sink chan<- *ContractBanValidator) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "BanValidator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractBanValidator)
				if err := _Contract.contract.UnpackLog(event, "BanValidator", log); err != nil {
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

// ParseBanValidator is a log parse operation binding the contract event 0xb6978e4c15338f096fdd02b00c667cb63eeedd4df4f2243213e9371dbc5d1ca7.
//
// Solidity: event BanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) ParseBanValidator(log types.Log) (*ContractBanValidator, error) {
	event := new(ContractBanValidator)
	if err := _Contract.contract.UnpackLog(event, "BanValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractClaimRewardsIterator is returned from FilterClaimRewards and is used to iterate over the raw logs and unpacked data for ClaimRewards events raised by the Contract contract.
type ContractClaimRewardsIterator struct {
	Event *ContractClaimRewards // Event containing the contract specifics and raw log

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
func (it *ContractClaimRewardsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractClaimRewards)
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
		it.Event = new(ContractClaimRewards)
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
func (it *ContractClaimRewardsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractClaimRewardsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractClaimRewards represents a ClaimRewards event raised by the Contract contract.
type ContractClaimRewards struct {
	WithdrawalAddress common.Address
	RewardAddress     common.Address
	ClaimableBalance  *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterClaimRewards is a free log retrieval operation binding the contract event 0x9aa05b3d70a9e3e2f004f039648839560576334fb45c81f91b6db03ad9e2efc9.
//
// Solidity: event ClaimRewards(address withdrawalAddress, address rewardAddress, uint256 claimableBalance)
func (_Contract *ContractFilterer) FilterClaimRewards(opts *bind.FilterOpts) (*ContractClaimRewardsIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ClaimRewards")
	if err != nil {
		return nil, err
	}
	return &ContractClaimRewardsIterator{contract: _Contract.contract, event: "ClaimRewards", logs: logs, sub: sub}, nil
}

// WatchClaimRewards is a free log subscription operation binding the contract event 0x9aa05b3d70a9e3e2f004f039648839560576334fb45c81f91b6db03ad9e2efc9.
//
// Solidity: event ClaimRewards(address withdrawalAddress, address rewardAddress, uint256 claimableBalance)
func (_Contract *ContractFilterer) WatchClaimRewards(opts *bind.WatchOpts, sink chan<- *ContractClaimRewards) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ClaimRewards")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractClaimRewards)
				if err := _Contract.contract.UnpackLog(event, "ClaimRewards", log); err != nil {
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

// ParseClaimRewards is a log parse operation binding the contract event 0x9aa05b3d70a9e3e2f004f039648839560576334fb45c81f91b6db03ad9e2efc9.
//
// Solidity: event ClaimRewards(address withdrawalAddress, address rewardAddress, uint256 claimableBalance)
func (_Contract *ContractFilterer) ParseClaimRewards(log types.Log) (*ContractClaimRewards, error) {
	event := new(ContractClaimRewards)
	if err := _Contract.contract.UnpackLog(event, "ClaimRewards", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractEtherReceivedIterator is returned from FilterEtherReceived and is used to iterate over the raw logs and unpacked data for EtherReceived events raised by the Contract contract.
type ContractEtherReceivedIterator struct {
	Event *ContractEtherReceived // Event containing the contract specifics and raw log

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
func (it *ContractEtherReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractEtherReceived)
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
		it.Event = new(ContractEtherReceived)
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
func (it *ContractEtherReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractEtherReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractEtherReceived represents a EtherReceived event raised by the Contract contract.
type ContractEtherReceived struct {
	Sender         common.Address
	DonationAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterEtherReceived is a free log retrieval operation binding the contract event 0x1e57e3bb474320be3d2c77138f75b7c3941292d647f5f9634e33a8e94e0e069b.
//
// Solidity: event EtherReceived(address sender, uint256 donationAmount)
func (_Contract *ContractFilterer) FilterEtherReceived(opts *bind.FilterOpts) (*ContractEtherReceivedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "EtherReceived")
	if err != nil {
		return nil, err
	}
	return &ContractEtherReceivedIterator{contract: _Contract.contract, event: "EtherReceived", logs: logs, sub: sub}, nil
}

// WatchEtherReceived is a free log subscription operation binding the contract event 0x1e57e3bb474320be3d2c77138f75b7c3941292d647f5f9634e33a8e94e0e069b.
//
// Solidity: event EtherReceived(address sender, uint256 donationAmount)
func (_Contract *ContractFilterer) WatchEtherReceived(opts *bind.WatchOpts, sink chan<- *ContractEtherReceived) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "EtherReceived")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractEtherReceived)
				if err := _Contract.contract.UnpackLog(event, "EtherReceived", log); err != nil {
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

// ParseEtherReceived is a log parse operation binding the contract event 0x1e57e3bb474320be3d2c77138f75b7c3941292d647f5f9634e33a8e94e0e069b.
//
// Solidity: event EtherReceived(address sender, uint256 donationAmount)
func (_Contract *ContractFilterer) ParseEtherReceived(log types.Log) (*ContractEtherReceived, error) {
	event := new(ContractEtherReceived)
	if err := _Contract.contract.UnpackLog(event, "EtherReceived", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractInitSmoothingPoolIterator is returned from FilterInitSmoothingPool and is used to iterate over the raw logs and unpacked data for InitSmoothingPool events raised by the Contract contract.
type ContractInitSmoothingPoolIterator struct {
	Event *ContractInitSmoothingPool // Event containing the contract specifics and raw log

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
func (it *ContractInitSmoothingPoolIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractInitSmoothingPool)
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
		it.Event = new(ContractInitSmoothingPool)
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
func (it *ContractInitSmoothingPoolIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractInitSmoothingPoolIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractInitSmoothingPool represents a InitSmoothingPool event raised by the Contract contract.
type ContractInitSmoothingPool struct {
	InitialSmoothingPoolSlot uint64
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterInitSmoothingPool is a free log retrieval operation binding the contract event 0x517462a977504b91ea9a39b7a880cff34a3a13d734f1b294ae4eb0e5c603c7d0.
//
// Solidity: event InitSmoothingPool(uint64 initialSmoothingPoolSlot)
func (_Contract *ContractFilterer) FilterInitSmoothingPool(opts *bind.FilterOpts) (*ContractInitSmoothingPoolIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "InitSmoothingPool")
	if err != nil {
		return nil, err
	}
	return &ContractInitSmoothingPoolIterator{contract: _Contract.contract, event: "InitSmoothingPool", logs: logs, sub: sub}, nil
}

// WatchInitSmoothingPool is a free log subscription operation binding the contract event 0x517462a977504b91ea9a39b7a880cff34a3a13d734f1b294ae4eb0e5c603c7d0.
//
// Solidity: event InitSmoothingPool(uint64 initialSmoothingPoolSlot)
func (_Contract *ContractFilterer) WatchInitSmoothingPool(opts *bind.WatchOpts, sink chan<- *ContractInitSmoothingPool) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "InitSmoothingPool")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractInitSmoothingPool)
				if err := _Contract.contract.UnpackLog(event, "InitSmoothingPool", log); err != nil {
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

// ParseInitSmoothingPool is a log parse operation binding the contract event 0x517462a977504b91ea9a39b7a880cff34a3a13d734f1b294ae4eb0e5c603c7d0.
//
// Solidity: event InitSmoothingPool(uint64 initialSmoothingPoolSlot)
func (_Contract *ContractFilterer) ParseInitSmoothingPool(log types.Log) (*ContractInitSmoothingPool, error) {
	event := new(ContractInitSmoothingPool)
	if err := _Contract.contract.UnpackLog(event, "InitSmoothingPool", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Contract contract.
type ContractInitializedIterator struct {
	Event *ContractInitialized // Event containing the contract specifics and raw log

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
func (it *ContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractInitialized)
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
		it.Event = new(ContractInitialized)
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
func (it *ContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractInitialized represents a Initialized event raised by the Contract contract.
type ContractInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractInitializedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractInitializedIterator{contract: _Contract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractInitialized) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractInitialized)
				if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) ParseInitialized(log types.Log) (*ContractInitialized, error) {
	event := new(ContractInitialized)
	if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Contract contract.
type ContractOwnershipTransferredIterator struct {
	Event *ContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractOwnershipTransferred)
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
		it.Event = new(ContractOwnershipTransferred)
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
func (it *ContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractOwnershipTransferred represents a OwnershipTransferred event raised by the Contract contract.
type ContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractOwnershipTransferredIterator{contract: _Contract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractOwnershipTransferred)
				if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Contract *ContractFilterer) ParseOwnershipTransferred(log types.Log) (*ContractOwnershipTransferred, error) {
	event := new(ContractOwnershipTransferred)
	if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRemoveOracleMemberIterator is returned from FilterRemoveOracleMember and is used to iterate over the raw logs and unpacked data for RemoveOracleMember events raised by the Contract contract.
type ContractRemoveOracleMemberIterator struct {
	Event *ContractRemoveOracleMember // Event containing the contract specifics and raw log

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
func (it *ContractRemoveOracleMemberIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRemoveOracleMember)
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
		it.Event = new(ContractRemoveOracleMember)
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
func (it *ContractRemoveOracleMemberIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRemoveOracleMemberIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRemoveOracleMember represents a RemoveOracleMember event raised by the Contract contract.
type ContractRemoveOracleMember struct {
	OracleMemberRemoved common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterRemoveOracleMember is a free log retrieval operation binding the contract event 0xc8391e8d83bfa93da9636d7a7928b59021752c6e0b74afba127e74914af730a2.
//
// Solidity: event RemoveOracleMember(address oracleMemberRemoved)
func (_Contract *ContractFilterer) FilterRemoveOracleMember(opts *bind.FilterOpts) (*ContractRemoveOracleMemberIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RemoveOracleMember")
	if err != nil {
		return nil, err
	}
	return &ContractRemoveOracleMemberIterator{contract: _Contract.contract, event: "RemoveOracleMember", logs: logs, sub: sub}, nil
}

// WatchRemoveOracleMember is a free log subscription operation binding the contract event 0xc8391e8d83bfa93da9636d7a7928b59021752c6e0b74afba127e74914af730a2.
//
// Solidity: event RemoveOracleMember(address oracleMemberRemoved)
func (_Contract *ContractFilterer) WatchRemoveOracleMember(opts *bind.WatchOpts, sink chan<- *ContractRemoveOracleMember) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RemoveOracleMember")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRemoveOracleMember)
				if err := _Contract.contract.UnpackLog(event, "RemoveOracleMember", log); err != nil {
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

// ParseRemoveOracleMember is a log parse operation binding the contract event 0xc8391e8d83bfa93da9636d7a7928b59021752c6e0b74afba127e74914af730a2.
//
// Solidity: event RemoveOracleMember(address oracleMemberRemoved)
func (_Contract *ContractFilterer) ParseRemoveOracleMember(log types.Log) (*ContractRemoveOracleMember, error) {
	event := new(ContractRemoveOracleMember)
	if err := _Contract.contract.UnpackLog(event, "RemoveOracleMember", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractReportConsolidatedIterator is returned from FilterReportConsolidated and is used to iterate over the raw logs and unpacked data for ReportConsolidated events raised by the Contract contract.
type ContractReportConsolidatedIterator struct {
	Event *ContractReportConsolidated // Event containing the contract specifics and raw log

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
func (it *ContractReportConsolidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractReportConsolidated)
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
		it.Event = new(ContractReportConsolidated)
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
func (it *ContractReportConsolidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractReportConsolidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractReportConsolidated represents a ReportConsolidated event raised by the Contract contract.
type ContractReportConsolidated struct {
	SlotNumber     *big.Int
	NewRewardsRoot [32]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterReportConsolidated is a free log retrieval operation binding the contract event 0x92f1a3bddcbec48ac79b5809ed37bf0491f0ad9e89ed4ff2f1ccd4dd9e5b5064.
//
// Solidity: event ReportConsolidated(uint256 slotNumber, bytes32 newRewardsRoot)
func (_Contract *ContractFilterer) FilterReportConsolidated(opts *bind.FilterOpts) (*ContractReportConsolidatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ReportConsolidated")
	if err != nil {
		return nil, err
	}
	return &ContractReportConsolidatedIterator{contract: _Contract.contract, event: "ReportConsolidated", logs: logs, sub: sub}, nil
}

// WatchReportConsolidated is a free log subscription operation binding the contract event 0x92f1a3bddcbec48ac79b5809ed37bf0491f0ad9e89ed4ff2f1ccd4dd9e5b5064.
//
// Solidity: event ReportConsolidated(uint256 slotNumber, bytes32 newRewardsRoot)
func (_Contract *ContractFilterer) WatchReportConsolidated(opts *bind.WatchOpts, sink chan<- *ContractReportConsolidated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ReportConsolidated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractReportConsolidated)
				if err := _Contract.contract.UnpackLog(event, "ReportConsolidated", log); err != nil {
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

// ParseReportConsolidated is a log parse operation binding the contract event 0x92f1a3bddcbec48ac79b5809ed37bf0491f0ad9e89ed4ff2f1ccd4dd9e5b5064.
//
// Solidity: event ReportConsolidated(uint256 slotNumber, bytes32 newRewardsRoot)
func (_Contract *ContractFilterer) ParseReportConsolidated(log types.Log) (*ContractReportConsolidated, error) {
	event := new(ContractReportConsolidated)
	if err := _Contract.contract.UnpackLog(event, "ReportConsolidated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractSetRewardRecipientIterator is returned from FilterSetRewardRecipient and is used to iterate over the raw logs and unpacked data for SetRewardRecipient events raised by the Contract contract.
type ContractSetRewardRecipientIterator struct {
	Event *ContractSetRewardRecipient // Event containing the contract specifics and raw log

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
func (it *ContractSetRewardRecipientIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractSetRewardRecipient)
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
		it.Event = new(ContractSetRewardRecipient)
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
func (it *ContractSetRewardRecipientIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractSetRewardRecipientIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractSetRewardRecipient represents a SetRewardRecipient event raised by the Contract contract.
type ContractSetRewardRecipient struct {
	WithdrawalAddress common.Address
	PoolRecipient     common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSetRewardRecipient is a free log retrieval operation binding the contract event 0xc6b66e0e282673c442421e1c6b89458b7631f26f5dcd0b2b216c45831ca1d7d5.
//
// Solidity: event SetRewardRecipient(address withdrawalAddress, address poolRecipient)
func (_Contract *ContractFilterer) FilterSetRewardRecipient(opts *bind.FilterOpts) (*ContractSetRewardRecipientIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "SetRewardRecipient")
	if err != nil {
		return nil, err
	}
	return &ContractSetRewardRecipientIterator{contract: _Contract.contract, event: "SetRewardRecipient", logs: logs, sub: sub}, nil
}

// WatchSetRewardRecipient is a free log subscription operation binding the contract event 0xc6b66e0e282673c442421e1c6b89458b7631f26f5dcd0b2b216c45831ca1d7d5.
//
// Solidity: event SetRewardRecipient(address withdrawalAddress, address poolRecipient)
func (_Contract *ContractFilterer) WatchSetRewardRecipient(opts *bind.WatchOpts, sink chan<- *ContractSetRewardRecipient) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "SetRewardRecipient")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractSetRewardRecipient)
				if err := _Contract.contract.UnpackLog(event, "SetRewardRecipient", log); err != nil {
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

// ParseSetRewardRecipient is a log parse operation binding the contract event 0xc6b66e0e282673c442421e1c6b89458b7631f26f5dcd0b2b216c45831ca1d7d5.
//
// Solidity: event SetRewardRecipient(address withdrawalAddress, address poolRecipient)
func (_Contract *ContractFilterer) ParseSetRewardRecipient(log types.Log) (*ContractSetRewardRecipient, error) {
	event := new(ContractSetRewardRecipient)
	if err := _Contract.contract.UnpackLog(event, "SetRewardRecipient", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractSubmitReportIterator is returned from FilterSubmitReport and is used to iterate over the raw logs and unpacked data for SubmitReport events raised by the Contract contract.
type ContractSubmitReportIterator struct {
	Event *ContractSubmitReport // Event containing the contract specifics and raw log

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
func (it *ContractSubmitReportIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractSubmitReport)
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
		it.Event = new(ContractSubmitReport)
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
func (it *ContractSubmitReportIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractSubmitReportIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractSubmitReport represents a SubmitReport event raised by the Contract contract.
type ContractSubmitReport struct {
	SlotNumber     *big.Int
	NewRewardsRoot [32]byte
	OracleMember   common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSubmitReport is a free log retrieval operation binding the contract event 0xa28058c376f3af4acde9ae77dfe7fc66bc12abc2ad14e37fe60d51e19a630571.
//
// Solidity: event SubmitReport(uint256 slotNumber, bytes32 newRewardsRoot, address oracleMember)
func (_Contract *ContractFilterer) FilterSubmitReport(opts *bind.FilterOpts) (*ContractSubmitReportIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "SubmitReport")
	if err != nil {
		return nil, err
	}
	return &ContractSubmitReportIterator{contract: _Contract.contract, event: "SubmitReport", logs: logs, sub: sub}, nil
}

// WatchSubmitReport is a free log subscription operation binding the contract event 0xa28058c376f3af4acde9ae77dfe7fc66bc12abc2ad14e37fe60d51e19a630571.
//
// Solidity: event SubmitReport(uint256 slotNumber, bytes32 newRewardsRoot, address oracleMember)
func (_Contract *ContractFilterer) WatchSubmitReport(opts *bind.WatchOpts, sink chan<- *ContractSubmitReport) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "SubmitReport")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractSubmitReport)
				if err := _Contract.contract.UnpackLog(event, "SubmitReport", log); err != nil {
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

// ParseSubmitReport is a log parse operation binding the contract event 0xa28058c376f3af4acde9ae77dfe7fc66bc12abc2ad14e37fe60d51e19a630571.
//
// Solidity: event SubmitReport(uint256 slotNumber, bytes32 newRewardsRoot, address oracleMember)
func (_Contract *ContractFilterer) ParseSubmitReport(log types.Log) (*ContractSubmitReport, error) {
	event := new(ContractSubmitReport)
	if err := _Contract.contract.UnpackLog(event, "SubmitReport", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractSubscribeValidatorIterator is returned from FilterSubscribeValidator and is used to iterate over the raw logs and unpacked data for SubscribeValidator events raised by the Contract contract.
type ContractSubscribeValidatorIterator struct {
	Event *ContractSubscribeValidator // Event containing the contract specifics and raw log

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
func (it *ContractSubscribeValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractSubscribeValidator)
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
		it.Event = new(ContractSubscribeValidator)
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
func (it *ContractSubscribeValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractSubscribeValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractSubscribeValidator represents a SubscribeValidator event raised by the Contract contract.
type ContractSubscribeValidator struct {
	Sender                 common.Address
	SubscriptionCollateral *big.Int
	ValidatorID            uint64
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterSubscribeValidator is a free log retrieval operation binding the contract event 0x1094f8cfeb6abd0fd67e4ce1e1d3999c0176a4d8c7f8325e3ecddb5a1249fde9.
//
// Solidity: event SubscribeValidator(address sender, uint256 subscriptionCollateral, uint64 validatorID)
func (_Contract *ContractFilterer) FilterSubscribeValidator(opts *bind.FilterOpts) (*ContractSubscribeValidatorIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "SubscribeValidator")
	if err != nil {
		return nil, err
	}
	return &ContractSubscribeValidatorIterator{contract: _Contract.contract, event: "SubscribeValidator", logs: logs, sub: sub}, nil
}

// WatchSubscribeValidator is a free log subscription operation binding the contract event 0x1094f8cfeb6abd0fd67e4ce1e1d3999c0176a4d8c7f8325e3ecddb5a1249fde9.
//
// Solidity: event SubscribeValidator(address sender, uint256 subscriptionCollateral, uint64 validatorID)
func (_Contract *ContractFilterer) WatchSubscribeValidator(opts *bind.WatchOpts, sink chan<- *ContractSubscribeValidator) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "SubscribeValidator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractSubscribeValidator)
				if err := _Contract.contract.UnpackLog(event, "SubscribeValidator", log); err != nil {
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

// ParseSubscribeValidator is a log parse operation binding the contract event 0x1094f8cfeb6abd0fd67e4ce1e1d3999c0176a4d8c7f8325e3ecddb5a1249fde9.
//
// Solidity: event SubscribeValidator(address sender, uint256 subscriptionCollateral, uint64 validatorID)
func (_Contract *ContractFilterer) ParseSubscribeValidator(log types.Log) (*ContractSubscribeValidator, error) {
	event := new(ContractSubscribeValidator)
	if err := _Contract.contract.UnpackLog(event, "SubscribeValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTransferGovernanceIterator is returned from FilterTransferGovernance and is used to iterate over the raw logs and unpacked data for TransferGovernance events raised by the Contract contract.
type ContractTransferGovernanceIterator struct {
	Event *ContractTransferGovernance // Event containing the contract specifics and raw log

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
func (it *ContractTransferGovernanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTransferGovernance)
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
		it.Event = new(ContractTransferGovernance)
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
func (it *ContractTransferGovernanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTransferGovernanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTransferGovernance represents a TransferGovernance event raised by the Contract contract.
type ContractTransferGovernance struct {
	NewPendingGovernance common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterTransferGovernance is a free log retrieval operation binding the contract event 0xde4aabcd09171142d82dd9e667db43bf0dca12f30fa0aec30859875d35ecb5d6.
//
// Solidity: event TransferGovernance(address newPendingGovernance)
func (_Contract *ContractFilterer) FilterTransferGovernance(opts *bind.FilterOpts) (*ContractTransferGovernanceIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "TransferGovernance")
	if err != nil {
		return nil, err
	}
	return &ContractTransferGovernanceIterator{contract: _Contract.contract, event: "TransferGovernance", logs: logs, sub: sub}, nil
}

// WatchTransferGovernance is a free log subscription operation binding the contract event 0xde4aabcd09171142d82dd9e667db43bf0dca12f30fa0aec30859875d35ecb5d6.
//
// Solidity: event TransferGovernance(address newPendingGovernance)
func (_Contract *ContractFilterer) WatchTransferGovernance(opts *bind.WatchOpts, sink chan<- *ContractTransferGovernance) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "TransferGovernance")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTransferGovernance)
				if err := _Contract.contract.UnpackLog(event, "TransferGovernance", log); err != nil {
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

// ParseTransferGovernance is a log parse operation binding the contract event 0xde4aabcd09171142d82dd9e667db43bf0dca12f30fa0aec30859875d35ecb5d6.
//
// Solidity: event TransferGovernance(address newPendingGovernance)
func (_Contract *ContractFilterer) ParseTransferGovernance(log types.Log) (*ContractTransferGovernance, error) {
	event := new(ContractTransferGovernance)
	if err := _Contract.contract.UnpackLog(event, "TransferGovernance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUnbanValidatorIterator is returned from FilterUnbanValidator and is used to iterate over the raw logs and unpacked data for UnbanValidator events raised by the Contract contract.
type ContractUnbanValidatorIterator struct {
	Event *ContractUnbanValidator // Event containing the contract specifics and raw log

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
func (it *ContractUnbanValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUnbanValidator)
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
		it.Event = new(ContractUnbanValidator)
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
func (it *ContractUnbanValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUnbanValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUnbanValidator represents a UnbanValidator event raised by the Contract contract.
type ContractUnbanValidator struct {
	ValidatorID uint64
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnbanValidator is a free log retrieval operation binding the contract event 0xaa0f7dc1145b1a2a0195373eca47d4c7a42be75b0579288456153a416494a581.
//
// Solidity: event UnbanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) FilterUnbanValidator(opts *bind.FilterOpts) (*ContractUnbanValidatorIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UnbanValidator")
	if err != nil {
		return nil, err
	}
	return &ContractUnbanValidatorIterator{contract: _Contract.contract, event: "UnbanValidator", logs: logs, sub: sub}, nil
}

// WatchUnbanValidator is a free log subscription operation binding the contract event 0xaa0f7dc1145b1a2a0195373eca47d4c7a42be75b0579288456153a416494a581.
//
// Solidity: event UnbanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) WatchUnbanValidator(opts *bind.WatchOpts, sink chan<- *ContractUnbanValidator) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UnbanValidator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUnbanValidator)
				if err := _Contract.contract.UnpackLog(event, "UnbanValidator", log); err != nil {
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

// ParseUnbanValidator is a log parse operation binding the contract event 0xaa0f7dc1145b1a2a0195373eca47d4c7a42be75b0579288456153a416494a581.
//
// Solidity: event UnbanValidator(uint64 validatorID)
func (_Contract *ContractFilterer) ParseUnbanValidator(log types.Log) (*ContractUnbanValidator, error) {
	event := new(ContractUnbanValidator)
	if err := _Contract.contract.UnpackLog(event, "UnbanValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUnsubscribeValidatorIterator is returned from FilterUnsubscribeValidator and is used to iterate over the raw logs and unpacked data for UnsubscribeValidator events raised by the Contract contract.
type ContractUnsubscribeValidatorIterator struct {
	Event *ContractUnsubscribeValidator // Event containing the contract specifics and raw log

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
func (it *ContractUnsubscribeValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUnsubscribeValidator)
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
		it.Event = new(ContractUnsubscribeValidator)
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
func (it *ContractUnsubscribeValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUnsubscribeValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUnsubscribeValidator represents a UnsubscribeValidator event raised by the Contract contract.
type ContractUnsubscribeValidator struct {
	Sender      common.Address
	ValidatorID uint64
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnsubscribeValidator is a free log retrieval operation binding the contract event 0x5a984d13ffecb21f69de43956dcd971abcff8e187fac09a0ce3209562da14e0a.
//
// Solidity: event UnsubscribeValidator(address sender, uint64 validatorID)
func (_Contract *ContractFilterer) FilterUnsubscribeValidator(opts *bind.FilterOpts) (*ContractUnsubscribeValidatorIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UnsubscribeValidator")
	if err != nil {
		return nil, err
	}
	return &ContractUnsubscribeValidatorIterator{contract: _Contract.contract, event: "UnsubscribeValidator", logs: logs, sub: sub}, nil
}

// WatchUnsubscribeValidator is a free log subscription operation binding the contract event 0x5a984d13ffecb21f69de43956dcd971abcff8e187fac09a0ce3209562da14e0a.
//
// Solidity: event UnsubscribeValidator(address sender, uint64 validatorID)
func (_Contract *ContractFilterer) WatchUnsubscribeValidator(opts *bind.WatchOpts, sink chan<- *ContractUnsubscribeValidator) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UnsubscribeValidator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUnsubscribeValidator)
				if err := _Contract.contract.UnpackLog(event, "UnsubscribeValidator", log); err != nil {
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

// ParseUnsubscribeValidator is a log parse operation binding the contract event 0x5a984d13ffecb21f69de43956dcd971abcff8e187fac09a0ce3209562da14e0a.
//
// Solidity: event UnsubscribeValidator(address sender, uint64 validatorID)
func (_Contract *ContractFilterer) ParseUnsubscribeValidator(log types.Log) (*ContractUnsubscribeValidator, error) {
	event := new(ContractUnsubscribeValidator)
	if err := _Contract.contract.UnpackLog(event, "UnsubscribeValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateCheckpointSlotSizeIterator is returned from FilterUpdateCheckpointSlotSize and is used to iterate over the raw logs and unpacked data for UpdateCheckpointSlotSize events raised by the Contract contract.
type ContractUpdateCheckpointSlotSizeIterator struct {
	Event *ContractUpdateCheckpointSlotSize // Event containing the contract specifics and raw log

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
func (it *ContractUpdateCheckpointSlotSizeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateCheckpointSlotSize)
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
		it.Event = new(ContractUpdateCheckpointSlotSize)
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
func (it *ContractUpdateCheckpointSlotSizeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateCheckpointSlotSizeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateCheckpointSlotSize represents a UpdateCheckpointSlotSize event raised by the Contract contract.
type ContractUpdateCheckpointSlotSize struct {
	NewCheckpointSlotSize uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterUpdateCheckpointSlotSize is a free log retrieval operation binding the contract event 0x8e0e8e986a04eea90f6e33488f9756f53cb482049ca8269e6864c797b8bcae6e.
//
// Solidity: event UpdateCheckpointSlotSize(uint64 newCheckpointSlotSize)
func (_Contract *ContractFilterer) FilterUpdateCheckpointSlotSize(opts *bind.FilterOpts) (*ContractUpdateCheckpointSlotSizeIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateCheckpointSlotSize")
	if err != nil {
		return nil, err
	}
	return &ContractUpdateCheckpointSlotSizeIterator{contract: _Contract.contract, event: "UpdateCheckpointSlotSize", logs: logs, sub: sub}, nil
}

// WatchUpdateCheckpointSlotSize is a free log subscription operation binding the contract event 0x8e0e8e986a04eea90f6e33488f9756f53cb482049ca8269e6864c797b8bcae6e.
//
// Solidity: event UpdateCheckpointSlotSize(uint64 newCheckpointSlotSize)
func (_Contract *ContractFilterer) WatchUpdateCheckpointSlotSize(opts *bind.WatchOpts, sink chan<- *ContractUpdateCheckpointSlotSize) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateCheckpointSlotSize")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateCheckpointSlotSize)
				if err := _Contract.contract.UnpackLog(event, "UpdateCheckpointSlotSize", log); err != nil {
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

// ParseUpdateCheckpointSlotSize is a log parse operation binding the contract event 0x8e0e8e986a04eea90f6e33488f9756f53cb482049ca8269e6864c797b8bcae6e.
//
// Solidity: event UpdateCheckpointSlotSize(uint64 newCheckpointSlotSize)
func (_Contract *ContractFilterer) ParseUpdateCheckpointSlotSize(log types.Log) (*ContractUpdateCheckpointSlotSize, error) {
	event := new(ContractUpdateCheckpointSlotSize)
	if err := _Contract.contract.UnpackLog(event, "UpdateCheckpointSlotSize", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdatePoolFeeIterator is returned from FilterUpdatePoolFee and is used to iterate over the raw logs and unpacked data for UpdatePoolFee events raised by the Contract contract.
type ContractUpdatePoolFeeIterator struct {
	Event *ContractUpdatePoolFee // Event containing the contract specifics and raw log

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
func (it *ContractUpdatePoolFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatePoolFee)
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
		it.Event = new(ContractUpdatePoolFee)
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
func (it *ContractUpdatePoolFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatePoolFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatePoolFee represents a UpdatePoolFee event raised by the Contract contract.
type ContractUpdatePoolFee struct {
	NewPoolFee *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterUpdatePoolFee is a free log retrieval operation binding the contract event 0x19d74da91b7de020180f04c1c60faba431bd76ecf962935e6f65ecbf0223ecfc.
//
// Solidity: event UpdatePoolFee(uint256 newPoolFee)
func (_Contract *ContractFilterer) FilterUpdatePoolFee(opts *bind.FilterOpts) (*ContractUpdatePoolFeeIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatePoolFee")
	if err != nil {
		return nil, err
	}
	return &ContractUpdatePoolFeeIterator{contract: _Contract.contract, event: "UpdatePoolFee", logs: logs, sub: sub}, nil
}

// WatchUpdatePoolFee is a free log subscription operation binding the contract event 0x19d74da91b7de020180f04c1c60faba431bd76ecf962935e6f65ecbf0223ecfc.
//
// Solidity: event UpdatePoolFee(uint256 newPoolFee)
func (_Contract *ContractFilterer) WatchUpdatePoolFee(opts *bind.WatchOpts, sink chan<- *ContractUpdatePoolFee) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatePoolFee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatePoolFee)
				if err := _Contract.contract.UnpackLog(event, "UpdatePoolFee", log); err != nil {
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

// ParseUpdatePoolFee is a log parse operation binding the contract event 0x19d74da91b7de020180f04c1c60faba431bd76ecf962935e6f65ecbf0223ecfc.
//
// Solidity: event UpdatePoolFee(uint256 newPoolFee)
func (_Contract *ContractFilterer) ParseUpdatePoolFee(log types.Log) (*ContractUpdatePoolFee, error) {
	event := new(ContractUpdatePoolFee)
	if err := _Contract.contract.UnpackLog(event, "UpdatePoolFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdatePoolFeeRecipientIterator is returned from FilterUpdatePoolFeeRecipient and is used to iterate over the raw logs and unpacked data for UpdatePoolFeeRecipient events raised by the Contract contract.
type ContractUpdatePoolFeeRecipientIterator struct {
	Event *ContractUpdatePoolFeeRecipient // Event containing the contract specifics and raw log

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
func (it *ContractUpdatePoolFeeRecipientIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatePoolFeeRecipient)
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
		it.Event = new(ContractUpdatePoolFeeRecipient)
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
func (it *ContractUpdatePoolFeeRecipientIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatePoolFeeRecipientIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatePoolFeeRecipient represents a UpdatePoolFeeRecipient event raised by the Contract contract.
type ContractUpdatePoolFeeRecipient struct {
	NewPoolFeeRecipient common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterUpdatePoolFeeRecipient is a free log retrieval operation binding the contract event 0xae901f1a96a9fc6852e6b162ea0b9887c37b667fb5d2d925b6e4a607aac0bf62.
//
// Solidity: event UpdatePoolFeeRecipient(address newPoolFeeRecipient)
func (_Contract *ContractFilterer) FilterUpdatePoolFeeRecipient(opts *bind.FilterOpts) (*ContractUpdatePoolFeeRecipientIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatePoolFeeRecipient")
	if err != nil {
		return nil, err
	}
	return &ContractUpdatePoolFeeRecipientIterator{contract: _Contract.contract, event: "UpdatePoolFeeRecipient", logs: logs, sub: sub}, nil
}

// WatchUpdatePoolFeeRecipient is a free log subscription operation binding the contract event 0xae901f1a96a9fc6852e6b162ea0b9887c37b667fb5d2d925b6e4a607aac0bf62.
//
// Solidity: event UpdatePoolFeeRecipient(address newPoolFeeRecipient)
func (_Contract *ContractFilterer) WatchUpdatePoolFeeRecipient(opts *bind.WatchOpts, sink chan<- *ContractUpdatePoolFeeRecipient) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatePoolFeeRecipient")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatePoolFeeRecipient)
				if err := _Contract.contract.UnpackLog(event, "UpdatePoolFeeRecipient", log); err != nil {
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

// ParseUpdatePoolFeeRecipient is a log parse operation binding the contract event 0xae901f1a96a9fc6852e6b162ea0b9887c37b667fb5d2d925b6e4a607aac0bf62.
//
// Solidity: event UpdatePoolFeeRecipient(address newPoolFeeRecipient)
func (_Contract *ContractFilterer) ParseUpdatePoolFeeRecipient(log types.Log) (*ContractUpdatePoolFeeRecipient, error) {
	event := new(ContractUpdatePoolFeeRecipient)
	if err := _Contract.contract.UnpackLog(event, "UpdatePoolFeeRecipient", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateQuorumIterator is returned from FilterUpdateQuorum and is used to iterate over the raw logs and unpacked data for UpdateQuorum events raised by the Contract contract.
type ContractUpdateQuorumIterator struct {
	Event *ContractUpdateQuorum // Event containing the contract specifics and raw log

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
func (it *ContractUpdateQuorumIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateQuorum)
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
		it.Event = new(ContractUpdateQuorum)
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
func (it *ContractUpdateQuorumIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateQuorumIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateQuorum represents a UpdateQuorum event raised by the Contract contract.
type ContractUpdateQuorum struct {
	NewQuorum uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUpdateQuorum is a free log retrieval operation binding the contract event 0xb600f3cf7f38a4b49bb0c75f722ef69f7e3e39ef3bb4aa8207fd86e724a23249.
//
// Solidity: event UpdateQuorum(uint64 newQuorum)
func (_Contract *ContractFilterer) FilterUpdateQuorum(opts *bind.FilterOpts) (*ContractUpdateQuorumIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateQuorum")
	if err != nil {
		return nil, err
	}
	return &ContractUpdateQuorumIterator{contract: _Contract.contract, event: "UpdateQuorum", logs: logs, sub: sub}, nil
}

// WatchUpdateQuorum is a free log subscription operation binding the contract event 0xb600f3cf7f38a4b49bb0c75f722ef69f7e3e39ef3bb4aa8207fd86e724a23249.
//
// Solidity: event UpdateQuorum(uint64 newQuorum)
func (_Contract *ContractFilterer) WatchUpdateQuorum(opts *bind.WatchOpts, sink chan<- *ContractUpdateQuorum) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateQuorum")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateQuorum)
				if err := _Contract.contract.UnpackLog(event, "UpdateQuorum", log); err != nil {
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

// ParseUpdateQuorum is a log parse operation binding the contract event 0xb600f3cf7f38a4b49bb0c75f722ef69f7e3e39ef3bb4aa8207fd86e724a23249.
//
// Solidity: event UpdateQuorum(uint64 newQuorum)
func (_Contract *ContractFilterer) ParseUpdateQuorum(log types.Log) (*ContractUpdateQuorum, error) {
	event := new(ContractUpdateQuorum)
	if err := _Contract.contract.UnpackLog(event, "UpdateQuorum", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateSubscriptionCollateralIterator is returned from FilterUpdateSubscriptionCollateral and is used to iterate over the raw logs and unpacked data for UpdateSubscriptionCollateral events raised by the Contract contract.
type ContractUpdateSubscriptionCollateralIterator struct {
	Event *ContractUpdateSubscriptionCollateral // Event containing the contract specifics and raw log

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
func (it *ContractUpdateSubscriptionCollateralIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateSubscriptionCollateral)
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
		it.Event = new(ContractUpdateSubscriptionCollateral)
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
func (it *ContractUpdateSubscriptionCollateralIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateSubscriptionCollateralIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateSubscriptionCollateral represents a UpdateSubscriptionCollateral event raised by the Contract contract.
type ContractUpdateSubscriptionCollateral struct {
	NewSubscriptionCollateral *big.Int
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterUpdateSubscriptionCollateral is a free log retrieval operation binding the contract event 0xdb50d3f51ff6294ba829f0d7f6b99b5b606a52807d5106ef44151d9297720217.
//
// Solidity: event UpdateSubscriptionCollateral(uint256 newSubscriptionCollateral)
func (_Contract *ContractFilterer) FilterUpdateSubscriptionCollateral(opts *bind.FilterOpts) (*ContractUpdateSubscriptionCollateralIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateSubscriptionCollateral")
	if err != nil {
		return nil, err
	}
	return &ContractUpdateSubscriptionCollateralIterator{contract: _Contract.contract, event: "UpdateSubscriptionCollateral", logs: logs, sub: sub}, nil
}

// WatchUpdateSubscriptionCollateral is a free log subscription operation binding the contract event 0xdb50d3f51ff6294ba829f0d7f6b99b5b606a52807d5106ef44151d9297720217.
//
// Solidity: event UpdateSubscriptionCollateral(uint256 newSubscriptionCollateral)
func (_Contract *ContractFilterer) WatchUpdateSubscriptionCollateral(opts *bind.WatchOpts, sink chan<- *ContractUpdateSubscriptionCollateral) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateSubscriptionCollateral")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateSubscriptionCollateral)
				if err := _Contract.contract.UnpackLog(event, "UpdateSubscriptionCollateral", log); err != nil {
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

// ParseUpdateSubscriptionCollateral is a log parse operation binding the contract event 0xdb50d3f51ff6294ba829f0d7f6b99b5b606a52807d5106ef44151d9297720217.
//
// Solidity: event UpdateSubscriptionCollateral(uint256 newSubscriptionCollateral)
func (_Contract *ContractFilterer) ParseUpdateSubscriptionCollateral(log types.Log) (*ContractUpdateSubscriptionCollateral, error) {
	event := new(ContractUpdateSubscriptionCollateral)
	if err := _Contract.contract.UnpackLog(event, "UpdateSubscriptionCollateral", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
