package oracle

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	api "github.com/attestantio/go-eth2-client/api"
	eth2 "github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/avast/retry-go/v4"
	"github.com/dappnode/mev-sp-oracle/config"
	"github.com/dappnode/mev-sp-oracle/constants"
	"github.com/dappnode/mev-sp-oracle/contract"
	"github.com/dappnode/mev-sp-oracle/eigenLayer/manager"
	"github.com/dappnode/mev-sp-oracle/eigenLayer/pod"
	"github.com/dappnode/mev-sp-oracle/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	log "github.com/sirupsen/logrus"
)

var MainnetChainId = uint64(1)
var GoerliChainId = uint64(5)
var HoleskyChainId = uint64(17000)

// This file provides different functions to access the blockchain state from both consensus and
// execution layer and modifying the its state via smart contract calls.
type EpochDuties struct {
	Epoch  uint64
	Duties []*v1.ProposerDuty
}

// Simple cache storing epoch -> proposer duties
// This is useful to not query the beacon node for each slot
// since ProposerDuties returns the duties for the whole epoch
// Note that the cache is meant to store only one epoch's duties
var ProposalDutyCache EpochDuties

type Onchain struct {
	ConsensusClient      *http.Service
	ExecutionClient      *ethclient.Client
	Contract             *contract.Contract
	EigenManagerContract *manager.ManagerCaller
	NumRetries           int
	updaterKey           *ecdsa.PrivateKey
	UpdaterAddress       common.Address
	PoolAddress          string
	ChainId              uint64
	validators           map[phase0.ValidatorIndex]*v1.Validator
}

func NewOnchain(cliCfg *config.CliConfig, updaterKey *ecdsa.PrivateKey) (*Onchain, error) {

	// Dial the execution client
	executionClient, err := ethclient.Dial(cliCfg.ExecutionEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "Error dialing execution client")
	}

	// Get chainid to ensure the endpoint is working
	chainId, err := executionClient.ChainID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching chainid from execution client")
	}
	log.Info("Connected succesfully to execution client. ChainId: ", chainId)

	// Dial the consensus client
	client, err := http.New(context.Background(),
		http.WithTimeout(120*time.Second),
		http.WithAddress(cliCfg.ConsensusEndpoint),
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error dialing consensus client")
	}
	consensusClient := client.(*http.Service)

	// Get deposit contract to ensure the endpoint is working
	depositContract, err := consensusClient.DepositContract(context.Background(), &api.DepositContractOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching deposit contract from consensus client")
	}
	log.Info("Connected succesfully to consensus client. ChainId: ", depositContract.Data.ChainID,
		" DepositContract: ", "0x"+hex.EncodeToString(depositContract.Data.Address[:]))

	if depositContract.Data.ChainID != uint64(chainId.Int64()) {
		return nil, errors.Wrap(err, fmt.Sprintf("Consensus and execution clients are not connected to the same chain %d vs %d",
			depositContract.Data.ChainID, chainId))
	}

	// Print sync status of consensus and execution client
	execSync, err := executionClient.SyncProgress(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching execution client sync progress")
	}

	// nil means synced
	if execSync == nil {
		header, err := executionClient.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return nil, errors.Wrap(err, "Error fetching execution client header")
		}
		log.Info("Execution client is in sync, block number: ", header.Number)
	} else {
		log.Info("Execution client is NOT in sync, current block: ", execSync.CurrentBlock)
	}

	consSync, err := consensusClient.NodeSyncing(context.Background(), &api.NodeSyncingOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching consensus client sync progress")
	}

	if consSync.Data.SyncDistance == 0 {
		log.Info("Consensus client is in sync, head slot: ", consSync.Data.HeadSlot)
	} else {
		log.Info("Consensus client is NOT in sync, slots behind: ", consSync.Data.SyncDistance)
	}

	// Instantiate the smoothing pool contract to run get/set operations on it
	address := common.HexToAddress(cliCfg.PoolAddress)
	contract, err := contract.NewContract(address, executionClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error instantiating contract")
	}

	// Instantiate the EigenManager contract to run get/set operations on it
	eigenManagerAddress := common.HexToAddress(cliCfg.EigenManagerAddress)
	eigenManagerContract, err := manager.NewManagerCaller(eigenManagerAddress, executionClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error instantiating EigenManager contract")
	}

	var updaterAddress common.Address
	if updaterKey != nil {
		updaterAddress = crypto.PubkeyToAddress(updaterKey.PublicKey)
	}

	return &Onchain{
		ConsensusClient:      consensusClient,
		ExecutionClient:      executionClient,
		PoolAddress:          cliCfg.PoolAddress,
		Contract:             contract,
		EigenManagerContract: eigenManagerContract,
		NumRetries:           cliCfg.NumRetries,
		ChainId:              uint64(chainId.Int64()),
		updaterKey:           updaterKey,
		UpdaterAddress:       updaterAddress,
	}, nil
}

// Gets the Pod from a given owner address. If no pods are found, returns an empty address
func (o *Onchain) GetOwnerToPod(address common.Address, opts ...retry.Option) (common.Address, error) {
	var podAddress common.Address
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			podAddress, err = o.EigenManagerContract.OwnerToPod(callOpts, address)
			if err != nil {
				log.Warn("Failed attempt to get pod from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get pod from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)
	if err != nil {
		return common.Address{}, errors.New("could not get pod from contract: " + err.Error())
	}
	return podAddress, nil
}

// GetValidatorRestakedEvents fetches all "ValidatorRestaked" events from the Pod
// contract at podAddress, returning a slice of the validator indices.
// TODO: think if we could filter by a specific validator index.
func (o *Onchain) GetValidatorRestakedEvents(
	podAddress common.Address,
	opts ...retry.Option,
) ([]*big.Int, error) {

	var validatorIndices []*big.Int

	// Retry logic in case of timeouts or transient failures
	err := retry.Do(
		func() error {
			// Instantiate the Pod contract
			podContract, err := pod.NewPod(podAddress, o.ExecutionClient)
			if err != nil {
				log.Warn("Failed to instantiate Pod contract: ", err)
				return errors.Wrap(err, "failed to instantiate Pod contract")
			}

			// Filter from block 0 up to the latest block (nil)
			filterOpts := &bind.FilterOpts{
				Start:   0,
				End:     nil,
				Context: context.Background(),
			}

			iter, err := podContract.FilterValidatorRestaked(filterOpts)
			if err != nil {
				log.Warn("Failed to filter ValidatorRestaked events: ", err)
				return errors.Wrap(err, "failed to filter ValidatorRestaked events")
			}
			defer iter.Close()

			// Iterate over all events found
			for iter.Next() {
				event := iter.Event
				// Collect the validator index
				validatorIndices = append(validatorIndices, event.ValidatorIndex)
			}

			// Check for iteration errors
			if err := iter.Error(); err != nil {
				log.Warn("Error iterating over ValidatorRestaked events: ", err)
				return errors.Wrap(err, "error iterating over ValidatorRestaked events")
			}

			return nil
		},
		o.GetRetryOpts(opts)...,
	)

	if err != nil {
		return nil, err
	}

	return validatorIndices, nil
}

func (o *Onchain) AreNodesInSync(opts ...retry.Option) (bool, error) {
	var err error
	var execSync *ethereum.SyncProgress
	var consSync *api.Response[*v1.SyncState]

	err = retry.Do(func() error {
		execSync, err = o.ExecutionClient.SyncProgress(context.Background())
		if err != nil {
			log.Warn("Failed attempt to fetch execution client sync progress: ", err.Error(), " Retrying...")
			return errors.New("Error fetching execution client sync progress: " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return false, errors.New("Could not fetch execution client sync progress: " + err.Error())
	}

	err = retry.Do(func() error {
		consSync, err = o.ConsensusClient.NodeSyncing(context.Background(), &api.NodeSyncingOpts{})
		if err != nil {
			log.Warn("Failed attempt to fetch consensus client sync progress: ", err.Error(), " Retrying...")
			return errors.New("Error fetching execution client sync progress: " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return false, errors.New("Could not fetch consensus client sync progress: " + err.Error())
	}

	// If no errors arised while fetching the sync progress of both clients, check if the clients are in sync
	// If the execution client is not nil, it means it is not synced
	if execSync != nil {
		log.Info("Exec client not in sync")
		return false, nil
	}

	// If the sync distance is greater than 5, the consensus client is not in sync
	if consSync.Data.SyncDistance > 5 {
		log.Info("Consensus client not in sync, Client is more than 5 slots behind")
		return false, nil
	}

	// if both clients are synced, return true
	return true, nil
}

func (o *Onchain) GetConsensusBlockAtSlot(slot uint64, opts ...retry.Option) (*spec.VersionedSignedBeaconBlock, error) {
	slotStr := strconv.FormatUint(slot, 10)
	var signedBeaconBlock *api.Response[*spec.VersionedSignedBeaconBlock]
	var err error

	err = retry.Do(func() error {
		signedBeaconBlock, err = o.ConsensusClient.SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
			Block: slotStr, // TODO: Says block but its slot in reality
		})
		if err != nil {
			// Its not possible to detect if the block was missed or if the block doesn't exist in the chain
			// GET failed with status 404: {"code":404,"message":"NOT_FOUND: beacon block at slot 7408169","stacktraces":[]}
			// To keep compatibility after go-eth2-client changes, we consider that a missed block is nil
			if strings.Contains(err.Error(), "404") {
				signedBeaconBlock = &api.Response[*spec.VersionedSignedBeaconBlock]{
					Data: nil,
				}
				return nil
			}
			log.Warn("Failed attempt to fetch block at slot ", slotStr, ": ", err.Error(), " Retrying...")
			return errors.New("Error fetching block at slot " + slotStr + ": " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Could not fetch block at slot " + slotStr + ": " + err.Error())
	}
	return signedBeaconBlock.Data, err
}

// Gets active validators by asking the chain. It does not get current Onchain object validators
func (o *Onchain) GetFinalizedValidators(opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	var validators *api.Response[map[phase0.ValidatorIndex]*v1.Validator]
	var err error

	err = retry.Do(func() error {
		validators, err = o.ConsensusClient.Validators(context.Background(), &api.ValidatorsOpts{
			State: "finalized",
			// Empty Indices means no filter = get all validators
		})
		if err != nil {
			log.Warn("Failed attempt to fetch finalized validators: ", err.Error(), " Retrying...")
			return errors.New("Error fetching finalized validators: " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Could not fetch finalized validators: " + err.Error())
	}
	return validators.Data, err
}

func (o *Onchain) FinalizedBeaconBlockHeader(opts ...retry.Option) (*v1.BeaconBlockHeader, error) {
	var beaconBlockHeader *api.Response[*v1.BeaconBlockHeader]
	var err error

	err = retry.Do(func() error {
		beaconBlockHeader, err = o.ConsensusClient.BeaconBlockHeader(context.Background(), &eth2.BeaconBlockHeaderOpts{
			Block: "finalized",
		})
		if err != nil {
			log.Warn("Failed attempt to fetch finalized beacon block header: ", err.Error(), " Retrying...")
			return errors.New("Error fetching finalized  beacon block header: " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Could not fetch finalized beacon block header: " + err.Error())
	}
	return beaconBlockHeader.Data, err
}

func (o *Onchain) GetSingleValidator(valIndex phase0.ValidatorIndex, slot string, opts ...retry.Option) (*v1.Validator, error) {
	var validators *api.Response[map[phase0.ValidatorIndex]*v1.Validator]
	var err error

	err = retry.Do(func() error {
		validatorIndices := []phase0.ValidatorIndex{valIndex}
		validators, err = o.ConsensusClient.Validators(context.Background(), &api.ValidatorsOpts{
			State:   slot,
			Indices: validatorIndices,
		})

		if err != nil {
			log.Warn("Failed attempt to fetch validator: ", err.Error(), " Retrying...")
			return errors.New("Error fetching validator: " + err.Error())
		}

		if len(validators.Data) > 1 {
			return errors.New("Error fetching validator: Requested one but got many")
		}

		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Could not fetch validator: " + err.Error())
	}

	// If empty, it means the validator doesnt exist
	if len(validators.Data) == 0 {
		return nil, nil
	}

	// Some sanity checks
	validator, found := validators.Data[valIndex]
	if !found {
		return nil, errors.New(fmt.Sprintf("Error fetching validator: Could not find index in response: %d",
			valIndex))
	}
	if validator.Index != valIndex {
		return nil, errors.New(fmt.Sprintf("Error fetching validator: Index mismatch in response: %d vs %d",
			valIndex, validator.Index))
	}
	return validator, err
}
func (o *Onchain) GetSetOfValidators(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	// If empty, return an error. We do this to avoid fetching all validators. GetSetOfValidators should be used with a set of indices
	if len(valIndices) == 0 {
		return nil, errors.New("Error: validator indices are empty")
	}

	// Fetch all validators at once
	var validators *api.Response[map[phase0.ValidatorIndex]*v1.Validator]
	var err error

	err = retry.Do(func() error {
		// Fetch all validators at once. Library will automatically batch the requests
		validators, err = o.ConsensusClient.Validators(context.Background(), &api.ValidatorsOpts{
			State:   slot,
			Indices: valIndices,
		})

		if err != nil {
			log.Warn("Failed attempt to fetch validators: ", err.Error(), " Retrying...")
			return errors.New("Error fetching validators: " + err.Error())
		}

		return nil
	}, o.GetRetryOpts(opts)...)

	// If there was an error fetching the validators, return it.
	if err != nil {
		return nil, errors.New("Could not fetch validators: " + err.Error())
	}

	// If the data is empty, return an error. This should never happen.
	if len(validators.Data) == 0 {
		return nil, errors.New("Error: no validators found onchain for the given indices")
	}

	// Sanity checks - Ensure all requested validators are found
	for _, idx := range valIndices {
		if _, found := validators.Data[idx]; !found {
			return nil, errors.New(fmt.Sprintf("Error fetching validators: Could not find index in response: %d", idx))
		}
	}

	return validators.Data, nil
}

func (o *Onchain) BlockByNumber(blockNumber *big.Int, opts ...retry.Option) (*types.Block, error) {
	var err error
	var block *types.Block

	err = retry.Do(func() error {
		block, err = o.ExecutionClient.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			log.Warn("Failed attempt to fetch block by number: ", err.Error(), " Retrying...")
			return errors.New("Error fetching block by number: " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Could not fetch block by number: " + err.Error())
	}
	return block, err
}

func (o *Onchain) GetProposalDuty(slot uint64, opts ...retry.Option) (*v1.ProposerDuty, error) {
	epoch := slot / constants.SlotsInEpoch
	slotWithinEpoch := slot % constants.SlotsInEpoch
	slotStr := strconv.FormatUint(slot, 10)

	// If cache hit, return the result
	if ProposalDutyCache.Epoch == epoch {
		// Sanity check that should never happen
		if ProposalDutyCache.Epoch != uint64(ProposalDutyCache.Duties[slotWithinEpoch].Slot/phase0.Slot(constants.SlotsInEpoch)) {
			return nil, errors.New("Proposal duty epoch does not match when converting slot to epoch")
		}
		return ProposalDutyCache.Duties[slotWithinEpoch], nil
	}

	// Empty indexes to force fetching all duties
	indexes := make([]phase0.ValidatorIndex, 0)
	var duties *api.Response[[]*v1.ProposerDuty]
	var err error

	err = retry.Do(func() error {
		duties, err = o.ConsensusClient.ProposerDuties(
			context.Background(), &api.ProposerDutiesOpts{
				Epoch:   phase0.Epoch(epoch),
				Indices: indexes,
			})
		if err != nil {
			log.Warn("Failed attempt to fetch proposal duties at slot ", slotStr, ": ", err.Error(), " Retrying...")
			return errors.New("Error fetching proposal duties at slot " + slotStr + ": " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("Error fetching proposal duties at slot " + slotStr + ": " + err.Error())
	}

	// If success, store result in cache
	ProposalDutyCache = EpochDuties{epoch, duties.Data}

	return duties.Data[slotWithinEpoch], nil
}

// This function is expensive as gets every tx receipt from the block. Use only if needed
func (o *Onchain) GetExecHeaderAndReceipts(
	blockNumber *big.Int,
	rawTxs []bellatrix.Transaction,
	opts ...retry.Option) (*types.Header, []*types.Receipt, error) {

	var header *types.Header
	var err error

	err = retry.Do(func() error {
		header, err = o.ExecutionClient.HeaderByNumber(context.Background(), blockNumber)
		if err != nil {
			log.Warn("Failed attempt to fetch header for block ", blockNumber.String(), ": ", err.Error(), " Retrying...")
			return errors.New("Error fetching header for block " + blockNumber.String() + ": " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, nil, errors.New("Could not fetch header for block " + blockNumber.String() + ": " + err.Error())
	}

	var receipts []*types.Receipt = make([]*types.Receipt, 0)
	for _, rawTx := range rawTxs {
		// This should never happen
		tx, err := utils.DecodeTx(rawTx)
		if err != nil {
			log.Fatal(err)
		}
		var receipt *types.Receipt

		err = retry.Do(func() error {
			receipt, err = o.ExecutionClient.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Warn("Failed attempt to fetch receipt for tx ", tx.Hash().String(), ": ", err.Error(), " Retrying...")
				return errors.New("Error fetching receipt for tx " + tx.Hash().String() + ": " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

		if err != nil {
			return nil, nil, errors.New("Could not fetch receipt for tx " + tx.Hash().String() + ": " + err.Error())
		}
		receipts = append(receipts, receipt)
	}
	return header, receipts, nil
}

func (o *Onchain) IsAddressWhitelisted(address common.Address, opts ...retry.Option) (bool, error) {
	var whitelistedAddresses []common.Address
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			whitelistedAddresses, err = o.Contract.GetAllOracleMembers(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get all oracle members from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get all oracle members from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return false, errors.New("could not get all oracle members from contract: " + err.Error())
	}
	for _, whitelistedAddress := range whitelistedAddresses {
		if whitelistedAddress == address {
			return true, nil
		}
	}
	return false, nil
}

type ReportType struct {
	Slot  uint64
	Votes uint64
}

func (o *Onchain) GetAddressToVotedReport(address common.Address, opts ...retry.Option) (ReportType, error) {
	var reportHash [32]byte
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			reportHash, err = o.Contract.AddressToVotedReportHash(callOpts, address)
			if err != nil {
				log.Warn("Failed attempt to get address to voted report from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get address to voted report from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return ReportType{}, errors.New("could not get address to voted report from contract: " + err.Error())
	}

	var report ReportType
	err = retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			report, err = o.Contract.ReportHashToReport(callOpts, reportHash)
			if err != nil {
				log.Warn("Failed attempt to get report from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get report from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)
	if err != nil {
		return ReportType{}, errors.New("could not get report from contract: " + err.Error())
	}
	return report, nil
}

func (o *Onchain) GetAllOracleMembers(opts ...retry.Option) ([]common.Address, error) {
	var addresses []common.Address
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			addresses, err = o.Contract.GetAllOracleMembers(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get all oracle members from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get all oracle members from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get all oracle members from contract: " + err.Error())
	}

	return addresses, nil
}

func (o *Onchain) GetQuorum(opts ...retry.Option) (uint64, error) {
	var quorum uint64
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			quorum, err = o.Contract.Quorum(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get quorum from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get quorum from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return 0, errors.New("could not get quorum from contract: " + err.Error())
	}
	return quorum, nil
}

func (o *Onchain) GetContractCollateral(opts ...retry.Option) (*big.Int, error) {
	subscriptionCollateral := new(big.Int)
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			var err error
			subscriptionCollateral, err = o.Contract.SubscriptionCollateral(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get subscription collateral from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get subscription collateral from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get subscription collateral from contract: " + err.Error())
	}
	return subscriptionCollateral, nil
}

func (o *Onchain) GetSlotCheckpointSize(opts ...retry.Option) (uint64, error) {
	var slotCheckpointSize uint64
	var err error

	err = retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			slotCheckpointSize, err = o.Contract.CheckpointSlotSize(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get slot checkpoint size from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get slot checkpoint size from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return 0, errors.New("could not get claimed balance from contract: " + err.Error())
	}

	return slotCheckpointSize, nil
}

func (o *Onchain) GetContractDeploymentBlock(opts ...retry.Option) (*big.Int, error) {
	var deploymentBlock *big.Int
	var err error

	err = retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			deploymentBlock, err = o.Contract.DeploymentBlockNumber(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get deployment block from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get deployment block from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get deployment block from contract: " + err.Error())
	}

	return deploymentBlock, nil
}

func (o *Onchain) GetPoolFee(opts ...retry.Option) (*big.Int, error) {
	var poolFee *big.Int
	var err error

	err = retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			poolFee, err = o.Contract.PoolFee(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get pool fee from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get pool fee from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get pool fee from contract: " + err.Error())
	}

	return poolFee, nil
}

func (o *Onchain) GetPoolFeeAddress(opts ...retry.Option) (string, error) {
	var poolFeeAddress common.Address
	var err error

	err = retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			poolFeeAddress, err = o.Contract.PoolFeeRecipient(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get pool fee address from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get pool fee address from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return "", errors.New("could not get pool fee address from contract: " + err.Error())
	}

	return poolFeeAddress.Hex(), nil
}

func (o *Onchain) GetRewardsRoot(opts ...retry.Option) (string, error) {
	var rewardsRootStr string

	// Retries multiple times before errorings
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			rewardsRoot, err := o.Contract.RewardsRoot(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get merkle root from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get rewards root from contract: " + err.Error())
			}
			rewardsRootStr = hexutil.Encode(rewardsRoot[:])
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return "", errors.New("could not get merkle root from contract: " + err.Error())
	}

	return rewardsRootStr, nil
}

func (o *Onchain) GetLastConsolidatedSlot(opts ...retry.Option) (uint64, error) {
	var lastConsolidatedSlot uint64

	// Retries multiple times before errorings
	err := retry.Do(
		func() error {
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			contractLastConsolidatedSlot, err := o.Contract.LastConsolidatedSlot(callOpts)
			if err != nil {
				log.Warn("Failed attempt to get last consolidated slot from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get last consolidated slot from contract: " + err.Error())
			}
			lastConsolidatedSlot = contractLastConsolidatedSlot
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return 0, errors.New("could not get last consolidated slot from contract: " + err.Error())
	}

	return lastConsolidatedSlot, nil
}

func (o *Onchain) GetOnchainSlotAndRoot(opts ...retry.Option) (string, uint64, error) {
	slot, err := o.GetLastConsolidatedSlot(opts...)
	if err != nil {
		return "", 0, errors.Wrap(err, "could not get last consolidated slot")
	}

	merkleRoot, err := o.GetRewardsRoot(opts...)
	if err != nil {
		return "", 0, errors.Wrap(err, "could not get merkle root")
	}

	return merkleRoot, slot, nil
}

// Returns the claimed balance of each address. Note that when using this for reconciliation you must run it
// in finalized blocks to be able to compare apples with apples.
func (o *Onchain) GetContractClaimedBalance(withdrawalAddress string, blockNumber *big.Int, opts ...retry.Option) (*big.Int, error) {
	var claimedBalance *big.Int
	var err error

	if !common.IsHexAddress(withdrawalAddress) {
		log.Fatal("Invalid withdrawal address: ", withdrawalAddress)
	}

	hexDepAddres := common.HexToAddress(withdrawalAddress)

	// Retries multiple times before errorings
	err = retry.Do(
		func() error {
			// Otherwise our local view and remote view can be different. See if it applies to other functions, like merkle tree.
			callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}
			if blockNumber != nil {
				callOpts.BlockNumber = blockNumber
			}
			claimedBalance, err = o.Contract.ClaimedBalance(callOpts, hexDepAddres)
			if err != nil {
				log.Warn("Failed attempt to get claimed balance from contract: ", err.Error(), " Retrying...")
				return errors.New("could not get claimed balance from contract: " + err.Error())
			}
			return nil
		}, o.GetRetryOpts(opts)...)

	if err != nil {
		return big.NewInt(0), errors.New("could not get claimed balance from contract: " + err.Error())
	}

	return claimedBalance, nil
}

func (o *Onchain) GetPoolEthBalance(blockNumber *big.Int, opts ...retry.Option) (*big.Int, error) {
	account := common.HexToAddress(o.PoolAddress)
	var err error
	var balanceWei *big.Int

	err = retry.Do(func() error {
		// If block number is nil latest known block is used
		balanceWei, err = o.ExecutionClient.BalanceAt(context.Background(), account, blockNumber)
		if err != nil {
			log.Warn("Failed attempt to get balance for pool address ", account.String(), ": ", err.Error(), " Retrying...")
			return errors.New("could not get balance for pool address " + account.String() + ": " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get balance for pool address " + account.String() + ": " + err.Error())
	}

	return balanceWei, nil
}

func (o *Onchain) GetAddressEthBalance(account common.Address, opts ...retry.Option) (*big.Int, error) {
	var err error
	var balanceWei *big.Int

	err = retry.Do(func() error {
		// If block number is nil latest known block is used
		balanceWei, err = o.ExecutionClient.BalanceAt(context.Background(), account, nil)
		if err != nil {
			log.Warn("Failed attempt to get balance for address ", account.String(), ": ", err.Error(), " Retrying...")
			return errors.New("could not get balance for address " + account.String() + ": " + err.Error())
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.New("could not get balance for address " + account.String() + ": " + err.Error())
	}

	return balanceWei, nil
}

// Oracle dependancy is injected here since we need to know wether the validator
// proposing the block at this slot is i) subscribed to the pool or ii) its reward
// goes to the pool. This allows to fetch less information on the blocks that are
// not relevant to the pool. If fetchAll is enabled, the whole content of the block
// is fetched no matter what, just for debugging purposes, will slow down sync
func (o *Onchain) FetchFullBlock(slot uint64, oracle *Oracle, opt ...bool) *FullBlock {
	var fetchAll bool
	if len(opt) > 1 {
		log.Fatal("invalid number of arguments, just one opt is allowed")
	} else if len(opt) == 1 {
		fetchAll = opt[0]
	} else {
		fetchAll = false
	}

	// Get who should propose the block
	slotDuty, err := o.GetProposalDuty(slot)
	if err != nil {
		log.Fatal("could not get proposal duty: ", err)
	}

	// Sanity check to ensure the slot duty is the one we requested
	if uint64(slotDuty.Slot) != slot {
		log.Fatal("slot duty slot does not match requested slot: ", slotDuty.Slot, " vs ", slot)
	}

	// Get the validator info that proposed (or should have proposed) the block
	currentSlotStr := strconv.FormatUint(slot, 10)
	validator, err := o.GetSingleValidator(slotDuty.ValidatorIndex, currentSlotStr)
	if err != nil {
		log.Fatal("could not get single validator: ", err)
	}

	// Create the full block with the duty, which is the minimum info it can have
	fullBlock := NewFullBlock(slotDuty, validator, o.ChainId)

	// Fetch the whole consensus block
	proposedBlock, err := o.GetConsensusBlockAtSlot(slot)
	if err != nil {
		log.Fatal("could not get block at slot:", err)
	}

	if proposedBlock == nil {
		// Mised block, nothing to do
	} else {
		// Succesfull proposal, fetch the info we need
		fullBlock.SetConsensusBlock(proposedBlock)

		// Sanity check to ensure the block is the one we requested
		if fullBlock.GetSlotUint64() != slot {
			log.Fatal("slot does not match requested slot: ", fullBlock.GetSlotUint64(), " vs ", slot)
		}

		etherReceived, err := o.GetEtherReceivedEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting ether received events: ", err)
		}

		subscribeValidator, err := o.GetSubscribeValidatorEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting subscribe validator events: ", err)
		}

		unsubscribeValidator, err := o.GetUnsubscribeValidatorEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting unsubscribe validator events: ", err)
		}

		updatePoolFee, err := o.GetUpdatePoolFeeEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting update pool fee events: ", err)
		}

		poolFeeRecipient, err := o.GetPoolFeeRecipientEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting pool fee recipient events: ", err)
		}

		checkpointSlotSize, err := o.GetCheckpointSlotSizeEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting checkpoint slot size events: ", err)
		}

		updateSubscriptionCollateral, err := o.GetUpdateSubscriptionCollateralEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting update subscription collateral events: ", err)
		}

		banValidator, err := o.GetBanValidatorEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting ban validator events: ", err)
		}

		unbanValidator, err := o.GetUnbanValidatorEvents(fullBlock.GetBlockNumber())
		if err != nil {
			log.Fatal("failed getting unban validator events: ", err)
		}

		// Not all events are fetched as they are not needed
		events := &Events{
			EtherReceived:      etherReceived,
			SubscribeValidator: subscribeValidator,
			//ClaimRewards: claimRewards,
			//SetRewardRecipient: setRewardRecipient,
			UnsubscribeValidator: unsubscribeValidator,
			//InitSmoothingPool: initSmoothingPool,
			UpdatePoolFee:                updatePoolFee,
			PoolFeeRecipient:             poolFeeRecipient,
			CheckpointSlotSize:           checkpointSlotSize,
			UpdateSubscriptionCollateral: updateSubscriptionCollateral,
			//SubmitReport: submitReport,
			//ReportConsolidated: reportConsolidated,
			//UpdateQuorum: updateQuorum,
			//AddOracleMember: addOracleMember,
			//RemoveOracleMember: removeOracleMember,
			//TransferGovernance: transferGovernance,
			//AcceptGovernance: acceptGovernance,
			BanValidator:   banValidator,
			UnbanValidator: unbanValidator,
		}

		// Add the events to the block
		fullBlock.SetEvents(events)

		// If we have subscriptions or unsubscriptions, we need the state of that validator(s) at the current slot
		validatorsSubs := make([]*v1.Validator, 0)
		validatorsUnsubs := make([]*v1.Validator, 0)

		for _, sub := range fullBlock.Events.SubscribeValidator {
			validatorSub, err := o.GetSingleValidator(phase0.ValidatorIndex(sub.ValidatorID), currentSlotStr)
			if err != nil {
				log.Fatal("could not get validator subscriptions: ", err)
			}
			validatorsSubs = append(validatorsSubs, validatorSub)
		}

		for _, unsub := range fullBlock.Events.UnsubscribeValidator {
			validatorsUnsub, err := o.GetSingleValidator(phase0.ValidatorIndex(unsub.ValidatorID), currentSlotStr)
			if err != nil {
				log.Fatal("could not get validator unsubscriptions: ", err)
			}
			validatorsUnsubs = append(validatorsUnsubs, validatorsUnsub)
		}

		fullBlock.ValidatorsSubs = validatorsSubs
		fullBlock.ValidatorsUnsubs = validatorsUnsubs

		// Check if the proposal is from a subscribed validator
		isFromSubscriber := oracle.isSubscribed(fullBlock.GetProposerIndexUint64())

		// Check if the reward was sent to the pool
		isPoolRewarded := fullBlock.isAddressRewarded(o.PoolAddress)

		// This calculation is expensive, do it only if the reward went to the pool or
		// if the block is from a subscribed validator.
		if fetchAll || (isFromSubscriber || isPoolRewarded) {
			header, receipts, err := o.GetExecHeaderAndReceipts(fullBlock.GetBlockNumberBigInt(), fullBlock.GetBlockTransactions())
			if err != nil {
				log.Fatal("failed getting header and receipts: ", err)
			}
			fullBlock.SetHeaderAndReceipts(header, receipts)
		}
	}

	return fullBlock
}

// TODO: This function is not wrapped with retries
// Given a block, returns the slot where that block was proposed
func (onchain *Onchain) GetSlotByBlock(deployedBlock *big.Int, genesisTime uint64) (uint64, error) {
	// Genesis can be fetched with
	/*

		genesis, err := onchain.ConsensusClient.Genesis(context.Background())
		if err != nil {
			log.Fatal("Could not get genesis: " + err.Error())
		}

		genesisTime := uint64(genesis.GenesisTime.Unix())
	*/

	// Get that block from the execution layer
	block, err := onchain.ExecutionClient.HeaderByNumber(context.Background(), deployedBlock)
	if err != nil {
		return 0, errors.Wrap(err, "could not get block by number: "+deployedBlock.String())
	}

	// Calculate the corresponding slot given the block time and genesis time
	blockTime := block.Time
	SecondsInSlot := uint64(12)
	slot := (blockTime - genesisTime) / SecondsInSlot

	// Now we get the info at that slot from the consensus client
	blockAtSlot, err := onchain.GetConsensusBlockAtSlot(slot)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("could not get block at slot:%d", slot))
	}

	if blockAtSlot == nil {
		return 0, errors.New(fmt.Sprintf("block at slot:%d is nil, probably a missed block", slot))
	}

	fullBlock := FullBlock{ConsensusBlock: blockAtSlot}

	// And assert that indeed it matches
	if fullBlock.GetBlockNumber() != deployedBlock.Uint64() {
		return 0, errors.New(fmt.Sprintf("Block number from consensus and execution client dont match: %d != %d",
			fullBlock.GetBlockNumber(), deployedBlock.Uint64()))
	}

	return fullBlock.GetSlotUint64(), nil
}

func (onchain *Onchain) GetConfigFromContract(
	cliCfg *config.CliConfig) *Config {

	chainId, err := onchain.ExecutionClient.ChainID(context.Background())
	if err != nil {
		log.Fatal("Could not get chainid: " + err.Error())
	}

	depositContract, err := onchain.ConsensusClient.DepositContract(context.Background(), &api.DepositContractOpts{})
	if err != nil {
		log.Fatal("Could not get deposit contract: " + err.Error())
	}

	if depositContract.Data.ChainID != uint64(chainId.Int64()) {
		log.Fatal("ChainID from consensus and execution client dont match: ",
			depositContract.Data.ChainID, " != ", chainId.Int64())
	}

	network := ""
	if depositContract.Data.ChainID == MainnetChainId {
		network = "mainnet"
	} else if depositContract.Data.ChainID == GoerliChainId {
		network = "goerli"
	} else if depositContract.Data.ChainID == HoleskyChainId {
		network = "holesky"
	} else {
		log.Fatal("ChainID not supported: ", depositContract.Data.ChainID)
	}

	genesis, err := onchain.ConsensusClient.Genesis(context.Background(), &api.GenesisOpts{})
	if err != nil {
		log.Fatal("Could not get genesis: " + err.Error())
	}

	genesisTime := uint64(genesis.Data.GenesisTime.Unix())

	log.Info("Configured smoothing pool address: ", cliCfg.PoolAddress, " in network: ", network)

	balance, err := onchain.GetPoolEthBalance(nil)
	if err != nil {
		log.Fatal("Could not get pool address balance: " + err.Error())
	}
	log.Info("Pool address balance: ", utils.WeiToEther(balance), " Eth")

	// We get the block when the contract was deployed from the contract itself
	deployedBlock, err := onchain.GetContractDeploymentBlock()
	if err != nil {
		log.Fatal("Could not get contract deployment block: " + err.Error())
	}
	log.Info("[Loaded from contract] Contract deployed at block: ", deployedBlock)

	// Get that block from the execution layer
	block, err := onchain.BlockByNumber(deployedBlock)
	if err != nil {
		log.Fatal("Could not get block by number: " + err.Error())
	}

	// Calculate the corresponding slot given the block time and genesis time
	blockTime := block.Time()
	SecondsInSlot := uint64(12)
	deployedSlot := (blockTime - genesisTime) / SecondsInSlot

	// Now we get the info at that slot from the consensus client
	blockAtSlot, err := onchain.GetConsensusBlockAtSlot(deployedSlot)
	if err != nil {
		log.Fatal("Could not get block at slot: " + err.Error())
	}

	fullBlock := FullBlock{ConsensusBlock: blockAtSlot}

	// And assert that indeed it matches
	if fullBlock.GetBlockNumber() != deployedBlock.Uint64() {
		log.Fatal("Block number from consensus and execution client dont match: ",
			fullBlock.GetBlockNumber(), " != ", deployedBlock.Uint64())
	}

	log.Info("[Loaded from contract] Contract deployed in slot: ", deployedSlot)

	checkPointSizeInSlots, err := onchain.GetSlotCheckpointSize()
	if err != nil {
		log.Fatal("Could not get slot checkpoint size: " + err.Error())
	}
	log.Info("[Loaded from contract] Checkpoints will be created every ", checkPointSizeInSlots, " slots (", utils.SlotsToTime(checkPointSizeInSlots, constants.SecondsInSlot), ")")

	poolFeesPercentTwoDecimals, err := onchain.GetPoolFee()
	if err != nil {
		log.Fatal("Could not get pool fee: " + err.Error())
	}
	log.Info("[Loaded from contract] Pool fees percent: ", float64(poolFeesPercentTwoDecimals.Uint64())/100, "% (raw value: ", poolFeesPercentTwoDecimals, ")")

	poolFeesAddress, err := onchain.GetPoolFeeAddress()
	if err != nil {
		log.Fatal("Could not get pool fee address: " + err.Error())
	}
	log.Info("[Loaded from contract] Pool fees address: ", poolFeesAddress, " (ensure you control its private key)")

	ethCollateralInWei, err := onchain.GetContractCollateral()
	if err != nil {
		log.Fatal("Could not get contract collateral: " + err.Error())
	}
	log.Info("[Loaded from contract] Required collateral to join the pool: ",
		ethCollateralInWei, " wei (", utils.WeiToEther(ethCollateralInWei), " Eth)")

	if cliCfg.DryRun {
		log.Warn("The pool contract WILL NOT be updated, running in dry-run mode")
	} else {
		log.Warn("The pool contract WILL BE updated, running in updater mode")
	}

	whitelistedAddresses, err := onchain.GetAllOracleMembers()
	if err != nil {
		log.Fatal("Could not get whitelisted addresses: " + err.Error())
	}

	log.Info("[Loaded from contract] Whitelisted addresses: ", whitelistedAddresses)

	quorum, err := onchain.GetQuorum()
	if err != nil {
		log.Fatal("Could not get quorum: " + err.Error())
	}
	log.Info("[Loaded from contract] Quorum: ", quorum, "/", len(whitelistedAddresses))

	conf := &Config{
		ConsensusEndpoint:        cliCfg.ConsensusEndpoint,
		ExecutionEndpoint:        cliCfg.ExecutionEndpoint,
		Network:                  network,
		PoolAddress:              cliCfg.PoolAddress,
		DeployedSlot:             deployedSlot,
		DeployedBlock:            deployedBlock.Uint64(),
		CheckPointSizeInSlots:    checkPointSizeInSlots,
		PoolFeesPercentOver10000: int(poolFeesPercentTwoDecimals.Uint64()),
		PoolFeesAddress:          poolFeesAddress,
		CollateralInWei:          ethCollateralInWei,
		DryRun:                   cliCfg.DryRun,
		NumRetries:               cliCfg.NumRetries,
		UpdaterKeyPass:           cliCfg.UpdaterKeyPass,
		UpdaterKeyFile:           cliCfg.UpdaterKeyFile,
	}

	return conf
}

// Wrappers to fetch every event with the retrial logic
func (o *Onchain) GetEtherReceivedEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractEtherReceived, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractEtherReceivedIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterEtherReceived(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetEtherReceivedEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get EtherReceived events")
	}

	var events []*contract.ContractEtherReceived
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close EtherReceived iterator")
	}
	return events, nil
}

func (o *Onchain) GetSubscribeValidatorEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractSubscribeValidator, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractSubscribeValidatorIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterSubscribeValidator(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetSubscribeValidatorEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get SubscribeValidator events")
	}

	var events []*contract.ContractSubscribeValidator
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close SubscribeValidator iterator")
	}
	return events, nil
}

func (o *Onchain) GetClaimRewardsEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractClaimRewards, error) {

	var events []*contract.ContractClaimRewards
	log.Fatal("Not implemented: GetClaimRewardsEvents is not implemented")

	return events, nil
}

func (o *Onchain) GetSetRewardRecipientEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractSetRewardRecipient, error) {

	var events []*contract.ContractSetRewardRecipient
	log.Fatal("Not implemented: GetSetRewardsRecipientEvents it not implemented")
	return events, nil
}

func (o *Onchain) GetUnsubscribeValidatorEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUnsubscribeValidator, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUnsubscribeValidatorIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUnsubscribeValidator(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetUnsubscribeValidatorEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get UnsubscribeValidator events")
	}

	var events []*contract.ContractUnsubscribeValidator
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close UnsubscribeValidator iterator")
	}
	return events, nil
}

func (o *Onchain) GetBanValidatorEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractBanValidator, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractBanValidatorIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterBanValidator(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetBanValidatorEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get BanValidator events")
	}

	var events []*contract.ContractBanValidator
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close BanValidator iterator")
	}
	return events, nil
}

func (o *Onchain) GetUnbanValidatorEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUnbanValidator, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUnbanValidatorIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUnbanValidator(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetUnbanValidatorEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get UnbanValidator events")
	}

	var events []*contract.ContractUnbanValidator
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close UnbanValidator iterator")
	}
	return events, nil
}

func (o *Onchain) GetInitSmoothingPoolEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractInitSmoothingPool, error) {

	var events []*contract.ContractInitSmoothingPool
	log.Fatal("Not implemented: GetInitSmoothingPoolEvents it not implemented")
	return events, nil
}

func (o *Onchain) GetUpdatePoolFeeEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUpdatePoolFee, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUpdatePoolFeeIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUpdatePoolFee(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetUpdatePoolFeeEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get UpdatePoolFeeEvents events")
	}

	var events []*contract.ContractUpdatePoolFee
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close ContractUpdatePoolFee iterator")
	}
	return events, nil
}
func (o *Onchain) GetPoolFeeRecipientEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUpdatePoolFeeRecipient, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUpdatePoolFeeRecipientIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUpdatePoolFeeRecipient(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetPoolFeeRecipientEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get PoolFeeRecipient events")
	}

	var events []*contract.ContractUpdatePoolFeeRecipient
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close ContractUpdatePoolFeeRecipient iterator")
	}
	return events, nil
}
func (o *Onchain) GetCheckpointSlotSizeEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUpdateCheckpointSlotSize, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUpdateCheckpointSlotSizeIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUpdateCheckpointSlotSize(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetCheckpointSlotSizeEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get CheckpointSlotSize events")
	}

	var events []*contract.ContractUpdateCheckpointSlotSize
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close ContractUpdateCheckpointSlotSize iterator")
	}
	return events, nil
}
func (o *Onchain) GetUpdateSubscriptionCollateralEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUpdateSubscriptionCollateral, error) {

	startBlock := uint64(blockNumber)
	endBlock := uint64(blockNumber)

	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: startBlock, End: &endBlock}

	var err error
	var itr *contract.ContractUpdateSubscriptionCollateralIterator

	err = retry.Do(func() error {
		itr, err = o.Contract.FilterUpdateSubscriptionCollateral(filterOpts)
		if err != nil {
			log.Warn("Failed attempt GetUpdateSubscriptionCollateralEvents for block ", strconv.FormatUint(blockNumber, 10), ": ", err.Error(), " Retrying...")
			return err
		}
		return nil
	}, o.GetRetryOpts(opts)...)

	if err != nil {
		return nil, errors.Wrap(err, "could not get UpdateSubscriptionCollateral events")
	}

	var events []*contract.ContractUpdateSubscriptionCollateral
	for itr.Next() {
		events = append(events, itr.Event)
	}
	err = itr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close ContractUpdateSubscriptionCollateral iterator")
	}
	return events, nil
}
func (o *Onchain) GetSubmitReportEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractSubmitReport, error) {

	var events []*contract.ContractSubmitReport
	log.Fatal("Not implemented: GetSubmitReportEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetReportConsolidatedEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractReportConsolidated, error) {

	var events []*contract.ContractReportConsolidated
	log.Fatal("Not implemented: GetReportConsolidatedEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetUpdateQuorumEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractUpdateQuorum, error) {

	var events []*contract.ContractUpdateQuorum
	log.Fatal("Not implemented: GetUpdateQuorumEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetAddOracleMemberEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractAddOracleMember, error) {

	var events []*contract.ContractAddOracleMember
	log.Fatal("Not implemented: GetAddOracleMemberEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetRemoveOracleMemberEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractRemoveOracleMember, error) {

	var events []*contract.ContractRemoveOracleMember
	log.Fatal("Not implemented: GetRemoveOracleMemberEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetTransferGovernanceEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractTransferGovernance, error) {

	var events []*contract.ContractTransferGovernance
	log.Fatal("Not implemented: GetTransferGovernanceEvents it not implemented")
	return events, nil
}
func (o *Onchain) GetAcceptGovernanceEvents(
	blockNumber uint64,
	opts ...retry.Option) ([]*contract.ContractAcceptGovernance, error) {

	var events []*contract.ContractAcceptGovernance
	log.Fatal("Not implemented: GetAcceptGovernanceEvents it not implemented")
	return events, nil
}

func (o *Onchain) GetClaimedPerWithdrawalAddress(addresses []string, finalizedBlock *big.Int, opts ...retry.Option) (map[string]*big.Int, error) {
	claimedMap := make(map[string]*big.Int)

	for _, address := range addresses {
		// Important to use finalized blocks. Otherwise our local oracle view and onchain view may differ
		claimed, err := o.GetContractClaimedBalance(address, finalizedBlock, opts...)
		if err != nil {
			return nil, errors.Wrap(err, "could not get claimed balance")
		}
		_, found := claimedMap[address]
		if found {
			log.Fatal("Duplicate deposit address found: ", address)
		}
		claimedMap[address] = claimed
	}
	return claimedMap, nil
}

func (o *Onchain) UpdateContractMerkleRoot(slot uint64, newMerkleRoot string) error {

	// Support both 0x prefixed and non prefixed merkle roots
	if strings.HasPrefix(newMerkleRoot, "0x") {
		newMerkleRoot = newMerkleRoot[2:]
	}

	// Parse merkle root to byte array
	newMerkleRootBytes := [32]byte{}
	unboundedBytes := common.Hex2Bytes(newMerkleRoot)

	if len(unboundedBytes) != 32 {
		return errors.New(fmt.Sprintf("merkle root must be 32 bytes: %s", newMerkleRoot))
	}
	copy(newMerkleRootBytes[:], common.Hex2Bytes(newMerkleRoot))

	// Sanity check to ensure the converted tree matches the original
	if hex.EncodeToString(newMerkleRootBytes[:]) != newMerkleRoot {
		return errors.New(fmt.Sprintf("merkle trees dont match, expected: %s", newMerkleRoot))
	}

	publicKey := o.updaterKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Info("Preparing tx from address: ", fromAddress.Hex())
	nonce, err := o.ExecutionClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return errors.New(fmt.Sprintf("could not get pending nonce: %s", err))
	}

	// Unused, leaving for reference. We rely on automatic gas estimation, see below (nil values)
	//gasTipCap, err := o.ExecutionClient.SuggestGasTipCap(context.Background())
	//if err != nil {
	//	return errors.New(fmt.Sprintf("could not get gas tip cap suggestion: %s", err))
	//}
	//_ = gasTipCap

	chaindId, err := o.ExecutionClient.NetworkID(context.Background())
	if err != nil {
		return errors.New(fmt.Sprintf("could not get chaind: %s", err))
	}

	auth, err := bind.NewKeyedTransactorWithChainID(o.updaterKey, chaindId)
	if err != nil {
		return errors.New(fmt.Sprintf("could not create NewKeyedTransactorWithChainID: %s", err))
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// Important that the value is 0. Otherwise we would be sending Eth
	// and thats not neccessary.
	auth.Value = big.NewInt(0)

	// nil prices automatically estimate prices
	auth.GasPrice = nil
	auth.GasFeeCap = nil
	auth.GasTipCap = nil
	auth.NoSend = false
	auth.Context = context.Background()

	address := common.HexToAddress(o.PoolAddress)

	instance, err := contract.NewContract(address, o.ExecutionClient)
	if err != nil {
		return errors.Wrap(err, "could not create contract instance")
	}

	// Create a tx calling the update rewards root function with the new merkle root
	tx, err := instance.SubmitReport(auth, slot, newMerkleRootBytes)
	if err != nil {
		return errors.Wrap(err, "could not create tx to call SubmitReport")
	}

	log.WithFields(log.Fields{
		"TxHash":        tx.Hash().Hex(),
		"NewMerkleRoot": newMerkleRoot,
		"Slot":          slot,
	}).Info("Tx sent to Ethereum updating rewards merkle root, wait to be validated")

	// Leave 60 minutes for the tx to be validated
	deadline := time.Now().Add(60 * time.Minute)
	ctx, cancelCtx := context.WithDeadline(context.Background(), deadline)
	defer cancelCtx()

	// It stops waiting when the context is canceled.
	receipt, err := bind.WaitMined(ctx, o.ExecutionClient, tx)
	if ctx.Err() != nil {
		return errors.Wrap(err,
			fmt.Sprint("timeout expired waiting for tx to be validated txHash: ",
				tx.Hash().Hex()))
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.Wrap(err,
			fmt.Sprintf("tx failed err: %d hash: %s", receipt.Status, tx.Hash().Hex()))
	}

	// Tx was sent and validated correctly, print receipt info
	log.WithFields(log.Fields{
		"Status":            receipt.Status,
		"CumulativeGasUsed": receipt.CumulativeGasUsed,
		"TxHash":            receipt.TxHash,
		"GasUsed":           receipt.GasUsed,
		"BlockHash":         receipt.BlockHash.Hex(),
		"BlockNumber":       receipt.BlockNumber,
	}).Info("Tx: ", tx.Hash().Hex(), " was validated ok. Receipt info:")

	return nil
}

// Loads all validator from the beacon chain into the oracle, must be called periodically
func (o *Onchain) RefreshBeaconValidators() {
	log.Info("Loading existing validators from the beacon chain")
	vals, err := o.GetFinalizedValidators()
	if err != nil {
		log.Fatal("Could not get validators: ", err)
	}
	o.validators = vals
	if len(vals) != 0 {
		log.WithFields(log.Fields{
			"TotalValidators":       len(vals),
			"LastIndex":             vals[phase0.ValidatorIndex(len(vals)-1)].Index,
			"ActivationSlotLastVal": utils.GetActivationSlotOfLatestProcessedValidator(vals),
		}).Info("Done loading beacon chain validators")
	} else {
		log.Fatal("No validators were loaded from the beacon chain")
	}
}

func (o *Onchain) Validators() map[phase0.ValidatorIndex]*v1.Validator {
	return o.validators
}

func (o *Onchain) GetRetryOpts(opts []retry.Option) []retry.Option {
	// Default retry options. This specifies what to do when a call to the
	// consensus or execution client fails. Default is to retry x times (see config)
	// with a 15 seconds delay and the default backoff strategy (see avas/retry-go)
	// Note that in some cases we might want to avoid retrying at all, for example
	// when serving data to an api, we may want to just fail fast and return an error
	// If this function is called with retry options, we use those instead as a way
	// to override the default retry options
	if len(opts) == 0 {
		return []retry.Option{
			retry.Attempts(uint(o.NumRetries)),
			retry.Delay(15 * time.Second),
		}
	} else {
		return opts
	}
}
