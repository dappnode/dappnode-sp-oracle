package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/avast/retry-go/v4"
	"github.com/pkg/errors"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/dappnode/mev-sp-oracle/contract"
	"github.com/dappnode/mev-sp-oracle/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/sirupsen/logrus"
)

// Default path of persisted state
var StateFolder = "oracle-data"
var StateJsonName = "state.json"

type GetSetOfValidatorsFunc func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error)
type GetOwnerToPodFunc func(address common.Address, opts ...retry.Option) (common.Address, error)
type GetValidatorRestakedEvents func(podAddress common.Address, opts ...retry.Option) ([]*big.Int, error)
type Oracle struct {
	cfg                        *Config
	state                      *OracleState
	mutex                      sync.RWMutex
	getSetOfValidators         GetSetOfValidatorsFunc
	getOwnerToPod              GetOwnerToPodFunc
	getValidatorRestakedEvents GetValidatorRestakedEvents
}

// Fork 1 changes two things:
// - minor fix in rewards calculation (some wei rouding)
// - exited and slahed validators no longer get fees
var SlotFork1 = map[string]uint64{
	"mainnet": uint64(10188220),
	"holesky": uint64(2720632),
}

func NewOracle(cfg *Config) *Oracle {
	state := &OracleState{
		StateHash:            "",
		LatestProcessedBlock: 0,
		LatestProcessedSlot:  cfg.DeployedSlot - 1,
		NextSlotToProcess:    cfg.DeployedSlot,
		PoolAccumulatedFees:  big.NewInt(0),
		Validators:           make(map[uint64]*ValidatorInfo, 0),
		CommitedStates:       make(map[uint64]*OnchainState, 0),
		SubscriptionEvents:   make([]*contract.ContractSubscribeValidator, 0),
		UnsubscriptionEvents: make([]*contract.ContractUnsubscribeValidator, 0),
		EtherReceivedEvents:  make([]*contract.ContractEtherReceived, 0),
		Donations:            make([]*contract.ContractEtherReceived, 0),
		ProposedBlocks:       make([]SummarizedBlock, 0),
		MissedBlocks:         make([]SummarizedBlock, 0),
		WrongFeeBlocks:       make([]SummarizedBlock, 0),

		// Config
		PoolFeesPercentOver10000: cfg.PoolFeesPercentOver10000,
		PoolAddress:              cfg.PoolAddress,
		Network:                  cfg.Network,
		PoolFeesAddress:          cfg.PoolFeesAddress,
		CheckPointSizeInSlots:    cfg.CheckPointSizeInSlots,
		DeployedBlock:            cfg.DeployedBlock,
		DeployedSlot:             cfg.DeployedSlot,
		CollateralInWei:          cfg.CollateralInWei,
	}

	oracle := &Oracle{
		cfg:                cfg,
		state:              state,
		getSetOfValidators: nil,
		getOwnerToPod:      nil,
	}

	return oracle
}

func (or *Oracle) SetGetSetOfValidatorsFunc(oc GetSetOfValidatorsFunc) {
	or.getSetOfValidators = oc
}

func (or *Oracle) SetOwnerToPodFunc(oc GetOwnerToPodFunc) {
	or.getOwnerToPod = oc
}

func (or *Oracle) SetValidatorRestakedEventsFunc(oc GetValidatorRestakedEvents) {
	or.getValidatorRestakedEvents = oc
}

// Returns the state of the oracle, containing all the information about the
// validatores, with their state, balances, etc
func (or *Oracle) State() *OracleState {
	or.mutex.RLock()
	defer or.mutex.RUnlock()
	return or.state
}

// Returns wether a checkpoint has been reached or not. A checkpoint is reached
// when CheckPointSizeInSlots have passed from the last checkpoint
func (or *Oracle) IsCheckpoint() (bool, error) {
	or.mutex.RLock()
	defer or.mutex.RUnlock()
	latestProcSlot := or.State().LatestProcessedSlot

	if latestProcSlot == 0 {
		return false, errors.New(
			fmt.Sprintf("cannot determine if checkpoint has been reached, no slots have been processed yet. latest=%d",
				latestProcSlot))
	}

	if or.cfg.DeployedSlot > latestProcSlot {
		return false, errors.New(fmt.Sprintf("deployed slot can't be greater than latest slot. deployed=%d, latest=%d",
			or.cfg.DeployedSlot, latestProcSlot))
	}

	if (latestProcSlot-or.cfg.DeployedSlot)%or.cfg.CheckPointSizeInSlots == 0 {
		return true, nil
	}
	return false, nil
}

// Returns the state of the oracle, recalculating the hash of the state for
// verification purposes
func (or *Oracle) StateWithHash() (*OracleState, error) {
	or.mutex.Lock()

	// Update hash
	err := or.hashStateLockFree()
	if err != nil {
		return nil, errors.Wrap(err, "error hashing the oracle state")
	}
	or.mutex.Unlock()
	return or.State(), nil
}

// Given a previous or.state, this function applies the new block to it, updating the or.state
// with the new subscriptions, unsubscriptions, donations, and rewards to the pool, updating
// the balance of all participating validators. Returns the slot that was processed and if there
// was an error.
func (or *Oracle) AdvanceStateToNextSlot(fullBlock *FullBlock) (uint64, error) {

	or.mutex.Lock()
	defer or.mutex.Unlock()

	// Ensure the slot to process is the last +1
	if or.state.NextSlotToProcess != (or.state.LatestProcessedSlot + 1) {
		return 0, errors.New(fmt.Sprint("Next slot to process is not the last processed slot + 1",
			or.state.NextSlotToProcess, " ", or.state.LatestProcessedSlot))
	}

	// Ensure the block to process matches the expected duty
	if or.state.NextSlotToProcess != uint64(fullBlock.ConsensusDuty.Slot) {
		return 0, errors.New(fmt.Sprint("Next slot to process is not the same as the block slot",
			or.state.NextSlotToProcess, " ", fullBlock.ConsensusDuty.Slot))
	}

	// Some misc validations
	err := or.validateFullBlockConfig(fullBlock, or.cfg)
	if err != nil {
		return 0, errors.Wrap(err, "Error validating full block config")
	}

	// Full block is too heavy to be stored in the state, so we summarize it
	summarizedBlock := fullBlock.SummarizedBlock(or, or.cfg.PoolAddress)

	// Ensure the block we process is the expected one
	if or.state.NextSlotToProcess != summarizedBlock.Slot {
		return 0, errors.New(fmt.Sprint("Next slot to process is not the same as the block slot",
			or.state.NextSlotToProcess, " ", summarizedBlock.Slot))
	}

	// Get donations to the pool in this block
	blockDonations := fullBlock.GetDonations(or.cfg.PoolAddress)

	// Store all events raw for trazability
	or.state.SubscriptionEvents = append(or.state.SubscriptionEvents, fullBlock.Events.SubscribeValidator...)
	or.state.UnsubscriptionEvents = append(or.state.UnsubscriptionEvents, fullBlock.Events.UnsubscribeValidator...)
	or.state.EtherReceivedEvents = append(or.state.EtherReceivedEvents, fullBlock.Events.EtherReceived...)

	// Handle subscriptions first thing
	or.handleManualSubscriptions(fullBlock.Events.SubscribeValidator, fullBlock.ValidatorsSubs)

	// If the validator was subscribed and missed proposed the block in this slot
	if summarizedBlock.BlockType == MissedProposal && or.isSubscribed(summarizedBlock.ValidatorIndex) {
		or.handleMissedBlock(summarizedBlock)
	}

	// If we have a successful block proposal BUT the validator has BLS keys, we cant auto subscribe it
	if summarizedBlock.BlockType == OkPoolProposalBlsKeys {
		or.handleBlsCorrectBlockProposal(summarizedBlock)
	}

	// If fee recipient matches the pool, we distribute the rewards and upate
	// the validator state. Automatic subscriptions are considered here
	if summarizedBlock.BlockType == OkPoolProposal {
		or.handleCorrectBlockProposal(summarizedBlock)
	}

	// If the validator was subscribed but the fee recipient was wrong we ban the validator
	if summarizedBlock.BlockType == WrongFeeRecipient && or.isSubscribed(summarizedBlock.ValidatorIndex) {
		or.handleBanValidator(summarizedBlock)
	}

	// Handle unsubscriptions the last thing after distributing rewards
	or.handleManualUnsubscriptions(fullBlock.Events.UnsubscribeValidator, fullBlock.ValidatorsUnsubs)

	// Handle the donations from this block
	or.handleDonations(blockDonations)

	// Manual bans/unbans should always be the last thing to be processed in each block, since
	// we want to ensure they persist to the next block
	// Handle manual bans
	or.handleManualBans(fullBlock.Events.BanValidator)

	// Handle manual unbans
	or.handleManualUnbans(fullBlock.Events.UnbanValidator)

	// Handle validator cleanup: redisitribute the pending rewards of validators subscribed to the pool
	// that are not in the beacon chain anymore (exited/slashed). We dont run this on every slot because
	// its expensive. Runs every 4 hours.
	if or.state.NextSlotToProcess%uint64(1200) == 0 {
		err = or.ValidatorCleanup(or.state.NextSlotToProcess)
		if err != nil {
			return 0, errors.Wrap(err, "could not cleanup validators")
		}
	}

	processedSlot := or.state.NextSlotToProcess
	or.state.LatestProcessedSlot = processedSlot
	or.state.NextSlotToProcess++
	if summarizedBlock.BlockType != MissedProposal {
		or.state.LatestProcessedBlock = summarizedBlock.Block
	}
	return processedSlot, nil
}

// Unsubscribes validators that are not active. Shares their pending rewards to the pool
func (or *Oracle) ValidatorCleanup(slot uint64) error {

	// Only cleanup if we're past the cleanup slot fork
	if slot >= SlotFork1[or.cfg.Network] {

		// Extract all validator indices from the oracle state
		indices := make([]phase0.ValidatorIndex, 0)
		for idx := range or.state.Validators {
			indices = append(indices, phase0.ValidatorIndex(idx))
		}

		// if oracle isn't tracking any validator, it means that nobody ever subscribed, nothing to cleanup
		if len(indices) == 0 {
			log.Info("No validators to cleanup, state has no validators")
			return nil
		}

		// Get the latest validator information for all subscribed validators at once
		validatorInfo, err := or.getSetOfValidators(indices, strconv.FormatUint(slot, 10))
		if err != nil {
			return errors.Wrap(err, "could not get validators info")
		}

		// Iterate over all validators. If two or more validators exit or get slashed in the same slot,
		// this cleanup will eventually set both of their pending rewards to 0 and share them among the pool
		rewardsToDistribute := big.NewInt(0)
		for _, validator := range validatorInfo {
			// If a validator is subscribed but not active onchain, we have to unsubscribe it and treat it as a ban:
			// this means setting the validator rewards to 0 and sharing them among the pool
			if !validator.Status.IsActive() && or.isSubscribed(uint64(validator.Index)) {
				log.WithFields(log.Fields{
					"PendingRewardsWei":     or.state.Validators[uint64(validator.Index)].PendingRewardsWei,
					"BeaconValidatorState":  or.state.Validators[uint64(validator.Index)].ValidatorStatus,
					"ValidatorIndex":        validator.Index,
					"OracleValidatorStatus": validator.Status,
					"Slot":                  slot,
					"Network":               or.cfg.Network,
				}).Info("Cleaning up validator")
				or.advanceStateMachine(uint64(validator.Index), Unsubscribe)
				rewardsToDistribute.Add(rewardsToDistribute, or.state.Validators[uint64(validator.Index)].PendingRewardsWei)
				or.resetPendingRewards(uint64(validator.Index))
			}
		}

		// Distribute the rewards among the pool. Majority of times this will be 0
		if rewardsToDistribute.Cmp(big.NewInt(0)) != 0 {
			or.increaseAllPendingRewards(rewardsToDistribute)
		}
		log.Info("Validator cleanup done! Redistributed a total of ", rewardsToDistribute, " wei in pending among the pool in slot ", slot)
	}

	return nil
}

// We use the following events to validate that the config has not changed. Dynamic
// parameters are not supported.
// UpdatePoolFee: Indicates the cut in %*100 the pool gets
// PoolFeeRecipient: Indicates the address that receives the pool fees
// CheckpointSlotSize: Indicates the size of the checkpoint in slots
// UpdateSubscriptionCollateral: Indicates the amount of ETH required to subscribe
func (or *Oracle) validateFullBlockConfig(fullBlock *FullBlock, config *Config) error {
	if len(fullBlock.Events.UpdatePoolFee) > 1 ||
		len(fullBlock.Events.PoolFeeRecipient) > 1 ||
		len(fullBlock.Events.CheckpointSlotSize) > 1 ||
		len(fullBlock.Events.UpdateSubscriptionCollateral) > 1 {
		return errors.New("more than one event of the same type in the same block, weird")
	}

	if len(fullBlock.Events.UpdatePoolFee) != 0 && big.NewInt(int64(config.PoolFeesPercentOver10000)).Cmp(fullBlock.Events.UpdatePoolFee[0].NewPoolFee) != 0 {
		return errors.New(fmt.Sprintf("pool fee has changed. config: %d, block: %d",
			config.PoolFeesPercentOver10000, fullBlock.Events.UpdatePoolFee[0].NewPoolFee))
	}

	if len(fullBlock.Events.PoolFeeRecipient) != 0 && !utils.Equals(config.PoolFeesAddress, fullBlock.Events.PoolFeeRecipient[0].NewPoolFeeRecipient.String()) {
		return errors.New(fmt.Sprintf("pool fee recipient has changed. config: %s, block: %s",
			config.PoolFeesAddress, fullBlock.Events.PoolFeeRecipient[0].NewPoolFeeRecipient.String()))
	}

	if len(fullBlock.Events.CheckpointSlotSize) != 0 && config.CheckPointSizeInSlots != fullBlock.Events.CheckpointSlotSize[0].NewCheckpointSlotSize {
		return errors.New(fmt.Sprintf("checkpoint size has changed. config: %d, block: %d",
			config.CheckPointSizeInSlots, fullBlock.Events.CheckpointSlotSize[0].NewCheckpointSlotSize))
	}

	if len(fullBlock.Events.UpdateSubscriptionCollateral) != 0 && config.CollateralInWei.Cmp(fullBlock.Events.UpdateSubscriptionCollateral[0].NewSubscriptionCollateral) != 0 {
		return errors.New(fmt.Sprintf("subscription collateral has changed. config: %d, block: %d",
			config.CollateralInWei, fullBlock.Events.UpdateSubscriptionCollateral[0].NewSubscriptionCollateral))
	}

	return nil
}

// Persist the state of the oracle to a JSON file. By default its stored
// as state.json but if saveSlot is true, it will store two copies,
// one updating the existing state.json and other as state_<slot>.json.
// The later is to be used mainly for debugging and recovery purposes.
func (or *Oracle) SaveToJson(saveSlot bool) error {
	// Not just read lock since we change the hash, minor thing
	// but it cant be just a read mutex

	or.mutex.Lock()

	log.Info("Saving oracle state to JSON file")

	err := or.hashStateLockFree()
	if err != nil {
		return errors.Wrap(err, "error hashing the oracle state")
	}

	jsonData, err := json.MarshalIndent(or.state, "", " ")
	if err != nil {
		return errors.Wrap(err, "could not marshal state to JSON")
	}
	or.mutex.Unlock()

	path := filepath.Join(StateFolder, StateJsonName)
	err = os.MkdirAll(StateFolder, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "could not create folder")
	}

	log.Trace("Saving state to file:", fmt.Sprintf("%s", jsonData))

	err = ioutil.WriteFile(path, jsonData, 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	log.WithFields(log.Fields{
		"LatestProcessedSlot":  or.state.LatestProcessedSlot,
		"LatestProcessedBlock": or.state.LatestProcessedBlock,
		"NextSlotToProcess":    or.state.NextSlotToProcess,
		"TotalValidators":      len(or.state.Validators),
		"Network":              or.state.Network,
		"PoolAddress":          or.state.PoolAddress,
		"Path":                 path,
		"Hash":                 or.state.StateHash,
	}).Info("Saved state to file")

	// If saveSlot is true, save a copy of the state with the slot number in the file
	if saveSlot {
		filename := fmt.Sprintf("state_%d.json", or.State().LatestProcessedSlot)
		path = filepath.Join(StateFolder, filename)

		log.WithFields(log.Fields{
			"LatestProcessedSlot": or.state.LatestProcessedSlot,
			"FileName":            filename,
		}).Info("Storing also a copy of the state")

		err = ioutil.WriteFile(path, jsonData, 0644)
		if err != nil {
			return errors.Wrap(err, "could not write file")
		}
	}

	return nil
}

// Loads the oracle state from a human readable json file. Multiple
// check are performed to ensure the state is valid such as checking
// the hash of the state and ensuring the configuation has not changed
func (or *Oracle) LoadFromJson() (bool, error) {
	path := filepath.Join(StateFolder, StateJsonName)
	has, err := or.LoadFromPath(path)
	return has, err
}

func (or *Oracle) LoadFromPath(path string) (bool, error) {
	or.mutex.Lock()
	defer or.mutex.Unlock()

	log.Info("Loading oracle state from json file: ", path)

	jsonFile, err := os.Open(path)
	defer jsonFile.Close()

	// Dont error if the file wasnt found, just return not found
	if err != nil {
		return false, nil
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return false, errors.Wrap(err, "could not read json file")
	}

	found, err := or.LoadFromBytes(byteValue)

	return found, err
}

func (or *Oracle) LoadFromBytes(rawBytes []byte) (bool, error) {
	var state OracleState

	err := json.Unmarshal(rawBytes, &state)
	if err != nil {
		return false, errors.Wrap(err, "could not unmarshal json file")
	}

	// Store the hash we recovered from the file
	recoveredHash := state.StateHash

	// Reset the hash since we want to hash the content without the hash
	state.StateHash = ""

	// Serialize the state without the hash
	jsonNoHash, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return false, errors.Wrap(err, "could not marshal state without hash")
	}

	log.Trace("Loaded state from file: ", fmt.Sprintf("%s", jsonNoHash))

	// We calculate the hash of the state we read
	calculatedHashByte := sha256.Sum256(jsonNoHash[:])
	calculatedHashString := hexutil.Encode(calculatedHashByte[:])

	// Hashes must match
	if !utils.Equals(recoveredHash, calculatedHashString) {
		return false, errors.New(fmt.Sprintf("hash mismatch, recovered: %s, calculated: %s",
			recoveredHash, calculatedHashString))
	}

	// Sanity check to ensure the oracle config matches the loaded state
	if state.Network != or.cfg.Network {
		return false, errors.New(fmt.Sprintf("network mismatch, recovered: %s, expected: %s",
			state.Network, or.cfg.Network))
	}

	if state.PoolAddress != or.cfg.PoolAddress {
		return false, errors.New(fmt.Sprintf("pool address mismatch, recovered: %s, expected: %s",
			state.PoolAddress, or.cfg.PoolAddress))
	}

	if state.PoolFeesAddress != or.cfg.PoolFeesAddress {
		return false, errors.New(fmt.Sprintf("pool fees address mismatch, recovered: %s, expected: %s",
			state.PoolFeesAddress, or.cfg.PoolFeesAddress))
	}

	if state.PoolFeesPercentOver10000 != or.cfg.PoolFeesPercentOver10000 {
		return false, errors.New(fmt.Sprintf("pool fees percent mismatch, recovered: %d, expected: %d",
			state.PoolFeesPercentOver10000, or.cfg.PoolFeesPercentOver10000))
	}

	if state.CollateralInWei.Cmp(or.cfg.CollateralInWei) != 0 {
		return false, errors.New(fmt.Sprintf("collateral mismatch, recovered: %d, expected: %d",
			state.CollateralInWei, or.cfg.CollateralInWei))
	}

	if state.CheckPointSizeInSlots != or.cfg.CheckPointSizeInSlots {
		return false, errors.New(fmt.Sprintf("check point size mismatch, recovered: %d, expected: %d",
			state.CheckPointSizeInSlots, or.cfg.CheckPointSizeInSlots))
	}

	if state.DeployedBlock != or.cfg.DeployedBlock {
		return false, errors.New(fmt.Sprintf("deployed block mismatch, recovered: %d, expected: %d",
			state.DeployedBlock, or.cfg.DeployedBlock))
	}

	if state.DeployedSlot != or.cfg.DeployedSlot {
		return false, errors.New(fmt.Sprintf("deployed slot mismatch, recovered: %d, expected: %d",
			state.DeployedSlot, or.cfg.DeployedSlot))
	}

	or.state = &state

	mRoot, enoughData := or.getMerkleRootIfAny()
	log.WithFields(log.Fields{
		"LatestProcessedSlot":  state.LatestProcessedSlot,
		"LatestProcessedBlock": state.LatestProcessedBlock,
		"NextSlotToProcess":    state.NextSlotToProcess,
		"Network":              state.Network,
		"PoolAddress":          state.PoolAddress,
		"MerkleRoot":           mRoot,
		"EnoughData":           enoughData,
	}).Info("Loaded state from file")
	return true, nil
}

func (or *Oracle) LoadGivenState(slotCheckpoint uint64) (bool, error) {
	// Try to load the given state
	path := filepath.Join(StateFolder, fmt.Sprintf("state_%d.json", slotCheckpoint))
	has, err := or.LoadFromPath(path)
	if err != nil {
		return false, err
	}

	// If not found, attemp to load previous states up to "attempts" checkpoints before
	attempts := 3
	if !has {
		for i := 1; i < attempts; i++ {
			trySlot := slotCheckpoint - or.cfg.CheckPointSizeInSlots*uint64(i)
			log.Info("Could not find slot for checkpoint, ", slotCheckpoint, ", trying slot: ", trySlot)
			path = filepath.Join(StateFolder, fmt.Sprintf("state_%d.json", trySlot))
			has, err = or.LoadFromPath(path)
			if has {
				break
			}
		}
	}

	return has, err
}

// Takes the current state, creates a copy of it and freezes it, storing
// it in a map slot->state. It also creates a set of merkle proof for each
// withdrawal address of each validator. Each of these frozen states maps
// to a commited onchain state, represented by a merkle root.
// Returns false if there wasnt enough data to create a merkle tree
func (or *Oracle) FreezeCheckpoint() bool {
	or.mutex.Lock()
	defer or.mutex.Unlock()

	validatorsCopy := make(map[uint64]*ValidatorInfo)
	utils.DeepCopy(or.state.Validators, &validatorsCopy)

	mk := NewMerklelizer()
	withdrawalToLeaf, withdrawalToRawLeaf, tree, enoughData := mk.GenerateTreeFromState(or.state)
	if !enoughData {
		return false
	}
	merkleRootStr := hexutil.Encode(tree.Root[:])

	log.WithFields(log.Fields{
		"Slot":       or.state.LatestProcessedSlot,
		"MerkleRoot": merkleRootStr,
	}).Info("Freezing state")

	// Merkle proofs for each withdrawal address
	proofs := make(map[string][]string)
	leafs := make(map[string]RawLeaf)
	for WithdrawalAddress, rawLeaf := range withdrawalToRawLeaf {

		// Extra sanity check to make sure the withdrawal address is the same as the key
		if !utils.Equals(WithdrawalAddress, rawLeaf.WithdrawalAddress) {
			log.Fatal("withdrawal address in raw leaf doesnt match the key")
		}

		block := withdrawalToLeaf[WithdrawalAddress]
		proof, err := tree.Proof(block)

		if err != nil {
			log.Fatal("could not generate proof for block: ", err)
		}

		// Store the proofs of the withdrawal address (to be used onchain)
		proofs[WithdrawalAddress] = utils.ByteArrayToArray(proof.Siblings)

		// Store the leafs (to be used onchain)
		leafs[WithdrawalAddress] = rawLeaf
	}

	state := &OnchainState{
		Validators: validatorsCopy,
		MerkleRoot: merkleRootStr,
		Slot:       or.state.LatestProcessedSlot,
		Proofs:     proofs,
		Leafs:      leafs,
	}

	or.state.CommitedStates[state.Slot] = state
	return true
}

// Returns true and the latest commited slot if there is any commited state
// false otherwise. Note that if there are checkpoints but without enough data
// to create a tree, it will still return false
func (or *Oracle) LatestCommitedSlot() (uint64, bool) {
	or.mutex.RLock()
	defer or.mutex.RUnlock()

	if len(or.State().CommitedStates) == 0 {
		return 0, false
	}
	latestCommitedSlot := uint64(0)
	for slot, _ := range or.State().CommitedStates {
		if slot > latestCommitedSlot {
			latestCommitedSlot = slot
		}
	}
	return latestCommitedSlot, true
}

// Returns the last commited state
func (or *Oracle) LatestCommitedState() *OnchainState {
	or.mutex.RLock()
	defer or.mutex.RUnlock()

	latestCommitedSlot, atLeastOne := or.LatestCommitedSlot()

	// If we havent commited any state yet
	if !atLeastOne {
		return nil
	}

	return or.State().CommitedStates[latestCommitedSlot]
}

// Check if the oracle is in sync with a given root and slot. Its considered in sync
// when the latest commited state has the same root and slot as the onchain state
func (or *Oracle) IsOracleInSyncWithChain(onchainRoot string, onchainSlot uint64) (bool, error) {
	latestCommitedSlot, atLeastOne := or.LatestCommitedSlot()

	// If we havent commited any state yet
	if !atLeastOne {
		log.Info("Oracle has no commited states, no checkpoints have passed or there is not enough data to create a merkle tree")
		// If the onchain state is the default, we can consider in sync as the contract has not root also
		if utils.Equals(onchainRoot, DefaultRoot) {
			log.WithFields(log.Fields{
				"OnchainRoot": onchainRoot,
				"OnchainSlot": onchainSlot,
			}).Info("Oracle IS in sync with the latest onchain root, nothing was commited onchain yet")
			return true, nil
		}
		// If the onchain state is not the default, we are not in sync
		log.WithFields(log.Fields{
			"OnchainRoot": onchainRoot,
			"OnchainSlot": onchainSlot,
		}).Info("Oracle IS NOT in sync with the latest onchain root, oracle has no commited states yet")
		return false, nil
	}

	latestOracleRoot := or.State().CommitedStates[latestCommitedSlot].MerkleRoot

	if utils.Equals(onchainRoot, latestOracleRoot) && onchainSlot == latestCommitedSlot {
		log.WithFields(log.Fields{
			"OnchainRoot": onchainRoot,
			"OnchainSlot": onchainSlot,
			"OracleRoot":  latestOracleRoot,
			"OracleSlot":  latestCommitedSlot,
		}).Info("Oracle IS in sync with the latest onchain root")
		return true, nil
	}

	// If roots match but not slots or viceversa. Something is wrong
	if (utils.Equals(onchainRoot, latestOracleRoot) && onchainSlot != latestCommitedSlot) ||
		(!utils.Equals(onchainRoot, latestOracleRoot) && onchainSlot == latestCommitedSlot) {
		return false, errors.New(fmt.Sprintf("Oracle root/slot does not match the onchain root/slot. "+
			"OracleRoot: %s, OracleSlot: %d, OnchainRoot: %s, OnchainSlot: %d",
			latestOracleRoot, latestCommitedSlot, onchainRoot, onchainSlot))
	}

	log.WithFields(log.Fields{
		"OnchainRoot": onchainRoot,
		"OracleRoot":  latestOracleRoot,
		"OracleSlot":  latestCommitedSlot,
	}).Info("Oracle IS NOT in sync with the latest onchain root")
	return false, nil
}

// Ensures that our liabilities are equal to our assets where:
// - liabilities: sum of all rewards of all validators + pool fees
// - assets: sum of all donations + block rewards
// They must be equal at any point in time and this function ensures so. Note
// that there are two scenarios to prevent:
// - liabilities > assets: means we are giving more money than we are receiving which will
// result in the pool being unable to pay.
// - assets > liabilities: means less rewards are distributed, and since everything is encoded
// in the root, this means some funds will be locked forever.
func (or *Oracle) RunOnchainReconciliation(
	contractBalanceWei *big.Int,
	claimedAmountsWei map[string]*big.Int) error {

	// We calculate:
	// 1. what we owe: total pending + accumulated rewards for all vlaidators + pool fees.
	// to this we have to substract the amount that each deposit address alredy claimed
	// 2. what we have: the amount in the smart contract

	// both amount have to match at any time, asssuming we run this on finalized
	// on the same slots (finalized epochs)

	// What we owe (1/2)
	totalCumulativeRewards := big.NewInt(0)
	for _, val := range or.state.Validators {
		totalCumulativeRewards.Add(totalCumulativeRewards, val.AccumulatedRewardsWei)
		totalCumulativeRewards.Add(totalCumulativeRewards, val.PendingRewardsWei)
	}
	totalCumulativeRewards.Add(totalCumulativeRewards, or.state.PoolAccumulatedFees)

	log.Info("[Reconciliation] Total amount of accumulated + pending rewards: ", totalCumulativeRewards)

	// What we owe (2/2)
	totalAlreadyClaimed := big.NewInt(0)
	for _, claimed := range claimedAmountsWei {
		totalAlreadyClaimed.Add(totalAlreadyClaimed, claimed)
	}

	log.Info("[Reconciliation] Total amount already claimed by all addresses: ", totalAlreadyClaimed)

	// What we really owe (total - already_claimed)
	totalLiabilities := big.NewInt(0).Sub(totalCumulativeRewards, totalAlreadyClaimed)

	log.Info("[Reconciliation] Total net liabilities (what we owe): ", totalLiabilities)

	log.Info("[Reconciliation] Total pool balance (what we have): ", contractBalanceWei)

	if totalLiabilities.Cmp(contractBalanceWei) != 0 {
		return errors.New(fmt.Sprintf("[Reconciliation] Liabilities and balance dont match: %d vs %d",
			totalLiabilities, contractBalanceWei))
	}

	log.Info("[Reconciliation] Success! Liabilities and balance match: ", totalLiabilities, " vs ", contractBalanceWei)

	return nil
}

func (or *Oracle) RunOffchainReconciliation() error {
	liabilities := big.NewInt(0)

	for _, val := range or.state.Validators {
		liabilities.Add(liabilities, val.AccumulatedRewardsWei)
		liabilities.Add(liabilities, val.PendingRewardsWei)
	}
	liabilities.Add(liabilities, or.state.PoolAccumulatedFees)

	assets := big.NewInt(0)

	for _, etherRx := range or.state.EtherReceivedEvents {
		assets.Add(assets, etherRx.DonationAmount)
	}
	for _, subs := range or.state.SubscriptionEvents {
		assets.Add(assets, subs.SubscriptionCollateral)
	}
	for _, vanilaBlock := range or.state.ProposedBlocks {
		if vanilaBlock.RewardType == VanilaBlock {
			assets.Add(assets, vanilaBlock.Reward)
		}
	}

	log.Info("[Offchain reconciliation] Liabilities: ", liabilities, "wei ", utils.WeiToEther(liabilities), " Ether")
	log.Info("[Offchain reconciliation] Assets: ", assets, "wei ", utils.WeiToEther(assets), " Ether")

	if liabilities.Cmp(assets) != 0 {
		return errors.New(fmt.Sprintf("Liabilities and assets dont match: %d vs %d",
			liabilities, assets))
	}

	return nil
}

func (or *Oracle) GetUniqueWithdrawalAddresses() []string {
	var uniqueWithAdd []string

	// Iterate all validators
	for _, validator := range or.State().Validators {
		skip := false
		// Iterate all unique deposit addresses processed before
		for _, u := range uniqueWithAdd {
			// If the deposit address is already in the list, skip it
			if utils.Equals(validator.WithdrawalAddress, u) {
				skip = true
				break
			}
		}
		// Not found, add it
		if !skip {
			uniqueWithAdd = append(uniqueWithAdd, validator.WithdrawalAddress)
		}
	}

	// Include also the pool address
	uniqueWithAdd = append(uniqueWithAdd, or.State().PoolFeesAddress)

	return uniqueWithAdd
}

// Returns true if a given validator can subscribe or not to the pool
// Accepted states are:
// -ValidatorStatePendingInitialized
// -ValidatorStatePendingQueued
// -ValidatorStateActiveOngoing
func CanValidatorSubscribeToPool(validator *v1.Validator) bool {
	if validator.Status != v1.ValidatorStateActiveExiting &&
		validator.Status != v1.ValidatorStateActiveSlashed &&
		validator.Status != v1.ValidatorStateExitedUnslashed &&
		validator.Status != v1.ValidatorStateExitedSlashed &&
		validator.Status != v1.ValidatorStateWithdrawalPossible &&
		validator.Status != v1.ValidatorStateWithdrawalDone &&
		validator.Status != v1.ValidatorStateUnknown {
		return true
	}
	return false
}

func (or *Oracle) hashStateLockFree() error {
	// We remove the hash before hashing, always hashing an empty hash
	or.state.StateHash = ""

	// Serialize the state
	jsonData, err := json.MarshalIndent(or.state, "", " ")
	if err != nil {
		return errors.Wrap(err, "could not marshal state to json")
	}

	// Calculate the hash of the state
	stateHash := sha256.Sum256(jsonData)
	stateHashStr := hexutil.Encode(stateHash[:])

	// Set the hash of the state
	or.state.StateHash = stateHashStr

	return nil
}

// Returns if a validator is subscribed to the pool. A validator is subscribed if
// its state is: active, yellowcard, redcard
func (or *Oracle) isSubscribed(validatorIndex uint64) bool {
	for valIndex, validator := range or.state.Validators {
		if valIndex == validatorIndex &&
			validator.ValidatorStatus != Banned &&
			validator.ValidatorStatus != NotSubscribed &&
			validator.ValidatorStatus != UnknownState {
			return true
		}
	}
	return false
}

// Returns true if a validator is banned
func (or *Oracle) isBanned(validatorIndex uint64) bool {
	validator, found := or.state.Validators[validatorIndex]
	if !found {
		return false
	}
	if validator.ValidatorStatus == Banned {
		return true
	}
	return false
}

// Returns true if a validator is tracked by the oracle. Any state is considered
// as tracked
func (or *Oracle) isTracked(validatorIndex uint64) bool {
	_, found := or.state.Validators[validatorIndex]
	if found {
		return true
	}
	return false
}

// Returns true if the given collateral is greater or equal the inputed one
func (or *Oracle) isCollateralEnough(collateral *big.Int) bool {
	return collateral.Cmp(or.state.CollateralInWei) >= 0
}

// Handles the donations of a given block
func (or *Oracle) handleDonations(donations []*contract.ContractEtherReceived) {
	// Ensure the donations are from the same block
	if len(donations) > 0 {
		blockReference := donations[0].Raw.BlockNumber
		for _, donation := range donations {
			if donation.Raw.BlockNumber != blockReference {
				log.Fatal("Handling donations from different blocks is not possible: ",
					donation.Raw.BlockNumber, " vs ", blockReference)
			}
		}
	}
	for _, donation := range donations {
		or.increaseAllPendingRewards(donation.DonationAmount)
		or.state.Donations = append(or.state.Donations, donation)
		log.WithFields(log.Fields{
			"RewardWei":   donation.DonationAmount,
			"BlockNumber": donation.Raw.BlockNumber,
			"Type":        "Donation",
			"TxHash":      donation.Raw.TxHash.String(),
		}).Info("[Reward]")
	}
}

// Handles a correct block proposal into the pool
func (or *Oracle) handleCorrectBlockProposal(block SummarizedBlock) {
	or.addSubscription(block.ValidatorIndex, block.WithdrawalAddress, block.ValidatorKey)
	or.advanceStateMachine(block.ValidatorIndex, ProposalOk)
	or.increaseAllPendingRewards(block.Reward)
	or.consolidateBalance(block.ValidatorIndex)
	or.state.ProposedBlocks = append(or.state.ProposedBlocks, block)

	log.WithFields(log.Fields{
		"Slot":       block.Slot,
		"Block":      block.Block,
		"ValIndex":   block.ValidatorIndex,
		"RewardWei":  block.Reward,
		"RewardType": block.RewardType.String(),
	}).Info("[Reward]")
}

// Handles the proposal of a block but that has BLS withdrawal keys
func (or *Oracle) handleBlsCorrectBlockProposal(block SummarizedBlock) {
	if block.BlockType != OkPoolProposalBlsKeys {
		log.Fatal("Block type is not OkPoolProposalBlsKeys, BlockType: ", block.BlockType)
	}

	log.WithFields(log.Fields{
		"Slot":       block.Slot,
		"Block":      block.Block,
		"ValIndex":   block.ValidatorIndex,
		"RewardWei":  block.Reward,
		"RewardType": block.RewardType.String(),
	}).Warn("[Reward] Block proposal was ok but bls keys are not supported, sending rewards to pool")

	or.sendRewardToPool(block.Reward)
	or.state.ProposedBlocks = append(or.state.ProposedBlocks, block)
}

// Handles a manual subscription to the pool, meaning that an event from the smart contract
// was triggered. This function asserts if the subscription was valid and updates the state
// of the validator accordingly
func (or *Oracle) handleManualSubscriptions(
	subsEvents []*contract.ContractSubscribeValidator,
	vals []*v1.Validator) {

	// Ensure the subscriptions events are from the same block
	if len(subsEvents) > 0 {
		blockReference := subsEvents[0].Raw.BlockNumber
		for _, donation := range subsEvents {
			if donation.Raw.BlockNumber != blockReference {
				log.Fatal("Handling manual subscriptions from different blocks is not possible: ",
					donation.Raw.BlockNumber, " vs ", blockReference)
			}
		}
	}

	if len(subsEvents) != len(vals) {
		log.Fatal("Number of subscriptions events and validators dont match: subs=",
			len(subsEvents), " vs vals=", len(vals))
	}

	for i, sub := range subsEvents {

		valIdx := sub.ValidatorID
		collateral := sub.SubscriptionCollateral
		sender := sub.Sender.String()
		validator := vals[i]

		// Subscription recevied for a validator index that doesnt exist
		if validator == nil {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"ValidatorIndex": valIdx,
			}).Warn("[Subscription]: for non existing validator, skipping")
			// Fees go to the pool, since validator does not exist in the pool and it is not tracked
			or.sendRewardToPool(collateral)
			continue
		}

		if valIdx != uint64(validator.Index) {
			log.Fatal("Subscription event validator index doesnt match the validator index. ",
				valIdx, " vs ", validator.Index)
		}

		// Subscription received for a validator that cannot subscribe (see states)
		if !CanValidatorSubscribeToPool(validator) {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"ValidatorIndex": valIdx,
				"ValidatorState": validator.Status,
			}).Warn("[Subscription]: for validator that cannot subscribe due to its state, skipping")
			// Fees go to the pool, since validator is not operational (withdrawn, slashed, etc)
			or.sendRewardToPool(collateral)
			continue
		}

		// Subscription received for a validator that dont have eth1 withdrawal address (bls)
		validatorWithdrawal, err := utils.GetEth1AddressByte(validator.Validator.WithdrawalCredentials)
		if err != nil {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"WithdrawalAddr": hexutil.Encode(validator.Validator.WithdrawalCredentials[:]),
				"ValidatorIndex": valIdx,
			}).Warn("[Subscription]: for validator with invalid withdrawal address (bls), skipping")
			// Fees go to the pool. A validator with a bls address can not be tracked since it has not been able to subscribe.
			or.sendRewardToPool(collateral)
			continue
		}

		// Subscription received from an address that is not the validator withdrawal address
		// AND sender is not the owner of a pod which staked validators include this transaction validator index
		if !utils.Equals(sender, validatorWithdrawal) {

			// Check if the sender is the owner of a pod which staked validators include this transaction validator index
			// If it is, we allow the subscription
			podAddress, err := or.getOwnerToPod(sub.Sender)
			if err != nil {
				log.Fatal("Error getting pod address from owner: ", err)
			}

			// if returned address is null, means the sender is not the owner of a pod
			if podAddress == (common.Address{}) {
				log.WithFields(log.Fields{
					"BlockNumber":         sub.Raw.BlockNumber,
					"Collateral":          sub.SubscriptionCollateral,
					"TxHash":              sub.Raw.TxHash,
					"ValidatorIndex":      valIdx,
					"Sender":              sender,
					"ValidatorWithdrawal": validatorWithdrawal,
				}).Warn("[Subscription] TX Sender is not the validator withdrawal address and has no EigenPod ownership, skipping")
				or.sendRewardToPool(collateral)
				continue
			}
			//filter events of eigenPod contract at address "podAddress" by ValidatorRestaked event and return the validators found
			validators, err := or.getValidatorRestakedEvents(podAddress)
			if err != nil {
				log.Fatal("Error getting ValidatorRestaked events: ", err)
			}
			// Check if the sender is the owner of a pod which staked validators include this transaction validator index
			// If it is, we allow the subscription. valIdx is uint64 and returned validators is an array of bigints, so we need to convert
			isOwner := false
			for _, v := range validators {
				if valIdx == v.Uint64() {
					isOwner = true
					break
				}
			}

			if !isOwner {
				log.WithFields(log.Fields{
					"BlockNumber":         sub.Raw.BlockNumber,
					"Collateral":          sub.SubscriptionCollateral,
					"TxHash":              sub.Raw.TxHash,
					"ValidatorIndex":      valIdx,
					"Sender":              sender,
					"ValidatorWithdrawal": validatorWithdrawal,
				}).Warn("[Subscription]: TX sender not the validator withdrawal address and it's pod does not own the validator, skipping")
				// Fees go to the pool.
				or.sendRewardToPool(collateral)
				continue
			}
		}

		// Subscription received for a banned validator
		if or.isBanned(valIdx) {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"ValidatorIndex": valIdx,
			}).Warn("[Subscription]: for banned validator, skipping")
			// Since we track this validator, give the collateral back
			or.increaseValidatorAccumulatedRewards(valIdx, collateral)
			continue
		}

		// Subscription received for an already subscribed validator
		if or.isSubscribed(valIdx) {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"ValidatorIndex": valIdx,
			}).Warn("[Subscription]: for an already subscribed validator, skipping")
			// Since we track this validator, return the collateral as accumulated balance
			or.increaseValidatorAccumulatedRewards(valIdx, collateral)
			continue
		}

		// Subscription received for a validator with not enough collateral
		if !or.isCollateralEnough(collateral) {
			log.WithFields(log.Fields{
				"BlockNumber":    sub.Raw.BlockNumber,
				"Collateral":     sub.SubscriptionCollateral,
				"TxHash":         sub.Raw.TxHash,
				"ValidatorIndex": valIdx,
			}).Warn("[Subscription]: for a validator with not enough collateral, skipping")
			// Fees go to the pool, since validator does not exist in the pool
			or.sendRewardToPool(collateral)
			continue
		}

		// Valid subscription
		if !or.isSubscribed(valIdx) &&
			!or.isBanned(valIdx) &&
			or.isCollateralEnough(collateral) {

			// Add valid subscription
			if !or.isTracked(valIdx) {
				// If its not tracked, we create a new subscription
				or.state.Validators[valIdx] = &ValidatorInfo{
					ValidatorStatus:       NotSubscribed,
					AccumulatedRewardsWei: big.NewInt(0),
					PendingRewardsWei:     big.NewInt(0),
					CollateralWei:         collateral,
					WithdrawalAddress:     validatorWithdrawal,
					ValidatorIndex:        valIdx,
					ValidatorKey:          hexutil.Encode(validator.Validator.PublicKey[:]),
				}
			}
			log.WithFields(log.Fields{
				"BlockNumber":      sub.Raw.BlockNumber,
				"Collateral":       sub.SubscriptionCollateral,
				"TxHash":           sub.Raw.TxHash,
				"ValidatorIndex":   valIdx,
				"WithdrawaAddress": validatorWithdrawal,
			}).Info("[Subscription]: Validator subscribed ok")
			or.state.Validators[valIdx].SubscriptionType = Manual
			or.increaseValidatorPendingRewards(valIdx, collateral)
			or.advanceStateMachine(valIdx, ManualSubscription)
			continue
		}

		// If we reach this point, its a case we havent considered, but its not valid
		log.WithFields(log.Fields{
			"BlockNumber":      sub.Raw.BlockNumber,
			"Collateral":       sub.SubscriptionCollateral,
			"TxHash":           sub.Raw.TxHash,
			"ValidatorIndex":   valIdx,
			"WithdrawaAddress": validatorWithdrawal,
		}).Info("[Subscription]: Not considered case meaning wrong subscription, skipping")
		// Send the collateral to the pool
		or.sendRewardToPool(collateral)
	}
}

// Handle the unsubscriptions detected as events triggered from the contract for a given block
// If the unsubscription matches some criteria, we update the state of the validator. Main criteria
// is that the sender matches the withdrawal address of the validator
func (or *Oracle) handleManualUnsubscriptions(
	unsubEvents []*contract.ContractUnsubscribeValidator,
	vals []*v1.Validator) {

	// Ensure the subscriptions events are from the same block
	if len(unsubEvents) > 0 {
		blockReference := unsubEvents[0].Raw.BlockNumber
		for _, unsub := range unsubEvents {
			if unsub.Raw.BlockNumber != blockReference {
				log.Fatal("Handling manual unsubscriptions from different blocks is not possible: ",
					unsub.Raw.BlockNumber, " vs ", blockReference)
			}
		}
	}

	if len(unsubEvents) != len(vals) {
		log.Fatal("Number of unsubscriptions events and validators dont match: ",
			len(unsubEvents), " vs ", len(vals))
	}

	for i, unsub := range unsubEvents {

		valIdx := unsub.ValidatorID
		sender := unsub.Sender.String()
		validator := vals[i]

		// Unsubscription but for a validator that doesnt exist
		if validator == nil {
			log.WithFields(log.Fields{
				"BlockNumber":    unsub.Raw.BlockNumber,
				"TxHash":         unsub.Raw.TxHash,
				"Sender":         sender,
				"ValidatorIndex": valIdx,
			}).Warn("[Unsubscription]: for validator index that does not exist, skipping")
			continue
		}

		if validator.Index != phase0.ValidatorIndex(valIdx) {
			log.Fatal("Unsubscription event validator index doesnt match the validator index. ",
				valIdx, " vs ", validator.Index)
		}

		// Unsubscription but for a validator that does not have an eth1 address. Should never happen
		withdrawalAddress, err := utils.GetEth1AddressByte(validator.Validator.WithdrawalCredentials)
		if err != nil {
			log.WithFields(log.Fields{
				"BlockNumber":    unsub.Raw.BlockNumber,
				"TxHash":         unsub.Raw.TxHash,
				"Sender":         sender,
				"ValidatorIndex": valIdx,
			}).Warn("[Unsubscription]: for validator with no eth1 withdrawal addres (bls), skipping")
			continue
		}

		// Its very important to check that the unsubscription was made from the withdrawal address
		// of the validator, otherwise anyone could call the unsubscription function.
		if !utils.Equals(sender, withdrawalAddress) {
			log.WithFields(log.Fields{
				"BlockNumber":      unsub.Raw.BlockNumber,
				"TxHash":           unsub.Raw.TxHash,
				"ValidatorIndex":   valIdx,
				"WithdrawaAddress": withdrawalAddress,
				"Sender":           sender,
			}).Warn("[Unsubscription] but sender does not match withdrawal address, skipping")
			continue
		}

		// After all the checks, we can proceed with the unsubscription
		if or.isSubscribed(valIdx) {
			or.advanceStateMachine(valIdx, Unsubscribe)
			or.increaseAllPendingRewards(or.state.Validators[valIdx].PendingRewardsWei)
			or.resetPendingRewards(valIdx)
			log.WithFields(log.Fields{
				"BlockNumber":      unsub.Raw.BlockNumber,
				"TxHash":           unsub.Raw.TxHash,
				"ValidatorIndex":   valIdx,
				"WithdrawaAddress": withdrawalAddress,
				"Sender":           sender,
			}).Info("[Unsubscription] Validator unsubscribed ok")
			continue
		}

		if !or.isSubscribed(valIdx) {
			log.WithFields(log.Fields{
				"BlockNumber":      unsub.Raw.BlockNumber,
				"TxHash":           unsub.Raw.TxHash,
				"ValidatorIndex":   valIdx,
				"WithdrawaAddress": withdrawalAddress,
				"Sender":           sender,
			}).Warn("[Unsubscription] but the validator is not subscribed, skipping")
			continue
		}

		// If we reach this point, its a case we havent considered, but its not valid
		log.WithFields(log.Fields{
			"BlockNumber":      unsub.Raw.BlockNumber,
			"TxHash":           unsub.Raw.TxHash,
			"ValidatorIndex":   valIdx,
			"WithdrawaAddress": withdrawalAddress,
			"Sender":           sender,
		}).Warn("[Unsubscription] Not considered case meaning wrong unsubscription, skipping")
	}
}

func (or *Oracle) handleManualBans(
	banEvents []*contract.ContractBanValidator) {

	// FIRST: healthy checks, ensure the bans events are okay.

	// Ensure the bans events are from the same block
	if len(banEvents) > 0 {
		blockReference := banEvents[0].Raw.BlockNumber
		for _, ban := range banEvents {
			if ban.Raw.BlockNumber != blockReference {
				log.Fatal("Handling manual bans from different blocks is not possible: ",
					ban.Raw.BlockNumber, " vs ", blockReference)
			}
		}
	}

	totalPending := big.NewInt(0)
	// SECOND: iterate over the ban events.
	//  - Advance state machine of all banned validators (move them to Banned state).
	//  - Sum all the pending rewards of the banned validators and share them among the rest.

	for _, ban := range banEvents {
		log.WithFields(log.Fields{
			"BlockNumber":    ban.Raw.BlockNumber,
			"TxHash":         ban.Raw.TxHash,
			"ValidatorIndex": ban.ValidatorID,
		}).Info("[Ban] Ban event received")
		or.advanceStateMachine(ban.ValidatorID, ManualBan)
		totalPending.Add(totalPending, or.state.Validators[ban.ValidatorID].PendingRewardsWei)

	}
	// THIRD: share the pending rewards of the banned validators among the rest
	or.increaseAllPendingRewards(totalPending)

	// FOURTH: reset the pending rewards of the banned validators
	for _, ban := range banEvents {
		or.resetPendingRewards(ban.ValidatorID)
	}
}

func (or *Oracle) handleManualUnbans(
	unbanEvents []*contract.ContractUnbanValidator) {

	// FIRST: healthy checks, ensure the unbans events are okay.
	if len(unbanEvents) > 0 {
		blockReference := unbanEvents[0].Raw.BlockNumber
		for _, ban := range unbanEvents {
			if ban.Raw.BlockNumber != blockReference {
				log.Fatal("Handling manual unbans from different blocks is not possible: ",
					ban.Raw.BlockNumber, " vs ", blockReference)
			}
		}
	}

	// SECOND: iterate over the unban events.
	//  - Advance state machine of all unbanned validators (move them to Active state).
	for _, unban := range unbanEvents {
		log.WithFields(log.Fields{
			"BlockNumber":    unban.Raw.BlockNumber,
			"TxHash":         unban.Raw.TxHash,
			"ValidatorIndex": unban.ValidatorID,
		}).Info("[Unban] Unban event received")
		or.advanceStateMachine(unban.ValidatorID, ManualUnban)
	}
}

// Banning a validator implies sharing its pending rewards among the rest
// of the validators and setting its pending to zero.
func (or *Oracle) handleBanValidator(block SummarizedBlock) {
	// First of all advance the state machine, so the banned validator is not
	// considered for the pending reward share
	or.advanceStateMachine(block.ValidatorIndex, ProposalWrongFee)
	or.increaseAllPendingRewards(or.state.Validators[block.ValidatorIndex].PendingRewardsWei)
	or.resetPendingRewards(block.ValidatorIndex)

	// Store the proof of the wrong fee block. Reason why it was banned
	or.state.WrongFeeBlocks = append(or.state.WrongFeeBlocks, block)
}

// Handles the case of a validator that has missed a block, only to be used
// with subscribed validators into the pool
func (or *Oracle) handleMissedBlock(block SummarizedBlock) {
	or.advanceStateMachine(block.ValidatorIndex, ProposalMissed)
	or.state.MissedBlocks = append(or.state.MissedBlocks, block)
}

// Subscribes a validator index with a given withdrawal address and validator key
func (or *Oracle) addSubscription(valIndex uint64, withdrawalAddress string, validatorKey string) {
	validator, found := or.state.Validators[valIndex]
	if !found {
		// If not found and not manually subscribed, we trigger the AutoSubscription event
		// Instantiate the validator
		validator = &ValidatorInfo{
			ValidatorStatus:       NotSubscribed,
			AccumulatedRewardsWei: big.NewInt(0),
			PendingRewardsWei:     big.NewInt(0),
			CollateralWei:         big.NewInt(0),
			WithdrawalAddress:     withdrawalAddress,
			ValidatorIndex:        valIndex,
			ValidatorKey:          validatorKey,
		}
		or.state.Validators[valIndex] = validator

		// And update it state according to the event
		or.advanceStateMachine(valIndex, AutoSubscription)

		// If subscription is new its auto
		or.state.Validators[valIndex].SubscriptionType = Auto
	} else {
		// If we found the validator and is not subscribed, advance the state machine
		// Most likely it was subscribed before, then unsubscribed and now auto subscribes
		if !or.isSubscribed(valIndex) {
			or.advanceStateMachine(valIndex, AutoSubscription)

			if !or.isBanned(valIndex) {
				// If it wasnt subscribed before, with this proposal its now auto
				or.state.Validators[valIndex].SubscriptionType = Auto
			}
		}
	}
}

// Consolidate the balance of a given validator index. This means moving the pending to its accumulated
// and setting the pending to zero.
func (or *Oracle) consolidateBalance(valIndex uint64) {

	beforePending := new(big.Int).Set(or.state.Validators[valIndex].PendingRewardsWei)
	beforeAccumulated := new(big.Int).Set(or.state.Validators[valIndex].AccumulatedRewardsWei)

	or.state.Validators[valIndex].AccumulatedRewardsWei.Add(or.state.Validators[valIndex].AccumulatedRewardsWei, or.state.Validators[valIndex].PendingRewardsWei)
	or.state.Validators[valIndex].PendingRewardsWei = big.NewInt(0)

	log.WithFields(log.Fields{
		"AccumulatedAfter":  or.state.Validators[valIndex].AccumulatedRewardsWei,
		"AccumulatedBefore": beforeAccumulated,
		"PendingAfter":      or.state.Validators[valIndex].PendingRewardsWei,
		"PendingBefore":     beforePending,
		"ValIndex":          valIndex,
	}).Debug("Consolidating balance")
}

// Returns a list of all the eligible validators for rewards.
func (or *Oracle) getEligibleValidators() []uint64 {
	eligibleValidators := make([]uint64, 0)

	for valIndex, validator := range or.state.Validators {
		if validator.ValidatorStatus == Active || validator.ValidatorStatus == YellowCard {
			eligibleValidators = append(eligibleValidators, valIndex)
		}
	}
	return eligibleValidators
}

// Increases the pending rewards of all validators, and gives the pool owner a cut
// of said rewards. Note that pending rewards cant be claimed until a block is proposed
// by the validator. But the pool owner can claim the pool cut at any time, so they are
// added as accumulated rewards.
func (or *Oracle) increaseAllPendingRewards(
	reward *big.Int) {

	eligibleValidators := or.getEligibleValidators()
	numEligibleValidators := big.NewInt(int64(len(eligibleValidators)))

	if len(eligibleValidators) == 0 {
		log.Warn("No validators are eligible to receive rewards, pool fees address will receive all")
		or.state.PoolAccumulatedFees.Add(or.state.PoolAccumulatedFees, reward)
		return
	}

	if or.state.PoolFeesPercentOver10000 > 100*100 {
		log.Fatal("Pool fees percent cannot be greater than 100% (10000) value: ", or.state.PoolFeesPercentOver10000)
	}

	// 100 is the % and the other 100 is because we use two decimals
	// eg 1000 is 10%
	// eg 50 is 0.5%
	over := big.NewInt(100 * 100)

	// The pool takes PoolFeesPercentOver10000 cut of the rewards
	aux := big.NewInt(0).Mul(reward, big.NewInt(int64(or.state.PoolFeesPercentOver10000)))

	// Calculate the pool cut (not taking into account the remainder)
	poolCut := big.NewInt(0).Div(aux, over)

	totalFees := big.NewInt(0)
	perValidatorReward := big.NewInt(0)
	if slotFork, found := SlotFork1[or.cfg.Network]; found {
		// Fixes minor bug in rewards calculation from a given slot. It just affects a few wei nothing
		// major, but this fixes the remainder1 not being scalled over 100.
		if or.state.NextSlotToProcess >= slotFork {

			log.WithFields(log.Fields{
				"SlotFork":           slotFork,
				"Slot":               or.state.NextSlotToProcess,
				"Network":            or.cfg.Network,
				"FeePercentOver1000": or.cfg.PoolFeesPercentOver10000,
				"Method":             "PostFork1",
			}).Debug("Calculating rewards")

			toShareAllValidators := big.NewInt(0).Sub(reward, poolCut)
			perValidatorReward = big.NewInt(0).Div(toShareAllValidators, numEligibleValidators)
			remainder := big.NewInt(0).Mod(toShareAllValidators, numEligibleValidators)
			totalFees = big.NewInt(0).Add(poolCut, remainder)
		} else {

			log.WithFields(log.Fields{
				"SlotFork":           slotFork,
				"Slot":               or.state.NextSlotToProcess,
				"Network":            or.cfg.Network,
				"FeePercentOver1000": or.cfg.PoolFeesPercentOver10000,
				"Method":             "PreFork1",
			}).Debug("Calculating rewards")

			// And remainder of above operation
			remainder1 := big.NewInt(0).Mod(aux, over)

			// The amount to share is the reward minus the pool cut + remainder
			toShareAllValidators := big.NewInt(0).Sub(reward, poolCut)
			toShareAllValidators.Sub(toShareAllValidators, remainder1)

			// Each validator gets that divided by numEligibleValidators
			perValidatorReward = big.NewInt(0).Div(toShareAllValidators, numEligibleValidators)
			// And remainder of above operation
			remainder2 := big.NewInt(0).Mod(toShareAllValidators, numEligibleValidators)

			// Total fees for the pool are: the cut (%) + the remainders
			totalFees = big.NewInt(0).Add(poolCut, remainder1)
			totalFees.Add(totalFees, remainder2)
		}
	} else {
		log.Fatal("Network not found in forks list: ", or.cfg.Network)
	}

	// Increase pool rewards (fees)
	or.state.PoolAccumulatedFees.Add(or.state.PoolAccumulatedFees, totalFees)

	// Extra check to ensure what we split and what we have match
	if big.NewInt(0).Add(big.NewInt(0).Mul(perValidatorReward, numEligibleValidators), totalFees).Cmp(reward) != 0 {
		log.WithFields(log.Fields{
			"perValidatorReward":    perValidatorReward,
			"totalFees":             totalFees,
			"numEligibleValidators": numEligibleValidators,
		}).Fatal("Total rewards dont match the sum of the rewards per validator and the pool fees")
	}

	log.WithFields(log.Fields{
		"AmountEligibleValidators": numEligibleValidators,
		"RewardPerValidatorWei":    perValidatorReward,
		"PoolFeesWei":              totalFees,
		"TotalRewardWei":           reward,
	}).Info("Increasing pending rewards of eligible validators")

	// Increase eligible validators rewards
	for _, eligibleIndex := range eligibleValidators {
		or.state.Validators[eligibleIndex].PendingRewardsWei.Add(or.state.Validators[eligibleIndex].PendingRewardsWei, perValidatorReward)
	}
}

// Increases the pending rewards of a given validator index.
func (or *Oracle) increaseValidatorPendingRewards(valIndex uint64, reward *big.Int) {
	beforePending := new(big.Int).Set(or.state.Validators[valIndex].PendingRewardsWei)
	or.state.Validators[valIndex].PendingRewardsWei.Add(or.state.Validators[valIndex].PendingRewardsWei, reward)

	log.WithFields(log.Fields{
		"PendingAfter":  or.state.Validators[valIndex].PendingRewardsWei,
		"PendingBefore": beforePending,
		"RewardShare":   reward,
		"ValIndex":      valIndex,
	}).Debug("Increasing validator pending rewards")
}

// Increases the accumulated rewards of a given validator index.
func (or *Oracle) increaseValidatorAccumulatedRewards(valIndex uint64, reward *big.Int) {
	accumulatedBefore := new(big.Int).Set(or.state.Validators[valIndex].AccumulatedRewardsWei)

	or.state.Validators[valIndex].AccumulatedRewardsWei.Add(or.state.Validators[valIndex].AccumulatedRewardsWei, reward)

	log.WithFields(log.Fields{
		"AccumulatedAfter":  or.state.Validators[valIndex].AccumulatedRewardsWei,
		"AccumulatedBefore": accumulatedBefore,
		"RewardShare":       reward,
		"ValIndex":          valIndex,
	}).Debug("Increasing validator accumulated rewards")
}

// Sends the pool cut to the pool reward address.
func (or *Oracle) sendRewardToPool(reward *big.Int) {

	poolAccumulatedBefore := new(big.Int).Set(or.state.PoolAccumulatedFees)
	or.state.PoolAccumulatedFees.Add(or.state.PoolAccumulatedFees, reward)

	log.WithFields(log.Fields{
		"PoolAccumulatedBefore": poolAccumulatedBefore,
		"PoolAccumulatedAfter":  or.state.PoolAccumulatedFees,
		"RewardShare":           reward,
	}).Debug("Sending reward cut to pool reward address")
}

// Resets the pending rewards of a given validator index.
func (or *Oracle) resetPendingRewards(valIndex uint64) {
	log.WithFields(log.Fields{
		"PendingRewardsBefore": or.state.Validators[valIndex].PendingRewardsWei,
		"ValIndex":             valIndex,
	}).Debug("Resetting pending rewards")
	or.state.Validators[valIndex].PendingRewardsWei = big.NewInt(0)
}

// Gets the merkle root of the state and returns if there was enough data
// to generate it or not.
func (or *Oracle) getMerkleRootIfAny() (string, bool) {
	mk := NewMerklelizer()
	_, _, tree, enoughData := mk.GenerateTreeFromState(or.state)
	if !enoughData {
		return "", enoughData
	}
	merkleRootStr := hexutil.Encode(tree.Root[:])

	return merkleRootStr, true
}

// Returns the 0x prefixed withdrawal credentials and its type: BlsWithdrawal or Eth1Withdrawal
func GetWithdrawalAndType(validator *v1.Validator) (string, WithdrawalType) {
	withdrawalCred := hex.EncodeToString(validator.Validator.WithdrawalCredentials)
	if len(withdrawalCred) != 64 {
		log.Fatal("withdrawal credentials are not a valid length: ", len(withdrawalCred))
	}

	if utils.IsBlsType(withdrawalCred) {
		return "0x" + withdrawalCred[2:], BlsWithdrawal
	} else if utils.IsEth1Type(withdrawalCred) {
		return "0x" + withdrawalCred[24:], Eth1Withdrawal
	}
	// can happen if a validator sets wrong withdrawal credentials (not very likely)
	// aka not respecting the 0x00 or 0x000000000000000000000000 prefixes
	// only concerning if the validator is subscribed to the pool
	log.WithFields(log.Fields{
		"WithdrawalCredentials": withdrawalCred,
		"ValidatorIndex":        validator.Index,
	}).Warn("withdrawal credentials are not valid, leaving empty")
	return "", 0
}

// See the spec for state diagram with states and transitions. This tracks all the different
// states and state transitions that a given validator can have from the oracle point of view
func (or *Oracle) advanceStateMachine(valIndex uint64, event Event) {
	switch or.state.Validators[valIndex].ValidatorStatus {
	case Active:
		switch event {
		case ProposalOk:
			log.WithFields(log.Fields{
				"Event":          "ProposalOk",
				"StateChange":    "Active -> Active",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Active
		case ProposalWrongFee:
			log.WithFields(log.Fields{
				"Event":          "ProposalWrongFee",
				"StateChange":    "Active -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned
		case ProposalMissed:
			log.WithFields(log.Fields{
				"Event":          "ProposalMissed",
				"StateChange":    "Active -> YellowCard",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = YellowCard
		case Unsubscribe:
			log.WithFields(log.Fields{
				"Event":          "Unsubscribe",
				"StateChange":    "Active -> NotSubscribed",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = NotSubscribed
		case ManualBan:
			log.WithFields(log.Fields{
				"Event":          "ManualBan",
				"StateChange":    "Active -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned

		}
	case YellowCard:
		switch event {
		case ProposalOk:
			log.WithFields(log.Fields{
				"Event":          "ProposalOk",
				"StateChange":    "YellowCard -> Active",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Active
		case ProposalWrongFee:
			log.WithFields(log.Fields{
				"Event":          "ProposalWrongFee",
				"StateChange":    "YellowCard -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned
		case ProposalMissed:
			log.WithFields(log.Fields{
				"Event":          "ProposalMissed",
				"StateChange":    "YellowCard -> RedCard",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = RedCard
		case Unsubscribe:
			log.WithFields(log.Fields{
				"Event":          "Unsubscribe",
				"StateChange":    "YellowCard -> NotSubscribed",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = NotSubscribed
		case ManualBan:
			log.WithFields(log.Fields{
				"Event":          "ManualBan",
				"StateChange":    "YellowCard -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned
		}
	case RedCard:
		switch event {
		case ProposalOk:
			log.WithFields(log.Fields{
				"Event":          "ProposalOk",
				"StateChange":    "RedCard -> YellowCard",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = YellowCard
		case ProposalWrongFee:
			log.WithFields(log.Fields{
				"Event":          "ProposalWrongFee",
				"StateChange":    "RedCard -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned
		case ProposalMissed:
			log.WithFields(log.Fields{
				"Event":          "ProposalMissed",
				"StateChange":    "RedCard -> RedCard",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = RedCard
		case Unsubscribe:
			log.WithFields(log.Fields{
				"Event":          "Unsubscribe",
				"StateChange":    "RedCard -> NotSubscribed",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = NotSubscribed
		case ManualBan:
			log.WithFields(log.Fields{
				"Event":          "ManualBan",
				"StateChange":    "RedCard -> Banned",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Banned
		}
	case NotSubscribed:
		switch event {
		case ManualSubscription:
			log.WithFields(log.Fields{
				"Event":          "ManualSubscription",
				"StateChange":    "NotSubscribed -> Active",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Active
		case AutoSubscription:
			log.WithFields(log.Fields{
				"Event":          "AutoSubscription",
				"StateChange":    "NotSubscribed -> Active",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Active
		}
	// TODO: A validator could return to the state it was after being banned, but we
	// return it always to the Active state for the sake of simplicity.
	case Banned:
		switch event {
		case ManualUnban:
			log.WithFields(log.Fields{
				"Event":          "ManualUnban",
				"StateChange":    "Banned -> Active",
				"ValidatorIndex": valIndex,
				"Slot":           or.state.NextSlotToProcess,
			}).Info("Validator state change")
			or.state.Validators[valIndex].ValidatorStatus = Active
		}
	}
}
