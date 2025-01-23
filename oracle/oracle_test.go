package oracle

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/avast/retry-go/v4"
	"github.com/dappnode/mev-sp-oracle/contract"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Tests AdvanceStateToNextSlot with real mocked data containing a variety of transactions
// events, reward, etc
func Test_AdvanceStateToNextSlot(t *testing.T) {

	// Run locally. Disabled since in CI we have some issues with git lfs bandwidth free limits
	t.Skip("Skipping test")

	oracleInstance := NewOracle(&Config{
		Network:                  "goerli",
		ConsensusEndpoint:        "http://127.0.0.1:5051",
		ExecutionEndpoint:        "http://127.0.0.1:8545",
		PoolAddress:              "0xF21fbbA423f3a893A2402d68240B219308AbCA46",
		CheckPointSizeInSlots:    7200,
		PoolFeesPercentOver10000: 1000,
		PoolFeesAddress:          "0xE46F9bE81f9a3ACA1808Bb8c36D353436bb96091",
		CollateralInWei:          big.NewInt(80000000000000000),
		DeployedSlot:             uint64(5840966), // same as first block
	})

	slotsToProcess := []uint64{
		5840966, //mev reward
		5843638, //vanila reward (auto subs)
		5844947, //vanila reward (auto subs)
		5846531, //mev reward
		5846747, //vanila reward (auto subs)
		5850959, //vanila reward (auto subs)
		5851651, //vanila reward (auto subs)
		5852212, //vanila reward (auto subs)
		5852262, //vanila reward (auto subs)
		5852659, //vanila reward (auto subs)
		5853824, //vanila reward (auto subs)
		5855268, //vanila reward (auto subs)
		5856619, //vanila reward (auto subs)
		5858585, //vanila reward (auto subs)
		5862054, //donation normal TODO
		5862104, //donation via smart contract TODO:

		5863539,
		5864096,
		5870291,
		5871368,
		5871701,
		5874576,
		5880967,
		5882954,
		5883240,
		5885240,
		5885987,
		5887583,

		// subs
		5888073,
		5888079,
		5888082,
		5888090, // already subs validator
		5888096,
		5888099,
		5888101,
		5888104,
		5888105,
		5888106,
		5888108,
		5888109,
		5888112,
		5888114,
		5888116,
		5888118,
		5888121,
		5888123,
		5888126,
		// freeze state

		// 0xb0f08efb67c59a4b16b143cf3a4850e786c4295909bee85b41cdc7d78db5d329

		5889932, // vanila rewar (auto subs)
		5890341, // vanila rewar (auto subs)
		5892032, // vanila rewar (auto subs)
		5893934, // vanila rewar (auto subs)
		5894030, // vanila rewar (auto subs)
		5895093, // vanila rewar (auto subs)
		5895373, // vanila rewar (already subscribed before with AUTO)

		5895384, // subs non existent validator
		5895415, //BLOCK!	9209397, // subscription of validator with BLS cred. skipped

		5896015, // vanila rewar (auto subs)
		5896730, // vanila rewar (auto subs)
		// 0xd9d4170d0a04dd0406961aaf574c18eda1c4f639226b1f2c85f9e91c5e211def

		5897820, // vanila rewar (auto subs of a validator subscribed before with MANUAL)    -> BUG HERE. subscription TYPE.

		5900857, // vanila rewar (auto subs)
		5901298, // vanila rewar (auto subs)

		5901838, //BLOCK 9214269,// subscription of validator with BLS cred. skipped
		5901840, //BLOCK 9214271// subscription of validator with BLS cred. skipped
		5901841, //BLOCK 9214272// subscription of validator with BLS cred. skipped
		5901843, //BLOCK 9214273// subscription of validator with BLS cred. skipped
		5901845, //BLOCK 9214275// subscription of validator with BLS cred. skipped
		5901846, //BLOCK 9214276// subscription of validator with BLS cred. skipped
		5901847, //BLOCK 9214277// subscription of validator with BLS cred. skipped
		5901849, //BLOCK 9214279// subscription of validator with BLS cred. skipped
		5901850, //BLOCK 9214280// subscription of validator with BLS cred. skipped
		5901852, //BLOCK 9214281// subscription of validator with BLS cred. skipped

		5901856, //block 9214285 // unsubscription
		5901861, //block 9214288 // unsubscription
		5901862, //block 9214289 // unsubscription
		5901865, //block 9214290 // unsubscription
		5901868, // block 9214293 // unsubscription
		5901870, //block 9214295 // unsubscription
		5901872, //block 9214296 // unsubscription
		5901874, //block 9214298 // unsubscription
		5901882, //block 9214306 // unsubscription
		5901885, //block 9214307 // unsubscription
		5901888, //block 9214310 		// unsubscription of an already unsubscribed validator.

		5902555, // vanila rewar (auto subs)
		// 0x5db21e0b873daedd188ff5976f4950d6f03d5db5bc46620036ecce45014792e9

		5904027, // vanila rewar (auto subs)
		5904240, // vanila rewar (auto subs)
		5907004, // vanila rewar (auto subs)
		5907780, // vanila rewar (auto subs)
		5908715, // vanila rewar (auto subs)

		5910468, // vanila rewar (auto subs)

		// 0x3b256a0d99ea9b781fb55349146c21209fd05deb5f33f605aa8868e82fbd3b03

		5911491, // vanila rewar (auto subs)

		5912693, // block number 9222701 // unsubscription of a validator that doesnt exist.
	}

	for _, slot := range slotsToProcess {
		var err1 error
		var err2 error
		var fullBlock *FullBlock

		// we force to process the slots we want
		oracleInstance.State().NextSlotToProcess = slot
		oracleInstance.State().LatestProcessedSlot = slot - 1

		// Quick way of loading the available block (with or without transactions)
		// Try with all transactions
		fullBlock, err1 = LoadFullBlock(slot, "5", true)

		// If not found try without transactions
		if err1 != nil {
			fullBlock, err2 = LoadFullBlock(slot, "5", false)
		}

		if err1 != nil && err2 != nil {
			require.Fail(t, "Failed to load block")
		}

		// Advance state to next slot based on the information we got from the block
		processedSlot, err := oracleInstance.AdvanceStateToNextSlot(fullBlock)
		require.NoError(t, err)

		log.Info("Processed slot: ", processedSlot)
	}

	oracleInstance.FreezeCheckpoint()
	//oracleInstance.SaveToJson(false)
	oracleInstance.RunOffchainReconciliation()
	//require.Equal(t, "0xb0f08efb67c59a4b16b143cf3a4850e786c4295909bee85b41cdc7d78db5d329", oracleInstance.LatestCommitedState().MerkleRoot)

}

func Test_SaveReadToFromJson(t *testing.T) {
	config := &Config{
		PoolAddress:              "0x0000000000000000000000000000000000000000",
		PoolFeesAddress:          "0x1000000000000000000000000000000000000000",
		PoolFeesPercentOver10000: 1000,
		Network:                  "mainnet",
		DeployedSlot:             1000,
		DeployedBlock:            1000,
		CheckPointSizeInSlots:    100,
		CollateralInWei:          big.NewInt(1000),
	}
	oracle := NewOracle(config)

	oracle.addSubscription(uint64(3), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(6434), "0x2000000000000000000000000000000000000000", "0x2000000000000000000000000000000000000000")

	oracle.FreezeCheckpoint()

	oracle.addSubscription(uint64(3), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(6434), "0x2000000000000000000000000000000000000000", "0x2000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(643344), "0x2000000000000000000000000000000000000000", "0x2000000000000000000000000000000000000000")

	oracle.FreezeCheckpoint()

	subs := []*contract.ContractSubscribeValidator{
		{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw: types.Log{
				TxHash:      [32]byte{0x1},
				Topics:      []common.Hash{{0x2}},
				Data:        []byte{0x3},
				BlockNumber: 124,
				BlockHash:   [32]byte{0x4},
				Index:       1,
				Removed:     false,
			},
			Sender: common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}
	oracle.state.SubscriptionEvents = subs

	defer os.Remove(filepath.Join(StateFolder, StateJsonName))
	defer os.RemoveAll(StateFolder)
	oracle.SaveToJson(false)

	// Oracle with same config
	newOracle := NewOracle(config)
	_, err := newOracle.LoadFromJson()
	require.NoError(t, err)

	oracle.state.StateHash = ""
	newOracle.state.StateHash = ""

	json1, err := json.MarshalIndent(oracle.state, "", " ")
	require.NoError(t, err)

	json2, err := json.MarshalIndent(newOracle.state, "", " ")
	require.NoError(t, err)

	// Serialized versions match
	require.Equal(t, string(json1), string(json2))

	// Structs match
	require.Equal(t, oracle.state, newOracle.state)

	// Now change the config
	config.Network = "testnet"
	config.PoolFeesAddress = "0xffff000000000000000000000000000000000000"

	oracleDifferentConfig := NewOracle(config)
	_, err = oracleDifferentConfig.LoadFromJson()

	// Expect error since the config now does now match
	require.Error(t, err)
}

func Test_SaveToJson(t *testing.T) {
	or := &Oracle{
		state: &OracleState{
			// Initialize the state with sample values for testing
			CommitedStates: map[uint64]*OnchainState{
				100: {
					Slot: 100,
				},
				200: {
					Slot: 200,
				},
			},
			LatestProcessedSlot:  200,
			LatestProcessedBlock: 100,
			NextSlotToProcess:    11,
			Validators: map[uint64]*ValidatorInfo{
				1: {
					ValidatorStatus:       1,
					AccumulatedRewardsWei: big.NewInt(100),
					PendingRewardsWei:     big.NewInt(50),
					CollateralWei:         big.NewInt(500),
					WithdrawalAddress:     "withdrawal_address_1",
					ValidatorIndex:        1,
					ValidatorKey:          "validator_key_1",
				},
				2: {
					ValidatorStatus:       1,
					AccumulatedRewardsWei: big.NewInt(200),
					PendingRewardsWei:     big.NewInt(100),
					CollateralWei:         big.NewInt(1000),
					WithdrawalAddress:     "withdrawal_address_2",
					ValidatorIndex:        2,
					ValidatorKey:          "validator_key_2",
				},
			},
			Network:     "testnet",
			PoolAddress: "pool_address",
			StateHash:   "state_hash",
		},
	}
	tempDir := t.TempDir()
	StateFolder = tempDir

	err := or.SaveToJson(false)
	require.NoError(t, err)

	expectedPath := filepath.Join(tempDir, StateJsonName)
	checkFileExists(t, expectedPath)
	err = os.RemoveAll(tempDir)
	require.NoError(t, err)

	err = or.SaveToJson(true)
	require.NoError(t, err)

	expectedPath1 := filepath.Join(tempDir, "state_200.json")
	expectedPath2 := filepath.Join(tempDir, "state.json")
	checkFileExists(t, expectedPath1)
	checkFileExists(t, expectedPath2)

	err = os.RemoveAll(tempDir)
	require.NoError(t, err)
}

func Test_FreezeCheckpoint(t *testing.T) {

	oracle := NewOracle(&Config{
		PoolFeesPercentOver10000: 0,
		PoolFeesAddress:          "0xfee0000000000000000000000000000000000000",
	})

	valInfo1 := &ValidatorInfo{
		ValidatorStatus:       Active,
		ValidatorIndex:        1,
		AccumulatedRewardsWei: big.NewInt(1000000000000000000),
		PendingRewardsWei:     big.NewInt(500000),
		WithdrawalAddress:     "0x1000000000000000000000000000000000000000",
	}

	valInfo2 := &ValidatorInfo{
		ValidatorStatus:       NotSubscribed,
		ValidatorIndex:        2,
		AccumulatedRewardsWei: big.NewInt(2000000000000000000),
		PendingRewardsWei:     big.NewInt(500000),
		// same withdrawal address as valInfo3
		WithdrawalAddress: "0x2000000000000000000000000000000000000000",
	}

	valInfo3 := &ValidatorInfo{
		ValidatorStatus:       NotSubscribed,
		ValidatorIndex:        3,
		AccumulatedRewardsWei: big.NewInt(2000000000000000000),
		PendingRewardsWei:     big.NewInt(500000),
		// same withdrawal address as valInfo2
		WithdrawalAddress: "0x2000000000000000000000000000000000000000",
	}

	oracle.state.Validators[1] = valInfo1
	oracle.state.Validators[2] = valInfo2
	oracle.state.Validators[3] = valInfo3

	// Function under test
	oracle.FreezeCheckpoint()

	commitedSlot := oracle.state.LatestProcessedSlot

	// Ensure all validators are present in the state
	require.Equal(t, valInfo1, oracle.state.CommitedStates[commitedSlot].Validators[1])
	require.Equal(t, valInfo2, oracle.state.CommitedStates[commitedSlot].Validators[2])
	require.Equal(t, valInfo3, oracle.state.CommitedStates[commitedSlot].Validators[3])

	// Ensure merkle root matches
	require.Equal(t, "0xd9a1eee574026532cddccbcce6320c0600f370a7c64ce30c5eafc63357449940", oracle.state.CommitedStates[commitedSlot].MerkleRoot)

	// Ensure proofs and leafs are correct
	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Proofs["0xfee0000000000000000000000000000000000000"], []string{"0x8bfb8acff6772a60d6641cb854587bb2b6f2100391fbadff2c34be0b8c20a0cc", "0x27205dd4c642acd1b1352617df2c4f410e20ff3fd6f3e3efddee9cea044921f8"})
	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Proofs["0x1000000000000000000000000000000000000000"], []string{"0xaaf838df9c8d5cec6ed77fcbc2cace945e8f2078eede4a0bb7164818d425f24d", "0x27205dd4c642acd1b1352617df2c4f410e20ff3fd6f3e3efddee9cea044921f8"})
	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Proofs["0x2000000000000000000000000000000000000000"], []string{"0xd643163144dcba353b4d27c50939b3d11133bd3c6916092de059d07353b4cb5f", "0xda53f5dd3e17f66f4a35c9c9d5fd27c094fa4249e2933fb819ac724476dc9ae1"})

	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Leafs["0xfee0000000000000000000000000000000000000"], RawLeaf{"0xfee0000000000000000000000000000000000000", big.NewInt(0)})
	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Leafs["0x1000000000000000000000000000000000000000"], RawLeaf{"0x1000000000000000000000000000000000000000", big.NewInt(1000000000000000000)})
	require.Equal(t, oracle.state.CommitedStates[commitedSlot].Leafs["0x2000000000000000000000000000000000000000"], RawLeaf{"0x2000000000000000000000000000000000000000", big.NewInt(4000000000000000000)})

	// Ensure LatestCommitedState contains a deep copy of the validators and not just a reference
	// This is very important since otherwise they will be modified when the state is modified
	// and we want a frozen snapshot of the state at that moment.

	// Do some changes in validators
	oracle.state.Validators[2].AccumulatedRewardsWei = big.NewInt(22)
	oracle.state.Validators[3].PendingRewardsWei = big.NewInt(22)

	// And assert the frozen state is not changes
	require.Equal(t, big.NewInt(2000000000000000000), oracle.state.CommitedStates[commitedSlot].Validators[2].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(500000), oracle.state.CommitedStates[commitedSlot].Validators[3].PendingRewardsWei)
}

func Test_LatestCommitedSlot_LatestCommitedState(t *testing.T) {
	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		PoolFeesAddress: "0x1123456789abcdef0123456789abcdef01234568",
	})

	// No data, no state
	oracle.FreezeCheckpoint()
	slot, stateExistst := oracle.LatestCommitedSlot()
	state := oracle.LatestCommitedState()
	require.Equal(t, uint64(0), slot)
	require.Equal(t, false, stateExistst)
	require.Nil(t, state)

	// Add state slot = 100
	oracle.state.LatestProcessedSlot = 100
	oracle.addSubscription(uint64(10), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(11), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(12), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.FreezeCheckpoint()
	slot, stateExistst = oracle.LatestCommitedSlot()
	state = oracle.LatestCommitedState()
	require.Equal(t, uint64(100), slot)
	require.Equal(t, true, stateExistst)
	require.Equal(t, uint64(100), state.Slot)
	require.Equal(t, 3, len(state.Validators))

	// Add state slot = 200
	oracle.state.LatestProcessedSlot = 200
	oracle.addSubscription(uint64(13), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(14), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(15), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.FreezeCheckpoint()
	slot, stateExistst = oracle.LatestCommitedSlot()
	state = oracle.LatestCommitedState()
	require.Equal(t, uint64(200), slot)
	require.Equal(t, true, stateExistst)
	require.Equal(t, uint64(200), state.Slot)
	require.Equal(t, 6, len(state.Validators))

}

func Test_IsOracleInSyncWithChain(t *testing.T) {

	oracle := NewOracle(&Config{
		PoolFeesAddress: "0x1123456789abcdef0123456789abcdef01234568",
	})

	// No states in oracle nor locally
	onchainRoot := DefaultRoot
	onchainSlot := uint64(0)
	isInSync, err := oracle.IsOracleInSyncWithChain(onchainRoot, onchainSlot)
	require.Equal(t, true, isInSync)
	require.NoError(t, err)

	// Add a state
	oracle.state.LatestProcessedSlot = 100
	oracle.addSubscription(uint64(10), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(11), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.addSubscription(uint64(12), "0x1000000000000000000000000000000000000000", "0x1000000000000000000000000000000000000000")
	oracle.FreezeCheckpoint()

	// In sync
	onchainRoot = "0xbb82bf59b1b6f3b0964c08ffb9336365153b34e2b30fb1230146428d153693b0"
	onchainSlot = uint64(100)
	isInSync, err = oracle.IsOracleInSyncWithChain(onchainRoot, onchainSlot)
	require.Equal(t, true, isInSync)
	require.NoError(t, err)

	// Not in sync
	onchainRoot = "0x1000000000000000000000000000000000000000000000000000000000000000"
	onchainSlot = uint64(200)
	isInSync, err = oracle.IsOracleInSyncWithChain(onchainRoot, onchainSlot)
	require.Equal(t, false, isInSync)
	require.NoError(t, err)

	// Roots match but not slots, expect error
	onchainRoot = "0xbb82bf59b1b6f3b0964c08ffb9336365153b34e2b30fb1230146428d153693b0"
	onchainSlot = uint64(200)
	isInSync, err = oracle.IsOracleInSyncWithChain(onchainRoot, onchainSlot)
	require.Equal(t, false, isInSync)
	require.Error(t, err)
}

func Test_GetUniqueDepositFromState(t *testing.T) {
	oracle1 := NewOracle(&Config{
		PoolFeesAddress: "0xfee0000000000000000000000000000000000000",
	})

	// Subscribe 3 validators with no balance
	oracle1.addSubscription(1, "0xa000000000000000000000000000000000000000", "0x")
	oracle1.addSubscription(2, "0xa000000000000000000000000000000000000000", "0x")
	oracle1.addSubscription(3, "0xa000000000000000000000000000000000000000", "0x")
	oracle1.addSubscription(4, "0xb000000000000000000000000000000000000000", "0x")
	oracle1.addSubscription(5, "0xc000000000000000000000000000000000000000", "0x")

	unique1 := oracle1.GetUniqueWithdrawalAddresses()
	require.Equal(t, 4, len(unique1))
	require.ElementsMatch(t, []string{
		"0xa000000000000000000000000000000000000000",
		"0xb000000000000000000000000000000000000000",
		"0xc000000000000000000000000000000000000000",
		"0xfee0000000000000000000000000000000000000"},
		unique1)

	oracle2 := NewOracle(&Config{
		PoolFeesAddress: "0xfee0000000000000000000000000000000000000",
	})

	// Subscribe 3 validators with no balance
	oracle2.addSubscription(1, "0xa000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(2, "0xa000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(3, "0xa000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(4, "0xb000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(5, "0xb000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(6, "0xc000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(7, "0xc000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(8, "0xd000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(9, "0xd000000000000000000000000000000000000000", "0x")
	oracle2.addSubscription(9, "0xc000000000000000000000000000000000000000", "0x")

	unique2 := oracle2.GetUniqueWithdrawalAddresses()
	require.Equal(t, 5, len(unique2))
	require.ElementsMatch(t, []string{
		"0xa000000000000000000000000000000000000000",
		"0xb000000000000000000000000000000000000000",
		"0xc000000000000000000000000000000000000000",
		"0xd000000000000000000000000000000000000000",
		"0xfee0000000000000000000000000000000000000"}, unique2)

	oracle3 := NewOracle(&Config{
		PoolFeesAddress: "0xfee0000000000000000000000000000000000000",
	})

	// Subscribe 3 validators with no balance
	oracle3.addSubscription(1, "0x1000000000000000000000000000000000000000", "0x")
	oracle3.addSubscription(2, "0x1000000000000000000000000000000000000000", "0x")
	oracle3.addSubscription(3, "0x1000000000000000000000000000000000000000", "0x")
	oracle3.addSubscription(4, "0x1000000000000000000000000000000000000000", "0x")
	oracle3.addSubscription(5, "0x1000000000000000000000000000000000000000", "0x")

	unique3 := oracle3.GetUniqueWithdrawalAddresses()
	require.Equal(t, 2, len(unique3))
	require.ElementsMatch(t, []string{
		"0x1000000000000000000000000000000000000000",
		"0xfee0000000000000000000000000000000000000"}, unique3)
}

func Test_Oracle_CanValidatorSubscribeToPool(t *testing.T) {

	val1 := &v1.Validator{
		Validator: &phase0.Validator{},
		Status:    v1.ValidatorStateActiveOngoing,
	}

	val2 := &v1.Validator{
		Validator: &phase0.Validator{},
		Status:    v1.ValidatorStateActiveExiting,
	}

	require.Equal(t, true, CanValidatorSubscribeToPool(val1))
	require.Equal(t, false, CanValidatorSubscribeToPool(val2))
}

func Test_addSubscription_1(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	oracle.addSubscription(10, "0x", "0x")
	oracle.increaseAllPendingRewards(big.NewInt(100))
	oracle.consolidateBalance(10)
	oracle.increaseAllPendingRewards(big.NewInt(200))
	require.Equal(t, big.NewInt(200), oracle.state.Validators[10].PendingRewardsWei)
	require.Equal(t, big.NewInt(100), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, Auto, oracle.state.Validators[10].SubscriptionType)

	// check that adding again doesnt reset the subscription
	oracle.addSubscription(10, "0x", "0x")
	require.Equal(t, big.NewInt(200), oracle.state.Validators[10].PendingRewardsWei)
	require.Equal(t, big.NewInt(100), oracle.state.Validators[10].AccumulatedRewardsWei)
}

func Test_addSubscription_2(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.addSubscription(uint64(100), "0x3000000000000000000000000000000000000000", "0xkey")
	require.Equal(t, 1, len(oracle.state.Validators))
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(0),
		PendingRewardsWei:     big.NewInt(0),
		CollateralWei:         big.NewInt(0),
		WithdrawalAddress:     "0x3000000000000000000000000000000000000000",
		ValidatorIndex:        100,
		ValidatorKey:          "0xkey",
		SubscriptionType:      Auto,
	}, oracle.state.Validators[100])

	// Modify the validator
	oracle.state.Validators[100].AccumulatedRewardsWei = big.NewInt(334545546)
	oracle.state.Validators[100].PendingRewardsWei = big.NewInt(87653)

	// If we call it again, it shouldnt be overwritten as its already there
	oracle.addSubscription(uint64(100), "0x3000000000000000000000000000000000000000", "0xkey")

	require.Equal(t, big.NewInt(334545546), oracle.state.Validators[100].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(87653), oracle.state.Validators[100].PendingRewardsWei)
}

func Test_handleDonations_PoolGetsAll(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	donations := []*contract.ContractEtherReceived{
		&contract.ContractEtherReceived{
			DonationAmount: big.NewInt(765432),
			Raw: types.Log{
				TxHash:      [32]byte{0x1},
				BlockNumber: uint64(100),
			},
		},
		&contract.ContractEtherReceived{
			DonationAmount: big.NewInt(30023456),
			Raw: types.Log{
				TxHash:      [32]byte{0x2},
				BlockNumber: uint64(100),
			},
		},
	}
	oracle.handleDonations(donations)

	require.Equal(t, big.NewInt(765432), oracle.state.Donations[0].DonationAmount)
	require.Equal(t, uint64(100), oracle.state.Donations[0].Raw.BlockNumber)
	require.Equal(t, "0x0100000000000000000000000000000000000000000000000000000000000000", oracle.state.Donations[0].Raw.TxHash.String())

	require.Equal(t, big.NewInt(30023456), oracle.state.Donations[1].DonationAmount)
	require.Equal(t, uint64(100), oracle.state.Donations[1].Raw.BlockNumber)
	require.Equal(t, "0x0200000000000000000000000000000000000000000000000000000000000000", oracle.state.Donations[1].Raw.TxHash.String())

	// No validators, pool gets it all
	require.Equal(t, big.NewInt(765432+30023456), oracle.state.PoolAccumulatedFees)
}

func Test_handleDonations_SharedEqual(t *testing.T) {
	oracle := NewOracle(&Config{
		PoolFeesPercentOver10000: 10 * 10, // 10%
		Network:                  "mainnet",
	})
	donations := []*contract.ContractEtherReceived{
		&contract.ContractEtherReceived{
			DonationAmount: big.NewInt(26543),
			Raw: types.Log{
				TxHash:      [32]byte{0x1},
				BlockNumber: uint64(100),
			},
		},
		&contract.ContractEtherReceived{
			DonationAmount: big.NewInt(100000),
			Raw: types.Log{
				TxHash:      [32]byte{0x2},
				BlockNumber: uint64(100),
			},
		},
	}
	oracle.addSubscription(10, "0x", "0x")
	oracle.addSubscription(20, "0x", "0x")
	oracle.handleDonations(donations)

	// Pool gets a share
	require.Equal(t, big.NewInt(5565), oracle.state.PoolAccumulatedFees)

	// Validator balances are updated ok
	require.Equal(t, big.NewInt(0), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].AccumulatedRewardsWei)

	require.Equal(t, big.NewInt(60489), oracle.state.Validators[10].PendingRewardsWei)
	require.Equal(t, big.NewInt(60489), oracle.state.Validators[20].PendingRewardsWei)
}

func Test_handleCorrectBlockProposal_AutoSubs(t *testing.T) {

	cfg := &Config{
		PoolFeesAddress:          "0xa",
		PoolFeesPercentOver10000: 0,
		CollateralInWei:          big.NewInt(1000000),
		Network:                  "mainnet",
	}

	oracle := NewOracle(cfg)

	// Block from a subscribed validator (manual)
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    10,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(50000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0ac",
	}
	oracle.handleCorrectBlockProposal(block1)

	require.Equal(t, big.NewInt(0), oracle.state.Validators[10].PendingRewardsWei)
	require.Equal(t, big.NewInt(50000000), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[10].ValidatorStatus)
}

func Test_handleCorrectBlockProposal_AlreadySub(t *testing.T) {

	cfg := &Config{
		PoolFeesAddress:          "0xa",
		PoolFeesPercentOver10000: 0,
		CollateralInWei:          big.NewInt(1000000),
		Network:                  "mainnet",
	}

	oracle := NewOracle(cfg)
	oracle.addSubscription(10, "0x", "0x")
	oracle.increaseValidatorPendingRewards(10, big.NewInt(1))
	oracle.increaseValidatorAccumulatedRewards(10, big.NewInt(1))

	// Block from a subscribed validator (manual)
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    10,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(50000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0ac",
	}
	oracle.handleCorrectBlockProposal(block1)

	require.Equal(t, big.NewInt(0), oracle.state.Validators[10].PendingRewardsWei)
	require.Equal(t, big.NewInt(50000000+1+1), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[10].ValidatorStatus)
	require.Equal(t, Auto, oracle.state.Validators[10].SubscriptionType)
}

func Test_handleManualSubscriptions_Valid(t *testing.T) {
	// Tests a valid subscription, with enough collateral to a not subscribed validator
	// and sent from the validator's withdrawal address

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, oracle.state.Validators[33], &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(0),
		PendingRewardsWei:     big.NewInt(1000),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        33,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	})
	require.Equal(t, 1, len(oracle.state.Validators))
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)
}

func Test_handleManualSubscriptions_FromWrongAddress(t *testing.T) {
	// Tests a subscription sent from a wrong address, meaning that it doesnt
	// match the validator's withdrawal address. No subscription is produced

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	sub1 := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	// No subscriptions are produced
	oracle.handleManualSubscriptions(sub1, vals)
	require.Equal(t, 0, len(oracle.state.Validators))
}

func Test_handleManualSubscriptions_AlreadySubscribed(t *testing.T) {
	// Test a subscription to an already subscribed validator, we return the collateral

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	// Run 3 subscriptions, only one should be added
	oracle.handleManualSubscriptions(subs, vals)

	require.Equal(t, oracle.state.Validators[33], &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(2000), // Second and third collateral are returned to the user
		PendingRewardsWei:     big.NewInt(1000),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        33,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	})
	require.Equal(t, 1, len(oracle.state.Validators))
}

func Test_handleManualSubscriptions_ThenSendBlock(t *testing.T) {
	// Test a subscription to an already subscribed validator, we return the collateral

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

	// Force auto block proposal
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

	// Another proposal keeps the validator subs type in manual
	block2 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block2)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)
}

func Test_handleManualSubscriptions_AutoThenSubscribe(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Force auto block proposal
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)

	// Now subscribe (but it was already auto subscribed)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)

	// State is keps in auto, since the subscription was ignored
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)
}

func Test_SubscribeUnsubscribe_Auto(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Now subscribe (but it was already auto subscribed)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}},
			Sender:      common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualUnsubscriptions(unsubs, vals)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

	// Force auto block proposal
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)

	// State is keps in auto, since the subscription was ignored
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)
}

func Test_AutoUnsubscribeThenManual(t *testing.T) { // TODO: Missing Then auto

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Force auto block proposal
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)

	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}},
			Sender:      common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualUnsubscriptions(unsubs, vals)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)

	// Now subscribe (but it was already auto subscribed)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

}

func Test_AutoUnsubscribeThenAuto(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Force auto block proposal
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)

	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}},
			Sender:      common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualUnsubscriptions(unsubs, vals)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)

	// Force auto block proposal again
	block2 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block2)
	require.Equal(t, Auto, oracle.state.Validators[33].SubscriptionType)
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
}

func Test_BannedValidatorAutoSubs(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Now subscribe (but it was already auto subscribed)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)

	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)

	// Force banned
	oracle.state.Validators[33].ValidatorStatus = Banned

	block2 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block2)

	require.Equal(t, Manual, oracle.state.Validators[33].SubscriptionType)
	require.Equal(t, Banned, oracle.state.Validators[33].ValidatorStatus)

	// Force auto block proposal again
	block3 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    33,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block3)
}

func Test_ManualBanUnbanValidators(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Now subscribe (but it was already auto subscribed)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)

	bans := []*contract.ContractBanValidator{
		&contract.ContractBanValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}, BlockNumber: 1},
		},
	}

	oracle.handleManualBans(bans)
	require.Equal(t, Banned, oracle.state.Validators[33].ValidatorStatus)

	unbans := []*contract.ContractUnbanValidator{
		&contract.ContractUnbanValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}, BlockNumber: 1},
		},
	}

	oracle.handleManualUnbans(unbans)
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
}


func TestBanUnbanWithRewards(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(0),
		Network:         "mainnet",
		PoolFeesPercentOver10000: 0, // No fees for the sake of test simplicity
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
		&v1.Validator{
			Index:  34,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey: phase0.BLSPubKey{128, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
		&v1.Validator{
			Index:  35,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 150, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey: phase0.BLSPubKey{127, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},

		&v1.Validator{
			Index:  36,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 151, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey: phase0.BLSPubKey{127, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(0),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            34,
			SubscriptionCollateral: big.NewInt(0),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            35,
			SubscriptionCollateral: big.NewInt(0),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{150, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            36,
			SubscriptionCollateral: big.NewInt(0),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{151, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
	require.Equal(t, Active, oracle.state.Validators[34].ValidatorStatus)
	require.Equal(t, Active, oracle.state.Validators[35].ValidatorStatus)
	require.Equal(t, Active, oracle.state.Validators[36].ValidatorStatus)

	// Give rewards to all validators.
	oracle.increaseValidatorAccumulatedRewards(33, big.NewInt(110000))
	oracle.increaseValidatorPendingRewards(33, big.NewInt(10000))

	oracle.increaseValidatorAccumulatedRewards(34, big.NewInt(120000))
	oracle.increaseValidatorPendingRewards(34, big.NewInt(10000))

	oracle.increaseValidatorAccumulatedRewards(35, big.NewInt(130000))
	oracle.increaseValidatorPendingRewards(35, big.NewInt(10000))

	oracle.increaseValidatorAccumulatedRewards(36, big.NewInt(140000))
	oracle.increaseValidatorPendingRewards(36, big.NewInt(10000))

	bans := []*contract.ContractBanValidator{
		&contract.ContractBanValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}, BlockNumber: 1},
		},
		&contract.ContractBanValidator{
			ValidatorID: 34,
			Raw:         types.Log{TxHash: [32]byte{0x1}, BlockNumber: 1},
		},
	}

	oracle.handleManualBans(bans)
	require.Equal(t, Banned, oracle.state.Validators[33].ValidatorStatus)

	// Test rewards after the validator ban. 
	require.Equal(t, big.NewInt(110000), oracle.state.Validators[33].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[33].PendingRewardsWei)

	require.Equal(t, big.NewInt(120000), oracle.state.Validators[34].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[34].PendingRewardsWei)

	require.Equal(t, big.NewInt(130000), oracle.state.Validators[35].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(20000), oracle.state.Validators[35].PendingRewardsWei)

	require.Equal(t, big.NewInt(140000), oracle.state.Validators[36].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(20000), oracle.state.Validators[36].PendingRewardsWei)

	unbans := []*contract.ContractUnbanValidator{
		&contract.ContractUnbanValidator{
			ValidatorID: 33,
			Raw:         types.Log{TxHash: [32]byte{0x1}, BlockNumber: 1},
		},
	}

	oracle.handleManualUnbans(unbans)
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[33].PendingRewardsWei)
	require.Equal(t, big.NewInt(110000), oracle.state.Validators[33].AccumulatedRewardsWei)

}



func Test_handleManualSubscriptions_AlreadySubscribed_WithBalance(t *testing.T) {
	// Test a subscription to an already subscribed validator, that already
	// has some balance. Assert that the existing balance is not touched and the
	// collateral is returned

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	val := &v1.Validator{
		Index:  33,
		Status: v1.ValidatorStateActiveOngoing,
		Validator: &phase0.Validator{
			// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
			WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
			PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
		},
	}

	sub1 := &contract.ContractSubscribeValidator{
		ValidatorID:            33,
		SubscriptionCollateral: big.NewInt(1000),
		Raw:                    types.Log{TxHash: [32]byte{0x1}},
		Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
	}

	// Validator is subscribed
	oracle.handleManualSubscriptions([]*contract.ContractSubscribeValidator{sub1}, []*v1.Validator{val})

	// And has some rewards
	oracle.increaseValidatorAccumulatedRewards(33, big.NewInt(9000))
	oracle.increaseValidatorPendingRewards(33, big.NewInt(44000))

	// Due to some mistake, the user subscribes again and again
	oracle.handleManualSubscriptions([]*contract.ContractSubscribeValidator{sub1, sub1}, []*v1.Validator{val, val})

	require.Equal(t, oracle.state.Validators[33], &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(9000 + 1000*2),
		PendingRewardsWei:     big.NewInt(44000 + 1000),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        33,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	})
	require.Equal(t, 1, len(oracle.state.Validators))
}

func Test_handleManualSubscriptions_Wrong_BlsCredentials(t *testing.T) {
	// A validator with wrong withdrawal address (bls) tries to subscribe. The validator
	// is nos subscribed and the collateral is given to the pool, since we dont have a way
	// to return it to its owner.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// bls address, not supported
				WithdrawalCredentials: []byte{0, 120, 22, 197, 153, 67, 183, 29, 244, 168, 13, 66, 101, 227, 165, 250, 41, 86, 97, 10, 40, 91, 140, 65, 154, 102, 143, 67, 117, 255, 140, 254},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, 0, len(oracle.state.Validators))
	require.Equal(t, big.NewInt(1000), oracle.state.PoolAccumulatedFees)
}

func Test_handleManualSubscriptions_NonExistent(t *testing.T) {
	// Test a subscription of a non-existent validator. Someone subscribes a validator
	// index that doesnt exist. Nothing happens, and the pool gets this collateral.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		nil, // dummuy validator
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)
	require.Equal(t, 0, len(oracle.state.Validators))
	require.Equal(t, big.NewInt(1000), oracle.state.PoolAccumulatedFees)
}

func Test_handleManualSubscriptions_WrongStateValidator(t *testing.T) {
	// Test a subscription of a validator in a wrong state (eg slashed validator or exited)
	// Nothing happens, and the pool gets this collateral.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  34,
			Status: v1.ValidatorStateActiveExiting,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			},
		},
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateExitedSlashed,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		// Its active but its exiting
		&contract.ContractSubscribeValidator{
			ValidatorID:            34,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x2}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		// Was slashed and exited
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)

	require.Equal(t, 0, len(oracle.state.Validators))
	require.Equal(t, big.NewInt(1000*2), oracle.state.PoolAccumulatedFees)
}

func Test_handleManualSubscriptions_BannedValidator(t *testing.T) {
	// Test a subscription of a banned validator. Check that the validator is not subscribed
	// and its kept in Banned state. Since we track this validator, we return the collateral
	// to the owner in good faith.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	bannedIndex := uint64(300000)

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  phase0.ValidatorIndex(bannedIndex),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			},
		},
	}

	oracle.state.Validators[bannedIndex] = &ValidatorInfo{
		ValidatorStatus:       Banned,
		AccumulatedRewardsWei: big.NewInt(0),
		PendingRewardsWei:     big.NewInt(0),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        bannedIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            bannedIndex,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)

	// Banned validator stays banned
	require.Equal(t, 1, len(oracle.state.Validators))

	// Note that since we track it, we return the collateral as accumulated rewards
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       Banned,
		AccumulatedRewardsWei: big.NewInt(1000),
		PendingRewardsWei:     big.NewInt(0),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        bannedIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}, oracle.state.Validators[bannedIndex])
}

func Test_handleManualSubscriptions(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},

		&v1.Validator{
			Index:  34,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{130, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},

		&v1.Validator{
			Index:  35,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 150, 39, 165, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{131, 170, 2, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	//Set up 3 new subs (val index 33,34,35), two valid and one invalid (low collateral)
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            34,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
		&contract.ContractSubscribeValidator{
			ValidatorID:            35,
			SubscriptionCollateral: big.NewInt(50),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{150, 39, 165, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	oracle.handleManualSubscriptions(subs, vals)

	// 3 validator tried to sub, 2 ok, 1 not enough collateral
	require.Equal(t, 2, len(oracle.state.Validators))

	//one validator subscribed with wrong collateral --> sent to the pool
	require.Equal(t, big.NewInt(50), oracle.state.PoolAccumulatedFees)

	// We keep track of [33 & 34] since subscription was valid
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
	require.Equal(t, Active, oracle.state.Validators[34].ValidatorStatus)

	// Accumulated rewards should be 0
	require.Equal(t, big.NewInt(0), oracle.state.Validators[33].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[34].AccumulatedRewardsWei)

	// Collateral should be 1000
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[33].CollateralWei)
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[34].CollateralWei)

	//Set up 2 new subs, both of already subscribed validators one sends configured collateral, the other does not
	subs2 := []*contract.ContractSubscribeValidator{
		// validator already subscribed sends subscription again with too much collateral
		&contract.ContractSubscribeValidator{
			ValidatorID:            33,
			SubscriptionCollateral: big.NewInt(5000000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},

		// validator already subscribed sends subscription again with correct collateral
		&contract.ContractSubscribeValidator{
			ValidatorID:            34,
			SubscriptionCollateral: big.NewInt(1000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}

	vals2 := []*v1.Validator{
		&v1.Validator{
			Index:  33,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{129, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},

		&v1.Validator{
			Index:  34,
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 149, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				// Valdator pubkey: 0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d
				PublicKey: phase0.BLSPubKey{130, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	oracle.handleManualSubscriptions(subs2, vals2)

	// [33] & [34] should still be active after trying to subscribe again
	require.Equal(t, Active, oracle.state.Validators[33].ValidatorStatus)
	require.Equal(t, Active, oracle.state.Validators[34].ValidatorStatus)

	// when an already subscribed validator manually subscribes again, we send the collateral to their accumulated rewards
	require.Equal(t, big.NewInt(5000000), oracle.state.Validators[33].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[34].AccumulatedRewardsWei)

	// Collateral does not change
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[33].CollateralWei)
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[34].CollateralWei)

	// Ban validator 34
	oracle.handleBanValidator(SummarizedBlock{
		Slot:           uint64(100),
		ValidatorIndex: uint64(34),
	})

	// Validator 34 should be banned
	require.Equal(t, Banned, oracle.state.Validators[34].ValidatorStatus)
	// Accumulated does not change
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[34].AccumulatedRewardsWei)

	// Accumulated rewards of other validators does not change because banned validator didnt have pending rewards
	require.Equal(t, big.NewInt(5000000), oracle.state.Validators[33].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(1000), oracle.state.Validators[34].AccumulatedRewardsWei)
}

func Test_handleManualUnsubscriptions_SubThenUnsubThenAuto(t *testing.T) {

	oracle := NewOracle(&Config{
		CollateralInWei:          big.NewInt(500000),
		PoolFeesPercentOver10000: 0,
		Network:                  "mainnet",
	})

	// Subscribe a validator
	valIdx := uint64(9000)

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  phase0.ValidatorIndex(valIdx),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey:             phase0.BLSPubKey{7, 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            valIdx,
			SubscriptionCollateral: big.NewInt(500000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}
	oracle.handleManualSubscriptions(subs, vals)

	require.Equal(t, Manual, oracle.state.Validators[valIdx].SubscriptionType)

	// Share some rewards with it
	oracle.state.Validators[valIdx].PendingRewardsWei = big.NewInt(10000)
	oracle.state.Validators[valIdx].AccumulatedRewardsWei = big.NewInt(20000)

	// Unsubscribe it
	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: valIdx,
			Sender:      common.Address{3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			Raw:         types.Log{TxHash: [32]byte{0x1}},
		},
	}

	oracle.handleManualUnsubscriptions(unsubs, vals)

	// Check is no longer subscribed and balances are kept (pending is reset)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[valIdx].PendingRewardsWei)
	require.Equal(t, big.NewInt(20000), oracle.state.Validators[valIdx].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[valIdx].ValidatorStatus)

	// Force automatic subscription
	block1 := SummarizedBlock{
		Slot:              0,
		ValidatorIndex:    valIdx,
		ValidatorKey:      "0x",
		Reward:            big.NewInt(90000000),
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
	}
	oracle.handleCorrectBlockProposal(block1)

	// Pending are 0 because the 90000000 are instantly consolidated
	require.Equal(t, big.NewInt(0), oracle.state.Validators[valIdx].PendingRewardsWei)
	// We have the new plus old ones
	require.Equal(t, big.NewInt(20000+90000000), oracle.state.Validators[valIdx].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[valIdx].ValidatorStatus)
	require.Equal(t, Auto, oracle.state.Validators[valIdx].SubscriptionType)
}

func Test_handleManualUnsubscriptions_ValidSubscription(t *testing.T) {
	// Unsubscribe an existing subscribed validator correctly, checking that the event is
	// sent from the withdrawal address of the validator. Check also that when unsubscribing
	// the pending validator rewards are shared among the rest of the validators.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(500000),
		Network:         "mainnet",
	})

	for _, valIdx := range []uint64{6, 9, 10, 15} {
		subs := []*contract.ContractSubscribeValidator{
			&contract.ContractSubscribeValidator{
				ValidatorID:            valIdx,
				SubscriptionCollateral: big.NewInt(500000),
				Raw:                    types.Log{TxHash: [32]byte{0x1}},
				Sender:                 common.Address{byte(valIdx), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			},
		}
		vals := []*v1.Validator{
			&v1.Validator{
				Index:  phase0.ValidatorIndex(valIdx),
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					// byte(valIdx) just to have different key/withdrawal addresses
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(valIdx), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
					PublicKey:             phase0.BLSPubKey{byte(valIdx), 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
				},
			},
		}
		oracle.handleManualSubscriptions(subs, vals)

		// Simulate some proposals increasing the rewards
		oracle.increaseValidatorAccumulatedRewards(valIdx, big.NewInt(3000))
		oracle.increaseValidatorPendingRewards(valIdx, big.NewInt(300000000000000000-500000))
	}

	require.Equal(t, 4, len(oracle.state.Validators))

	// Receive valid unsubscription event for index 6
	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: 6,
			// Same as withdrawal credential without the prefix
			Sender: common.Address{byte(6), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			Raw:    types.Log{TxHash: [32]byte{0x1}},
		},
	}
	vals := []*v1.Validator{
		&v1.Validator{
			Index:  phase0.ValidatorIndex(6),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// byte(valIdx) just to have different key/withdrawal addresses
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(6), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey:             phase0.BLSPubKey{byte(6), 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}
	oracle.handleManualUnsubscriptions(unsubs, vals)

	require.Equal(t, oracle.state.Validators[6], &ValidatorInfo{
		ValidatorStatus:       NotSubscribed,    // Validator is still tracked but not subscribed
		AccumulatedRewardsWei: big.NewInt(3000), // Accumulated rewards are kept
		PendingRewardsWei:     big.NewInt(0),    // Pending rewards are cleared
		CollateralWei:         big.NewInt(500000),
		WithdrawalAddress:     "0x0627a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        6,
		ValidatorKey:          "0x06aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	})
	require.Equal(t, 4, len(oracle.state.Validators))

	// The rest get the pending of valIndex=6
	require.Equal(t, oracle.state.Validators[9].PendingRewardsWei, big.NewInt(300000000000000000+300000000000000000/3))
	require.Equal(t, oracle.state.Validators[10].PendingRewardsWei, big.NewInt(300000000000000000+300000000000000000/3))
	require.Equal(t, oracle.state.Validators[15].PendingRewardsWei, big.NewInt(300000000000000000+300000000000000000/3))

	// And accumulated do not change
	require.Equal(t, oracle.state.Validators[9].AccumulatedRewardsWei, big.NewInt(3000))
	require.Equal(t, oracle.state.Validators[10].AccumulatedRewardsWei, big.NewInt(3000))
	require.Equal(t, oracle.state.Validators[15].AccumulatedRewardsWei, big.NewInt(3000))

	// And state of the rest is not changed
	require.Equal(t, oracle.state.Validators[9].ValidatorStatus, Active)
	require.Equal(t, oracle.state.Validators[10].ValidatorStatus, Active)
	require.Equal(t, oracle.state.Validators[15].ValidatorStatus, Active)

	// Unsubscribe all remaining validators
	newUnsubs := make([]*contract.ContractUnsubscribeValidator, 0)
	vals2 := make([]*v1.Validator, 0)
	for _, valIdx := range []uint64{ /*6*/ 9, 10, 15} {
		newUnsubs = append(newUnsubs,
			&contract.ContractUnsubscribeValidator{
				ValidatorID: valIdx,
				// Same as withdrawal credential without the prefix
				Sender: common.Address{byte(valIdx), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:    types.Log{TxHash: [32]byte{0x1}},
			})
		vals2 = append(vals2, &v1.Validator{
			Index:  phase0.ValidatorIndex(valIdx),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// byte(valIdx) just to have different key/withdrawal addresses
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(valIdx), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey:             phase0.BLSPubKey{byte(valIdx), 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		})
	}

	// Unsubscribe all at once
	oracle.handleManualUnsubscriptions(newUnsubs, vals2)

	require.Equal(t, 4, len(oracle.state.Validators))
	require.Equal(t, oracle.state.Validators[6].ValidatorStatus, NotSubscribed)
	require.Equal(t, oracle.state.Validators[9].ValidatorStatus, NotSubscribed)
	require.Equal(t, oracle.state.Validators[10].ValidatorStatus, NotSubscribed)
	require.Equal(t, oracle.state.Validators[15].ValidatorStatus, NotSubscribed)

	require.Equal(t, oracle.state.Validators[6].PendingRewardsWei, big.NewInt(0))
	require.Equal(t, oracle.state.Validators[9].PendingRewardsWei, big.NewInt(0))
	require.Equal(t, oracle.state.Validators[10].PendingRewardsWei, big.NewInt(0))
	require.Equal(t, oracle.state.Validators[15].PendingRewardsWei, big.NewInt(0))
}

func Test_handleManualUnsubscriptions_NonExistentValidator(t *testing.T) {
	// We receive an unsubscription for a validator that does not exist in the beacon
	// chain. Nothing happens to existing subscribed validators.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	vals := []*v1.Validator{
		nil,
	}

	// Simulate subscription of validator 33
	oracle.state.Validators[33] = &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(9000 + 1000*2), // Second and third collateral are added to accumulated rewards (returned)
		PendingRewardsWei:     big.NewInt(44000 + 1000),  // First collateral is added to pending (claimable in next block)
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        33,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}

	// Receive event of a validator index that doesnt exist in the beacon chain
	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: 900300,
			// Same as withdrawal credential without the prefix
			Sender: common.Address{byte(50), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			Raw:    types.Log{TxHash: [32]byte{0x1}},
		},
	}
	oracle.handleManualUnsubscriptions(unsubs, vals)

	// Check that the existing validator is not affected
	require.Equal(t, 1, len(oracle.state.Validators))
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(9000 + 1000*2), // Second and third collateral are added to accumulated rewards (returned)
		PendingRewardsWei:     big.NewInt(44000 + 1000),  // First collateral is added to pending (claimable in next block)
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        33,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}, oracle.state.Validators[33])
}

func Test_handleManualUnsubscriptions_NotSubscribedValidator(t *testing.T) {
	// We receive an unsubscription for a validator that is not subscribed but exists in
	// the beacon chain. Nothing happens, and no subscriptions are added.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	// Unsubscribe event of a validator index that BUT is not subscribed
	valIdx := uint64(730100)
	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: valIdx,
			// Same as withdrawal credential without the prefix
			Sender: common.Address{byte(valIdx), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			Raw:    types.Log{TxHash: [32]byte{0x1}},
		},
	}
	vals := []*v1.Validator{
		&v1.Validator{
			Index:  730100,
			Status: v1.ValidatorStateActiveExiting,
			Validator: &phase0.Validator{
				// Valid eth1 address: 0x9427a30991170f917d7b83def6e44d26577871ed
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			},
		},
	}
	oracle.handleManualUnsubscriptions(unsubs, vals)
	require.Equal(t, 0, len(oracle.state.Validators))
}

func Test_handleManualUnsubscriptions_FromWrongAddress(t *testing.T) {
	// An unsubscription for a subscribed validator is received, but the sender is not the
	// withdrawal address of that validator. Nothing happens to this validator

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(1000),
		Network:         "mainnet",
	})

	// Simulate subscription of validator 750100
	valIndex := uint64(750100)
	oracle.state.Validators[valIndex] = &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(5000000000000000000),
		PendingRewardsWei:     big.NewInt(3000000000000000000),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        valIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  phase0.ValidatorIndex(valIndex),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// byte(valIdx) just to have different key/withdrawal addresses
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(valIndex), 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey:             phase0.BLSPubKey{byte(valIndex + 1), 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: valIndex,
			// Wrong sender address (see WithdrawalCredentials)
			Sender: common.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Raw:    types.Log{TxHash: [32]byte{0x1}},
		},
	}
	oracle.handleManualUnsubscriptions(unsubs, vals)

	// Validator remains intact, since unsubscription event was wrong
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(5000000000000000000),
		PendingRewardsWei:     big.NewInt(3000000000000000000),
		CollateralWei:         big.NewInt(1000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        valIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}, oracle.state.Validators[valIndex])
}

func Test_handleManualUnsubscriptions_AndRejoin(t *testing.T) {
	// A validator subscribes, the unsubscribes and the rejoins. Check that its accumulated balances
	// are kept, and that it can rejoin succesfully.

	oracle := NewOracle(&Config{
		CollateralInWei: big.NewInt(500000),
		Network:         "mainnet",
	})

	valIndex := uint64(750100)

	vals := []*v1.Validator{
		&v1.Validator{
			Index:  phase0.ValidatorIndex(valIndex),
			Status: v1.ValidatorStateActiveOngoing,
			Validator: &phase0.Validator{
				// byte(valIdx) just to have different key
				WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				PublicKey:             phase0.BLSPubKey{byte(valIndex), 170, 231, 9, 230, 174, 231, 237, 73, 205, 21, 185, 65, 216, 91, 150, 122, 252, 200, 184, 68, 238, 32, 188, 126, 19, 150, 46, 132, 132, 87, 44, 27, 67, 212, 190, 117, 101, 33, 25, 236, 53, 60, 26, 50, 68, 62, 13},
			},
		},
	}

	// Simulate subscription of validator 750100

	oracle.state.Validators[valIndex] = &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(0),
		PendingRewardsWei:     big.NewInt(0),
		CollateralWei:         big.NewInt(500000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        valIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}

	// Add some rewards
	oracle.increaseValidatorAccumulatedRewards(valIndex, big.NewInt(1000000000000000000))
	oracle.increaseValidatorPendingRewards(valIndex, big.NewInt(5000000000000000000))

	// Now it unsubscribes ok
	unsubs := []*contract.ContractUnsubscribeValidator{
		&contract.ContractUnsubscribeValidator{
			ValidatorID: valIndex,
			// Wrong sender address (see WithdrawalCredentials)
			Sender: common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
			Raw:    types.Log{TxHash: [32]byte{0x1}},
		},
	}
	oracle.handleManualUnsubscriptions(unsubs, vals)

	// Unsubscription is ok
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       NotSubscribed,
		AccumulatedRewardsWei: big.NewInt(1000000000000000000),
		PendingRewardsWei:     big.NewInt(0),
		CollateralWei:         big.NewInt(500000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        valIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}, oracle.state.Validators[valIndex])

	// Now the same validator tries to rejoin
	subs := []*contract.ContractSubscribeValidator{
		&contract.ContractSubscribeValidator{
			ValidatorID:            valIndex,
			SubscriptionCollateral: big.NewInt(500000),
			Raw:                    types.Log{TxHash: [32]byte{0x1}},
			Sender:                 common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}
	oracle.handleManualSubscriptions(subs, vals)

	// Its subscribed again with its old accumulated rewards
	require.Equal(t, &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(1000000000000000000),
		PendingRewardsWei:     big.NewInt(500000),
		CollateralWei:         big.NewInt(500000),
		WithdrawalAddress:     "0x9427a30991170f917d7b83def6e44d26577871ed",
		ValidatorIndex:        valIndex,
		ValidatorKey:          "0x81aae709e6aee7ed49cd15b941d85b967afcc8b844ee20bc7e13962e8484572c1b43d4be75652119ec353c1a32443e0d",
	}, oracle.state.Validators[valIndex])
}

func Test_handleBanValidator(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	oracle.addSubscription(1, "0xa", "0xb")
	oracle.addSubscription(2, "0xa", "0xb")
	oracle.addSubscription(3, "0xa", "0xb")

	// New reward arrives
	oracle.increaseAllPendingRewards(big.NewInt(99))

	// Shared equally among all validators
	require.Equal(t, big.NewInt(33), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(33), oracle.state.Validators[2].PendingRewardsWei)
	require.Equal(t, big.NewInt(33), oracle.state.Validators[3].PendingRewardsWei)

	// Ban validator 3
	oracle.handleBanValidator(SummarizedBlock{ValidatorIndex: 3})

	// Its pending balance is shared equally among the rest
	require.Equal(t, big.NewInt(49), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(49), oracle.state.Validators[2].PendingRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[3].PendingRewardsWei)

	// The pool fee address gets the rounding errors (1 wei, neglectable)
	require.Equal(t, big.NewInt(1), oracle.state.PoolAccumulatedFees)
}

func Test_handleMissedBlock(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	oracle.addSubscription(1, "0xa", "0xb")
	oracle.addSubscription(2, "0xa", "0xb")

	oracle.increaseValidatorPendingRewards(1, big.NewInt(100))
	oracle.increaseValidatorAccumulatedRewards(1, big.NewInt(200))

	missed := SummarizedBlock{
		Slot:              uint64(100),
		ValidatorIndex:    uint64(1),
		ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
		BlockType:         MissedProposal,
		RewardType:        VanilaBlock,
		WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
	}

	oracle.handleMissedBlock(missed)
	// State is updated
	require.Equal(t, YellowCard, oracle.state.Validators[1].ValidatorStatus)
	// Rewards are not touched
	require.Equal(t, big.NewInt(100), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(200), oracle.state.Validators[1].AccumulatedRewardsWei)
	require.Equal(t, missed, oracle.state.MissedBlocks[0])

	// Missed again
	oracle.handleMissedBlock(missed)
	// State is updated
	require.Equal(t, RedCard, oracle.state.Validators[1].ValidatorStatus)
	// Rewards are not touched
	require.Equal(t, big.NewInt(100), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(200), oracle.state.Validators[1].AccumulatedRewardsWei)
}

func Test_handleBlsCorrectBlockProposal_NotSubscribed(t *testing.T) {
	oracle := NewOracle(&Config{
		PoolFeesPercentOver10000: 100, // 1%
	})

	oracle.addSubscription(888, "0xa", "0xb")
	oracle.addSubscription(999, "0xa", "0xb")

	blsBlock := SummarizedBlock{
		Block:             1,
		Slot:              uint64(100),
		ValidatorIndex:    uint64(1),
		ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
		BlockType:         OkPoolProposalBlsKeys,
		RewardType:        VanilaBlock,
		Reward:            big.NewInt(100),
		WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
	}

	oracle.handleBlsCorrectBlockProposal(blsBlock)

	// no automatic subscription is produced. The 2 validators are not the BLS ones.
	require.Equal(t, 2, len(oracle.state.Validators))

	// all rewards go to the pool
	require.Equal(t, big.NewInt(100), oracle.state.PoolAccumulatedFees)

	// Run reconciliation
	err := oracle.RunOffchainReconciliation()
	require.NoError(t, err)

	require.Equal(t, 1, len(oracle.state.ProposedBlocks))
	require.Equal(t, blsBlock, oracle.state.ProposedBlocks[0])

	// check all validators rewards.
	require.Equal(t, big.NewInt(0), oracle.state.Validators[888].PendingRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[999].PendingRewardsWei)

	// The proposer is not tracked
	_, exists := oracle.state.Validators[1]
	require.False(t, exists)
}

func Test_handleBlsCorrectBlockProposal_Subscribed(t *testing.T) {
	// This should never happen
	oracle := NewOracle(&Config{})
	oracle.addSubscription(1, "0xa", "0xb")

	missed := SummarizedBlock{
		Slot:              uint64(100),
		ValidatorIndex:    uint64(1),
		ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
		BlockType:         OkPoolProposalBlsKeys,
		RewardType:        VanilaBlock,
		Reward:            big.NewInt(100),
		WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
	}

	oracle.handleBlsCorrectBlockProposal(missed)

	// no automatic subscription is produced
	require.Equal(t, 1, len(oracle.state.Validators))

	// validator gets no rewards
	require.Equal(t, big.NewInt(0), oracle.state.Validators[1].PendingRewardsWei)

	// pool gets rewards
	require.Equal(t, big.NewInt(100), oracle.state.PoolAccumulatedFees)
}

func Test_increaseAllPendingRewards_1(t *testing.T) {

	oracle := NewOracle(&Config{
		PoolFeesPercentOver10000: 0,
		PoolFeesAddress:          "0x",
		Network:                  "mainnet",
	})

	// Subscribe 3 validators with no balance
	oracle.addSubscription(1, "0x", "0x")
	oracle.addSubscription(2, "0x", "0x")
	oracle.addSubscription(3, "0x", "0x")

	oracle.increaseAllPendingRewards(big.NewInt(10000))

	// Note that in this case even with PoolFeesPercentOver10000: 0, the pool gets the remainder
	require.Equal(t, big.NewInt(3333), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(3333), oracle.state.Validators[2].PendingRewardsWei)
	require.Equal(t, big.NewInt(3333), oracle.state.Validators[3].PendingRewardsWei)
	require.Equal(t, big.NewInt(1), oracle.state.PoolAccumulatedFees)
}

func Test_increaseAllPendingRewards_2(t *testing.T) {

	oracle := NewOracle(&Config{
		PoolFeesPercentOver10000: 10 * 100, // 10%
		PoolFeesAddress:          "0x",
		Network:                  "mainnet",
	})

	// Subscribe 3 validators with no balance
	oracle.addSubscription(1, "0x", "0x")
	oracle.addSubscription(2, "0x", "0x")
	oracle.addSubscription(3, "0x", "0x")

	oracle.increaseAllPendingRewards(big.NewInt(10000))

	// Note that in this case even with PoolFeesPercentOver10000: 0, the pool gets the remainder
	require.Equal(t, big.NewInt(3000), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(3000), oracle.state.Validators[2].PendingRewardsWei)
	require.Equal(t, big.NewInt(3000), oracle.state.Validators[3].PendingRewardsWei)
	require.Equal(t, big.NewInt(1000), oracle.state.PoolAccumulatedFees)
}

func Test_increaseAllPendingRewards_3(t *testing.T) {

	// Multiple test with different combinations of: fee, reward, validators

	type pendingRewardTest struct {
		FeePercent       int
		Reward           []*big.Int
		AmountValidators int
	}

	tests := []pendingRewardTest{
		// FeePercent |Reward | AmountValidators
		{0, []*big.Int{big.NewInt(100)}, 1},
		{0, []*big.Int{big.NewInt(500)}, 2},
		{0, []*big.Int{big.NewInt(398)}, 3},
		{10 * 100, []*big.Int{big.NewInt(0)}, 1},
		{15 * 100, []*big.Int{big.NewInt(23033)}, 1},
		{33 * 100, []*big.Int{big.NewInt(99999)}, 5},
		{33 * 100, []*big.Int{big.NewInt(1)}, 5},
		{33 * 100, []*big.Int{big.NewInt(1), big.NewInt(403342)}, 200},
		{12 * 100, []*big.Int{big.NewInt(32000000000000), big.NewInt(333333333333), big.NewInt(345676543234567)}, 233},
		{14 * 100, []*big.Int{big.NewInt(32000000000000), big.NewInt(333333333333), big.NewInt(345676543234567), big.NewInt(9)}, 99},
	}

	for _, test := range tests {
		oracle := NewOracle(&Config{
			PoolFeesPercentOver10000: test.FeePercent,
			PoolFeesAddress:          "0x",
			Network:                  "mainnet",
		})

		for i := 0; i < test.AmountValidators; i++ {
			oracle.addSubscription(uint64(i), "0x", "0x")
		}

		totalRewards := big.NewInt(0)
		for _, reward := range test.Reward {
			oracle.increaseAllPendingRewards(reward)
			totalRewards.Add(totalRewards, reward)
		}

		totalDistributedRewards := big.NewInt(0)
		totalDistributedRewards.Add(totalDistributedRewards, oracle.state.PoolAccumulatedFees)
		for i := 0; i < test.AmountValidators; i++ {
			totalDistributedRewards.Add(totalDistributedRewards, oracle.state.Validators[uint64(i)].PendingRewardsWei)
		}

		// Assert that the rewards that were shared, equal the ones that we had
		// kirchhoff law, what comes in = what it goes out!
		require.Equal(t, totalDistributedRewards, totalRewards)
	}
}

func Test_increaseAllPendingRewards_4(t *testing.T) {

	// Multiple test with different combinations of: fee, reward, validators

	type pendingRewardTest struct {
		FeePercentX100   int
		Reward           *big.Int
		AmountValidators int
		NewPendingArray  []*big.Int
		FeesAddress      *big.Int
	}

	tests := []pendingRewardTest{
		// FeePercentX100 (100 = 1%) | Reward | AmountValidators | RewardsPerValidator

		// 0%
		{0, big.NewInt(0), 1, []*big.Int{big.NewInt(0)}, big.NewInt(0)},

		// 100%
		{100 * 100, big.NewInt(2345676543), 0, []*big.Int{big.NewInt(0)}, big.NewInt(2345676543)},

		// 1%
		{1 * 100, big.NewInt(100), 1, []*big.Int{big.NewInt(99)}, big.NewInt(1)},

		// 10%
		{10 * 100, big.NewInt(10000000000), 1, []*big.Int{big.NewInt(9000000000)}, big.NewInt(1000000000)},

		// 10 % with 2 validators
		{10 * 100, big.NewInt(10000000000), 2, []*big.Int{big.NewInt(4500000000), big.NewInt(4500000000)}, big.NewInt(1000000000)},

		// 0%
		{0, big.NewInt(555555555555), 5, []*big.Int{big.NewInt(111111111111), big.NewInt(111111111111), big.NewInt(111111111111), big.NewInt(111111111111), big.NewInt(111111111111)}, big.NewInt(0)},

		// 2.5%
		{2.5 * 100, big.NewInt(15000), 5, []*big.Int{big.NewInt(2925), big.NewInt(2925), big.NewInt(2925), big.NewInt(2925), big.NewInt(2925)}, big.NewInt(375)},

		// 0.25%: 87654567898 * 25 / 10000 = 219136419 (+ remainder of 7450). Then 887654567898-(219136419 + 7450))/5 = 17487084805 (+ reminder of 4)
		{0.25 * 100, big.NewInt(87654567898), 5, []*big.Int{big.NewInt(17487084805), big.NewInt(17487084805), big.NewInt(17487084805), big.NewInt(17487084805), big.NewInt(17487084805)}, big.NewInt(219136419 + 7450 + 4)},
	}

	for _, test := range tests {
		oracle := NewOracle(&Config{
			PoolFeesPercentOver10000: test.FeePercentX100,
			PoolFeesAddress:          "0x",
			Network:                  "mainnet",
		})
		for i := 0; i < test.AmountValidators; i++ {
			oracle.addSubscription(uint64(i), "0x", "0x")
		}
		oracle.increaseAllPendingRewards(test.Reward)
		for i := 0; i < test.AmountValidators; i++ {
			require.Equal(t, test.NewPendingArray[i], oracle.state.Validators[uint64(i)].PendingRewardsWei)
		}

		require.Equal(t, test.FeesAddress, oracle.state.PoolAccumulatedFees)

		// Ensure that what we gave away matches the reward we had
		totalDistributedRewards := big.NewInt(0)
		totalDistributedRewards.Add(totalDistributedRewards, oracle.state.PoolAccumulatedFees)

		// Since all validators rewards are equal, just take the reward of the first one
		rewardsToValidators := big.NewInt(0)
		if len(oracle.state.Validators) > 0 {
			rewardsToValidators = oracle.state.Validators[0].PendingRewardsWei
		}
		rewardsToValidators.Mul(rewardsToValidators, big.NewInt(int64(test.AmountValidators)))
		totalDistributedRewards.Add(totalDistributedRewards, rewardsToValidators)
		require.Equal(t, totalDistributedRewards, test.Reward)
	}
}

func Test_increaseAllPendingRewards_5(t *testing.T) {

	MainnetRewardsSlotFork := SlotFork1["mainnet"]

	type pendingRewardTest struct {
		FeePercentX100   int
		Reward           *big.Int
		AmountValidators int
		ValidatorReward  *big.Int
		PoolReward       *big.Int
		Slot             uint64
	}

	tests := []pendingRewardTest{
		// FeePercentX100 (100 = 1%) | Reward | AmountValidators | ValidatorReward | PoolReward | Slot

		// 0%
		{0, big.NewInt(0), 1, big.NewInt(0), big.NewInt(0), 0},

		// 7 %
		{7 * 100, big.NewInt(30500333098045431), 1665, big.NewInt(17036222090799), big.NewInt(2135023316865096), 0},

		// 7 %
		{7 * 100, big.NewInt(200000000000000000), 2, big.NewInt(93000000000000000), big.NewInt(14000000000000000), 0},

		// 7 %
		{7 * 100, big.NewInt(43446233255383379), 503, big.NewInt(80328025700797), big.NewInt(3041236327882488), 0},

		// 7 %
		{7 * 100, big.NewInt(24448528911304907), 1034, big.NewInt(21989489252909), big.NewInt(1711397023797001), 0},

		// 7 % (before change, was wrong)
		{7 * 100, big.NewInt(1), 1670, big.NewInt(-1), big.NewInt(1671), MainnetRewardsSlotFork - 1},

		// 7 % (before change, was wrong)
		{7 * 100, big.NewInt(751283351135293312), 1959, big.NewInt(356658252453197), big.NewInt(52589834579480389), MainnetRewardsSlotFork - 1},

		// Examples taken from mainnet
		{700, big.NewInt(48913644130771431), 1975, big.NewInt(23032753945121), big.NewInt(3423955089157456), MainnetRewardsSlotFork - 1},
		{700, big.NewInt(1511244003192947), 1975, big.NewInt(711623758464), big.NewInt(105787080226547), MainnetRewardsSlotFork - 1},
		{700, big.NewInt(24117725873436208), 1976, big.NewInt(11350953978891), big.NewInt(1688240811147592), MainnetRewardsSlotFork - 10000},
		{700, big.NewInt(91043549594056858), 1975, big.NewInt(42871139808846), big.NewInt(6373048471586008), MainnetRewardsSlotFork - 1},

		// Wrong but legacy
		{700, big.NewInt(1), 1987, big.NewInt(-1), big.NewInt(1988), MainnetRewardsSlotFork - 100000},

		// New fixed
		{700, big.NewInt(1), 1987, big.NewInt(0), big.NewInt(1), MainnetRewardsSlotFork + 100000},

		// Reward of 0 does not break
		{700, big.NewInt(0), 1987, big.NewInt(0), big.NewInt(0), MainnetRewardsSlotFork + 100000},
		{700, big.NewInt(0), 1987, big.NewInt(0), big.NewInt(0), MainnetRewardsSlotFork - 100000},

		// After fork (exact slot)
		{7 * 100, big.NewInt(1), 1670, big.NewInt(0), big.NewInt(1), MainnetRewardsSlotFork},

		// After fork (exact slot)
		{7 * 100, big.NewInt(751283351135293312), 1959, big.NewInt(356658252453202), big.NewInt(52589834579470594), MainnetRewardsSlotFork},

		// After fork (few slot later)
		{7 * 100, big.NewInt(751283351135293312), 1959, big.NewInt(356658252453202), big.NewInt(52589834579470594), MainnetRewardsSlotFork + 3000},

		// Some after fork tests (new reward algorithm)
		{10 * 100, big.NewInt(100000000000001), 100, big.NewInt(900000000000), big.NewInt(10000000000001), MainnetRewardsSlotFork + 1},
		{123, big.NewInt(345432454323432), 3454, big.NewInt(98779280583), big.NewInt(4248819189750), MainnetRewardsSlotFork + 2},
		{123, big.NewInt(345432454323432), 3454, big.NewInt(98779280583), big.NewInt(4248819189750), MainnetRewardsSlotFork + 3},
		{999, big.NewInt(99999999999999), 5, big.NewInt(18002000000000), big.NewInt(9989999999999), MainnetRewardsSlotFork + 1000},
		{9999, big.NewInt(99999999999999), 99999, big.NewInt(100001), big.NewInt(99990000000000), MainnetRewardsSlotFork + 30000},
	}

	for _, test := range tests {
		oracle := NewOracle(&Config{
			PoolFeesPercentOver10000: test.FeePercentX100,
			PoolFeesAddress:          "0x",
			Network:                  "mainnet",
		})
		for i := 0; i < test.AmountValidators; i++ {
			oracle.addSubscription(uint64(i), "0x", "0x")
		}
		oracle.state.NextSlotToProcess = test.Slot
		oracle.increaseAllPendingRewards(test.Reward)
		for i := 0; i < test.AmountValidators; i++ {
			require.Equal(t, test.ValidatorReward, oracle.state.Validators[uint64(i)].PendingRewardsWei)
		}

		require.Equal(t, test.PoolReward, oracle.state.PoolAccumulatedFees)
	}
}

func Test_ValidatorCleanup_1(t *testing.T) {

	// TODO: This can be improved with some refactor to reduce the boilerplate

	mainnetFork1 := SlotFork1["mainnet"]

	// Test1:
	log.Info("Test1: No validators")
	oracle := NewOracle(&Config{
		Network: "mainnet",
	})
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return nil, nil
	})

	err := oracle.ValidatorCleanup(uint64(mainnetFork1 - 1))
	require.NoError(t, err)

	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)

	// Test2:
	log.Info("Test2: Active and subcribed validator is not modified")
	oracle = NewOracle(&Config{
		Network: "mainnet",
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(765432345), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)

	// Test3:
	log.Info("Test3: Exiting but not exited is not modified")
	oracle = NewOracle(&Config{
		Network: "mainnet",
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateActiveExiting,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(765432345), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)

	// Test4:
	log.Info("Test4: Exited validator rewards are reset and go to the rest")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(76548235), oracle.state.PoolAccumulatedFees)

	// Test5:
	log.Info("Test5: Exited validator rewards are reset and go to the rest (including yellow)")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       YellowCard,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, YellowCard, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(76548235), oracle.state.PoolAccumulatedFees)

	// Test6:
	log.Info("Test6: Slashed validator in the beacon chain. Pending goes to the rest")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       YellowCard,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedSlashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, YellowCard, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(344442055), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(76548235), oracle.state.PoolAccumulatedFees)

	// Test7:
	log.Info("Test7: Exited validator rewards are reset and go to the rest (not red)")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(765432345),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       RedCard,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, RedCard, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(688884111), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(76548234), oracle.state.PoolAccumulatedFees)

	// Test8:
	log.Info("Test8: Multiple events (exit + slash) are handled correctly")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(1000),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(1000),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
			21: {
				Index:  21,
				Status: v1.ValidatorStateExitedSlashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(1800), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, Active, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(200), oracle.state.PoolAccumulatedFees)

	// Test9:
	log.Info("Test9: All existing validators exit")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(1000),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(1000),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.state.Validators[22] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(3000),
		AccumulatedRewardsWei: big.NewInt(99),
		ValidatorStatus:       Active,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
			21: {
				Index:  21,
				Status: v1.ValidatorStateExitedSlashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
			22: {
				Index:  22,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[22].PendingRewardsWei)
	require.Equal(t, big.NewInt(99), oracle.state.Validators[22].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[22].ValidatorStatus)
	require.Equal(t, big.NewInt(5000), oracle.state.PoolAccumulatedFees)

	// Test10:
	log.Info("Test10: Test a banned validator that exists")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(0),
		AccumulatedRewardsWei: big.NewInt(543245),
		ValidatorStatus:       Banned,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(543245), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, Banned, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.PoolAccumulatedFees)

	// Test11:
	log.Info("Test11: Yellow and Red go to NoSubscribed after being slashed/exited")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(10),
		AccumulatedRewardsWei: big.NewInt(20),
		ValidatorStatus:       YellowCard,
	}
	oracle.state.Validators[21] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(30),
		AccumulatedRewardsWei: big.NewInt(40),
		ValidatorStatus:       RedCard,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
			21: {
				Index:  21,
				Status: v1.ValidatorStateExitedSlashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 + 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(20), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[21].PendingRewardsWei)
	require.Equal(t, big.NewInt(40), oracle.state.Validators[21].AccumulatedRewardsWei)
	require.Equal(t, NotSubscribed, oracle.state.Validators[21].ValidatorStatus)
	require.Equal(t, big.NewInt(40), oracle.state.PoolAccumulatedFees)

	// Test12:
	log.Info("Test12: Nothing happens before the fork")
	oracle = NewOracle(&Config{
		Network:                  "mainnet",
		PoolFeesPercentOver10000: 10 * 100,
	})
	oracle.state.Validators[20] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(10),
		AccumulatedRewardsWei: big.NewInt(20),
		ValidatorStatus:       YellowCard,
	}
	oracle.SetGetSetOfValidatorsFunc(func(valIndices []phase0.ValidatorIndex, slot string, opts ...retry.Option) (map[phase0.ValidatorIndex]*v1.Validator, error) {
		return map[phase0.ValidatorIndex]*v1.Validator{
			20: {
				Index:  20,
				Status: v1.ValidatorStateExitedUnslashed,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1},
					PublicKey:             phase0.BLSPubKey{1},
				},
			},
		}, nil
	})
	err = oracle.ValidatorCleanup(uint64(mainnetFork1 - 1))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(10), oracle.state.Validators[20].PendingRewardsWei)
	require.Equal(t, big.NewInt(20), oracle.state.Validators[20].AccumulatedRewardsWei)
	require.Equal(t, YellowCard, oracle.state.Validators[20].ValidatorStatus)
	require.Equal(t, big.NewInt(0), oracle.state.PoolAccumulatedFees)
}

func Test_increaseValidatorPendingRewards(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.state.Validators[12] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(100),
		AccumulatedRewardsWei: big.NewInt(0),
	}
	oracle.state.Validators[200] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(100),
		AccumulatedRewardsWei: big.NewInt(0),
	}

	oracle.increaseValidatorPendingRewards(12, big.NewInt(8765432))
	require.Equal(t, big.NewInt(8765432+100), oracle.state.Validators[12].PendingRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[12].AccumulatedRewardsWei)

	oracle.increaseValidatorPendingRewards(200, big.NewInt(0))
	require.Equal(t, big.NewInt(100), oracle.state.Validators[200].PendingRewardsWei)

	oracle.increaseValidatorPendingRewards(12, big.NewInt(1))
	require.Equal(t, big.NewInt(8765432+100+1), oracle.state.Validators[12].PendingRewardsWei)
}

func Test_increaseValidatorAccumulatedRewards(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.state.Validators[9999999] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(100),
		AccumulatedRewardsWei: big.NewInt(99999999999999),
	}
	oracle.increaseValidatorAccumulatedRewards(9999999, big.NewInt(87676545432))
	require.Equal(t, big.NewInt(87676545432+99999999999999), oracle.state.Validators[9999999].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(100), oracle.state.Validators[9999999].PendingRewardsWei)
}

func Test_sendRewardToPool(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.sendRewardToPool(big.NewInt(10456543212340))
	require.Equal(t, big.NewInt(10456543212340), oracle.state.PoolAccumulatedFees)

	oracle.sendRewardToPool(big.NewInt(99999))
	require.Equal(t, big.NewInt(10456543212340+99999), oracle.state.PoolAccumulatedFees)
}

func Test_resetPendingRewards(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.state.Validators[1] = &ValidatorInfo{
		PendingRewardsWei:     big.NewInt(99999999999999),
		AccumulatedRewardsWei: big.NewInt(99999999999999),
	}
	oracle.resetPendingRewards(1)

	require.Equal(t, big.NewInt(0), oracle.state.Validators[1].PendingRewardsWei)
	require.Equal(t, big.NewInt(99999999999999), oracle.state.Validators[1].AccumulatedRewardsWei)
}

func Test_IncreasePendingRewards(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	oracle.state.Validators[12] = &ValidatorInfo{
		WithdrawalAddress: "0xaa",
		ValidatorStatus:   Active,
		PendingRewardsWei: big.NewInt(100),
	}
	totalAmount := big.NewInt(130)

	require.Equal(t, big.NewInt(100), oracle.state.Validators[12].PendingRewardsWei)
	oracle.increaseAllPendingRewards(totalAmount)
	require.Equal(t, big.NewInt(230), oracle.state.Validators[12].PendingRewardsWei)
}

func Test_IncreasePendingEmptyPool(t *testing.T) {
	// Test a case where a new rewards adds to the pool but no validators are subscribed
	// This can happen when a donation is recived to the pool but no validators are subscribed
	oracle := NewOracle(&Config{Network: "mainnet"})

	// This prevents division by zero
	oracle.increaseAllPendingRewards(big.NewInt(10000))

	// Pool gets all rewards
	require.Equal(t, big.NewInt(10000), oracle.state.PoolAccumulatedFees)
}

func Test_consolidateBalance_Eligible(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	oracle.state.Validators[10] = &ValidatorInfo{
		AccumulatedRewardsWei: big.NewInt(77),
		PendingRewardsWei:     big.NewInt(23),
	}

	require.Equal(t, big.NewInt(77), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(23), oracle.state.Validators[10].PendingRewardsWei)

	oracle.consolidateBalance(10)

	require.Equal(t, big.NewInt(100), oracle.state.Validators[10].AccumulatedRewardsWei)
	require.Equal(t, big.NewInt(0), oracle.state.Validators[10].PendingRewardsWei)
}

func Test_StateMachine(t *testing.T) {
	oracle := NewOracle(&Config{Network: "mainnet"})
	valIndex1 := uint64(1000)
	valIndex2 := uint64(2000)

	type stateTest struct {
		From  ValidatorStatus
		Event Event
		End   ValidatorStatus
	}

	stateMachineTestVector := []stateTest{
		// FromState | Event | EndState
		{Active, ProposalOk, Active},
		{Active, ProposalMissed, YellowCard},
		{Active, ProposalWrongFee, Banned},
		{Active, Unsubscribe, NotSubscribed},

		{YellowCard, ProposalOk, Active},
		{YellowCard, ProposalMissed, RedCard},
		{YellowCard, ProposalWrongFee, Banned},
		{YellowCard, Unsubscribe, NotSubscribed},

		{RedCard, ProposalOk, YellowCard},
		{RedCard, ProposalMissed, RedCard},
		{RedCard, ProposalWrongFee, Banned},
		{RedCard, Unsubscribe, NotSubscribed},

		{NotSubscribed, ManualSubscription, Active},
		{NotSubscribed, AutoSubscription, Active},
	}

	for _, testState := range stateMachineTestVector {
		oracle.state.Validators[valIndex1] = &ValidatorInfo{
			ValidatorStatus: testState.From,
		}
		oracle.state.Validators[valIndex2] = &ValidatorInfo{
			ValidatorStatus: testState.From,
		}

		oracle.advanceStateMachine(valIndex1, testState.Event)
		oracle.advanceStateMachine(valIndex2, testState.Event)

		require.Equal(t, testState.End, oracle.state.Validators[valIndex1].ValidatorStatus)
		require.Equal(t, testState.End, oracle.state.Validators[valIndex2].ValidatorStatus)
	}
}

func Test_IsValidatorSubscribed(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.state.Validators[10] = &ValidatorInfo{
		ValidatorStatus:       Active,
		AccumulatedRewardsWei: big.NewInt(100),
		PendingRewardsWei:     big.NewInt(200),
	}
	oracle.state.Validators[20] = &ValidatorInfo{
		ValidatorStatus:       YellowCard,
		AccumulatedRewardsWei: big.NewInt(300),
		PendingRewardsWei:     big.NewInt(300),
	}
	oracle.state.Validators[30] = &ValidatorInfo{
		ValidatorStatus:       RedCard,
		AccumulatedRewardsWei: big.NewInt(900),
		PendingRewardsWei:     big.NewInt(100),
	}
	oracle.state.Validators[40] = &ValidatorInfo{
		ValidatorStatus:       NotSubscribed,
		AccumulatedRewardsWei: big.NewInt(50),
		PendingRewardsWei:     big.NewInt(10),
	}
	oracle.state.Validators[40] = &ValidatorInfo{
		ValidatorStatus:       Banned,
		AccumulatedRewardsWei: big.NewInt(50),
		PendingRewardsWei:     big.NewInt(10),
	}
	require.Equal(t, true, oracle.isSubscribed(10))
	require.Equal(t, true, oracle.isSubscribed(20))
	require.Equal(t, true, oracle.isSubscribed(30))
	require.Equal(t, false, oracle.isSubscribed(40))
	require.Equal(t, false, oracle.isSubscribed(50))
}

func Test_isBanned(t *testing.T) {
	oracle := NewOracle(&Config{})
	oracle.state.Validators[1] = &ValidatorInfo{
		ValidatorStatus: Active,
	}
	oracle.state.Validators[2] = &ValidatorInfo{
		ValidatorStatus: YellowCard,
	}
	oracle.state.Validators[3] = &ValidatorInfo{
		ValidatorStatus: RedCard,
	}
	oracle.state.Validators[4] = &ValidatorInfo{
		ValidatorStatus: NotSubscribed,
	}
	oracle.state.Validators[5] = &ValidatorInfo{
		ValidatorStatus: Banned,
	}

	require.Equal(t, false, oracle.isBanned(1))
	require.Equal(t, false, oracle.isBanned(2))
	require.Equal(t, false, oracle.isBanned(3))
	require.Equal(t, false, oracle.isBanned(4))
	require.Equal(t, true, oracle.isBanned(5))
}

func Test_CanValidatorSubscribeToPool(t *testing.T) {

	require.Equal(t, CanValidatorSubscribeToPool(&v1.Validator{
		Status: v1.ValidatorStatePendingInitialized,
	}), true)

	require.Equal(t, CanValidatorSubscribeToPool(&v1.Validator{
		Status: v1.ValidatorStatePendingQueued,
	}), true)

	require.Equal(t, CanValidatorSubscribeToPool(&v1.Validator{
		Status: v1.ValidatorStateActiveOngoing,
	}), true)
}

func Test_getMerkleRootIfAny(t *testing.T) {
	oracle := NewOracle(&Config{
		PoolFeesAddress: "0x1123456789abcdef0123456789abcdef01234568",
	})
	oracle.state.LatestProcessedSlot = 100
	oracle.addSubscription(uint64(10), "0x1123456789abcdef0123456789abcdef01234568", "0x1123456789abcdef0123456789abcdef01234568")
	oracle.addSubscription(uint64(11), "0x1123456789abcdef0123456789abcdef01234568", "0x1123456789abcdef0123456789abcdef01234568")
	oracle.addSubscription(uint64(12), "0x1123456789abcdef0123456789abcdef01234568", "0x1123456789abcdef0123456789abcdef01234568")

	root, enough := oracle.getMerkleRootIfAny()
	require.Equal(t, "0x3ba6b7c80fed7f5f5f5796c610c7dc5bbabf408b8525cbcef67086766ab51863", root)
	require.Equal(t, true, enough)
}

func Test_IsCheckpoint(t *testing.T) {
	oracle := NewOracle(&Config{
		DeployedSlot:          7750448,
		CheckPointSizeInSlots: 28800,
	})

	// We are behind the checkpoint
	oracle.state.LatestProcessedSlot = 7750448 + 100
	isCheckpoint, err := oracle.IsCheckpoint()
	require.NoError(t, err)
	require.Equal(t, false, isCheckpoint)

	// We are at the checkpoint
	oracle.state.LatestProcessedSlot = 7750448 + 28800
	isCheckpoint, err = oracle.IsCheckpoint()
	require.NoError(t, err)
	require.Equal(t, true, isCheckpoint)

	// We are at the checkpoint way in the future
	oracle.state.LatestProcessedSlot = 7750448 + 28800*10
	isCheckpoint, err = oracle.IsCheckpoint()
	require.NoError(t, err)
	require.Equal(t, true, isCheckpoint)

	// We are not at the checkpoint but way in the future
	oracle.state.LatestProcessedSlot = 7750448 + 28800*10 + 7
	isCheckpoint, err = oracle.IsCheckpoint()
	require.NoError(t, err)
	require.Equal(t, false, isCheckpoint)

	// Errors if no last processed slot
	oracle.state.LatestProcessedSlot = 0
	isCheckpoint, err = oracle.IsCheckpoint()
	require.Error(t, err)
	require.Equal(t, false, isCheckpoint)
}

func Test_GetWithdrawalAndType(t *testing.T) {
	// Test eth1 credentials
	validator1 := &v1.Validator{
		Validator: &phase0.Validator{
			WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
		},
	}
	with1, type1 := GetWithdrawalAndType(validator1)

	require.Equal(t, with1, "0x9427a30991170f917d7b83def6e44d26577871ed")
	require.Equal(t, type1, Eth1Withdrawal)

	validator2 := &v1.Validator{
		Validator: &phase0.Validator{
			WithdrawalCredentials: []byte{0, 237, 117, 12, 189, 237, 170, 57, 218, 105, 83, 46, 238, 100, 154, 93, 58, 32, 43, 49, 13, 226, 166, 100, 90, 241, 221, 125, 172, 160, 253, 34},
		},
	}
	with2, type2 := GetWithdrawalAndType(validator2)

	require.Equal(t, with2, "0xed750cbdedaa39da69532eee649a5d3a202b310de2a6645af1dd7daca0fd22")
	require.Equal(t, type2, BlsWithdrawal)
}

// Not a test per se but a util to estimate how much memory the oracle will use
// depending on the number of validators and checkpoints
func Test_StateSize(t *testing.T) {
	for i := 0; i < 3; i++ {
		oracle := NewOracle(&Config{
			CollateralInWei: big.NewInt(1000),
			//PoolAddress:     "0x0123456789abcdef0123456789abcdef01234569",
			PoolFeesAddress: "0x0123456789abcdef0123456789abcdef01234568",
		})

		numValidators := 2000
		blockEachType := 5000

		// One year of checkpoints (1 every 3 days)
		numCheckpoints := 130

		// Add validators
		for i := 0; i < numValidators; i++ {
			oracle.addSubscription(uint64(i), fmt.Sprintf("0x%d123456789abcdef0123456789abcdef01234567", i%9), "0x0123456789abcdef0123456789abcdef01234567")
		}

		// Add blocks
		for i := 0; i < blockEachType; i++ {
			oracle.state.ProposedBlocks = append(oracle.state.ProposedBlocks, SummarizedBlock{
				Slot:              uint64(100),
				ValidatorIndex:    uint64(i),
				ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
				Reward:            big.NewInt(5000000000000000000),
				RewardType:        VanilaBlock,
				WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
				BlockType:         OkPoolProposal,
			})
			oracle.state.MissedBlocks = append(oracle.state.MissedBlocks, SummarizedBlock{
				Slot:              uint64(100),
				ValidatorIndex:    uint64(i),
				ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
				Reward:            big.NewInt(5000000000000000000),
				RewardType:        VanilaBlock,
				WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
				BlockType:         MissedProposal,
			})
			oracle.state.WrongFeeBlocks = append(oracle.state.WrongFeeBlocks, SummarizedBlock{
				Slot:              uint64(100),
				ValidatorIndex:    uint64(i),
				ValidatorKey:      "0x0123456789abcdef0123456789abcdef01234567",
				Reward:            big.NewInt(5000000000000000000),
				RewardType:        VanilaBlock,
				WithdrawalAddress: "0x0123456789abcdef0123456789abcdef01234567",
				BlockType:         WrongFeeRecipient,
			})
		}

		for i := 0; i < numCheckpoints; i++ {
			require.Equal(t, true, oracle.FreezeCheckpoint())
		}

		path := filepath.Join(StateFolder, StateJsonName)
		defer os.Remove(path)
		defer os.RemoveAll(StateFolder)
		require.NoError(t, oracle.SaveToJson(false))

		// Get file information
		fileInfo, err := os.Stat(path)
		require.NoError(t, err)

		// Get file size in bytes
		fileSize := fileInfo.Size()
		fileSizeMB := float64(fileSize) / (1024 * 1024)

		// Print the file size
		log.Info("File size:", fileSizeMB, "MB")
	}
}

func checkFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected file does not exist: %s", path)
	}
}
