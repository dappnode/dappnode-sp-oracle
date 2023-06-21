package api

import (
	"math/big"
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/dappnode/mev-sp-oracle/contract"
	"github.com/dappnode/mev-sp-oracle/oracle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func Test_ApplyNonFinalizedState_Subscription(t *testing.T) {

	api := NewApiService(&oracle.Config{
		CollateralInWei: big.NewInt(1000),
	}, nil, nil)

	type test struct {
		Collateral          *big.Int
		OracleState         oracle.ValidatorStatus
		UpdatedState        oracle.ValidatorStatus
		EvenValidatorIndex  uint64
		BeaconState         v1.ValidatorState
		ValidatorWithdrawal []byte
		EventSender         common.Address
		BeforePending       *big.Int
		AfterPending        *big.Int
	}

	testVector := []test{
		// Valid subscription
		{big.NewInt(1000), oracle.Untracked, oracle.Active, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(1000)},

		// Valid subscription with prev balance
		{big.NewInt(1000), oracle.Untracked, oracle.Active, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(93565432), big.NewInt(93565432 + 1000)},

		// More collateral than needed
		{big.NewInt(9999999), oracle.Untracked, oracle.Active, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(9999999)},

		// Not enough collateral
		{big.NewInt(1), oracle.Untracked, oracle.Untracked, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},

		// Already subscribed
		{big.NewInt(1000), oracle.Active, oracle.Active, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},

		// Yellow card subscribes again
		{big.NewInt(1000), oracle.YellowCard, oracle.YellowCard, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},

		// Banned subscribes
		{big.NewInt(1000), oracle.Banned, oracle.Banned, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},

		// Unsubscribed subscribes again
		{big.NewInt(1000), oracle.NotSubscribed, oracle.Active, 1, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(1000)},

		// Slashed tries to subcribe
		{big.NewInt(1000), oracle.NotSubscribed, oracle.NotSubscribed, 1, v1.ValidatorStateExitedSlashed, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},

		// Sender is not withdrwal address
		{big.NewInt(1000), oracle.NotSubscribed, oracle.NotSubscribed, 1, v1.ValidatorStateExitedSlashed, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, big.NewInt(0), big.NewInt(0)},

		// Subscription sent for another validator index
		{big.NewInt(1000), oracle.Untracked, oracle.Untracked, 99, v1.ValidatorStateActiveOngoing, []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(0), big.NewInt(0)},
	}

	for _, test := range testVector {
		subs := []oracle.Subscription{
			{
				Event: &contract.ContractSubscribeValidator{
					ValidatorID:            test.EvenValidatorIndex,
					SubscriptionCollateral: test.Collateral,
					Sender:                 test.EventSender,
					Raw:                    types.Log{BlockNumber: 1},
				},
				Validator: &v1.Validator{
					Index:  1,
					Status: test.BeaconState,
					Validator: &phase0.Validator{
						WithdrawalCredentials: test.ValidatorWithdrawal,
					},
				},
			},
		}
		validators := map[uint64]*oracle.ValidatorInfo{
			1: {
				ValidatorStatus:   test.OracleState,
				WithdrawalAddress: hexutil.Encode(test.ValidatorWithdrawal[:]),
				ValidatorIndex:    test.EvenValidatorIndex,
				PendingRewardsWei: test.BeforePending,
			},
		}
		api.ApplyNonFinalizedState(subs, []oracle.Unsubscription{}, validators)
		require.Equal(t, test.UpdatedState, validators[1].ValidatorStatus)
		require.Equal(t, test.AfterPending, validators[1].PendingRewardsWei)
		require.Equal(t, "0x9427a30991170f917d7b83def6e44d26577871ed", validators[1].WithdrawalAddress)
	}
}

func Test_ApplyNonFinalizedState_Unsubscribe(t *testing.T) {
	api := NewApiService(&oracle.Config{
		CollateralInWei: big.NewInt(1000),
	}, nil, nil)

	type test struct {
		OracleState         oracle.ValidatorStatus
		UpdatedState        oracle.ValidatorStatus
		EventValidatorIndex uint64
		EventSender         common.Address
		BeforePending       *big.Int
		AfterPending        *big.Int
	}

	testVector := []test{
		// Valid unsubscription
		{oracle.Active, oracle.NotSubscribed, 500, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(900000), big.NewInt(0)},

		// Unsubscription sent for another validator index
		{oracle.Active, oracle.Active, 1, common.Address{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237}, big.NewInt(50000), big.NewInt(50000)},

		// Unsubscription sent from a wrong with address
		{oracle.Active, oracle.Active, 500, common.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, big.NewInt(676543), big.NewInt(676543)},
	}

	for _, test := range testVector {
		validators := map[uint64]*oracle.ValidatorInfo{
			500: {
				ValidatorStatus:   test.OracleState,
				WithdrawalAddress: "0x9427a30991170f917d7b83def6e44d26577871ed",
				ValidatorIndex:    500,
				PendingRewardsWei: test.BeforePending,
			},
		}

		unsubs := []oracle.Unsubscription{
			{
				Event: &contract.ContractUnsubscribeValidator{
					ValidatorID: test.EventValidatorIndex,
					Sender:      test.EventSender,
					Raw:         types.Log{BlockNumber: 1},
				},
				Validator: &v1.Validator{
					Index: 500,
					Validator: &phase0.Validator{
						WithdrawalCredentials: []byte{148, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
					},
				},
			},
		}
		api.ApplyNonFinalizedState([]oracle.Subscription{}, unsubs, validators)
		require.Equal(t, test.UpdatedState, validators[500].ValidatorStatus)
		require.Equal(t, test.AfterPending, validators[500].PendingRewardsWei)
	}

}

func Test_ApplyNonFinalizedState_MultipleEvents(t *testing.T) {
	api := NewApiService(&oracle.Config{
		CollateralInWei: big.NewInt(1000),
	}, nil, nil)

	validators := map[uint64]*oracle.ValidatorInfo{
		1: {
			ValidatorStatus:   oracle.NotSubscribed,
			WithdrawalAddress: "0x0127a30991170f917d7b83def6e44d26577871ed",
			ValidatorIndex:    1,
			PendingRewardsWei: big.NewInt(0),
		},
		2: {
			ValidatorStatus:   oracle.Untracked,
			WithdrawalAddress: "0x0227a30991170f917d7b83def6e44d26577871ed",
			ValidatorIndex:    2,
			PendingRewardsWei: big.NewInt(0),
		},
		3: {
			ValidatorStatus:   oracle.Untracked,
			WithdrawalAddress: "0x0327a30991170f917d7b83def6e44d26577871ed",
			ValidatorIndex:    3,
			PendingRewardsWei: big.NewInt(0),
		},
	}

	// Apply multiple events to the validator in blocks

	// Subscriptions in block 1000, 2000 and 3000. Unordered
	subs := []oracle.Subscription{
		{
			// Subscribe validator 3
			Event: &contract.ContractSubscribeValidator{
				ValidatorID:            3,
				SubscriptionCollateral: big.NewInt(1000),
				Sender:                 common.Address{3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:                    types.Log{BlockNumber: 5000},
			},
			Validator: &v1.Validator{
				Index:  3,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},

		{
			// Subscribe validator 1
			Event: &contract.ContractSubscribeValidator{
				ValidatorID:            1,
				SubscriptionCollateral: big.NewInt(1000),
				Sender:                 common.Address{1, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:                    types.Log{BlockNumber: 1000},
			},
			Validator: &v1.Validator{
				Index:  1,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},

		{
			// Subscribe validator 2
			Event: &contract.ContractSubscribeValidator{
				ValidatorID:            2,
				SubscriptionCollateral: big.NewInt(1000),
				Sender:                 common.Address{2, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:                    types.Log{BlockNumber: 3000},
			},
			Validator: &v1.Validator{
				Index:  2,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},
	}
	unsubs := []oracle.Unsubscription{
		{
			// Unsubscribe validator 2
			Event: &contract.ContractUnsubscribeValidator{
				ValidatorID: 2,
				Sender:      common.Address{2, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:         types.Log{BlockNumber: 3000},
			},
			Validator: &v1.Validator{
				Index:  2,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},
	}

	api.ApplyNonFinalizedState(subs, unsubs, validators)
	require.Equal(t, oracle.Active, validators[1].ValidatorStatus)
	require.Equal(t, oracle.NotSubscribed, validators[2].ValidatorStatus)
	require.Equal(t, oracle.Active, validators[3].ValidatorStatus)

	// Unsubscribe val 1 and 3, same block
	unsubs = []oracle.Unsubscription{
		{
			Event: &contract.ContractUnsubscribeValidator{
				ValidatorID: 1,
				Sender:      common.Address{1, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:         types.Log{BlockNumber: 99999},
			},
			Validator: &v1.Validator{
				Index:  1,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},
		{
			Event: &contract.ContractUnsubscribeValidator{
				ValidatorID: 3,
				Sender:      common.Address{3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				Raw:         types.Log{BlockNumber: 99999},
			},
			Validator: &v1.Validator{
				Index:  3,
				Status: v1.ValidatorStateActiveOngoing,
				Validator: &phase0.Validator{
					WithdrawalCredentials: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 39, 163, 9, 145, 23, 15, 145, 125, 123, 131, 222, 246, 228, 77, 38, 87, 120, 113, 237},
				},
			},
		},
	}

	api.ApplyNonFinalizedState([]oracle.Subscription{}, unsubs, validators)
	require.Equal(t, oracle.NotSubscribed, validators[1].ValidatorStatus)
	require.Equal(t, oracle.NotSubscribed, validators[2].ValidatorStatus)
	require.Equal(t, oracle.NotSubscribed, validators[3].ValidatorStatus)
}

// Can be used to test the API endpoints, mocking the endpoint
func Test_ApiEndpoint(t *testing.T) {
	/*
		r, _ := http.NewRequest("GET", "/test/abcd", nil)
		w := httptest.NewRecorder()
		vars := map[string]string{
			"mystring": "abcd",
		}
		api := NewApiService(nil, nil, nil)
		r = mux.SetURLVars(r, vars)
		api.handleMemoryValidatorsByWithdrawal(w, r)
		require.Equal(t, 1, 1)
	*/
}

// func TestHandleGetOrderedBlocksByValidatorKey(t *testing.T) {

// 	// Create a mock config
// 	mockConfig := &oracle.Config{
// 		ConsensusEndpoint:     "http://consensus_endpoint",
// 		ExecutionEndpoint:     "http://execution_endpoint",
// 		Network:               "testnet",
// 		PoolAddress:           "pool_address",
// 		DeployedSlot:          1000,
// 		DeployedBlock:         10000,
// 		CheckPointSizeInSlots: 10,
// 		PoolFeesAddress:       "fees_address",
// 		DryRun:                true,
// 		NumRetries:            3,
// 		CollateralInWei:       big.NewInt(1000),
// 		UpdaterKeyPass:        "key_pass",
// 		UpdaterKeyPath:        "key_path",
// 	}

// 	// Create a mock Oracle using the NewOracle function
// 	mockOracle := oracle.NewOracle(mockConfig)
// 	//oracle.NewOracleState(mockConfig)
// 	numBlocks := 100
// 	for i := 0; i < numBlocks; i++ {
// 		// Generate a random ValidatorIndex between 0 and 99
// 		validatorIndex := rand.Intn(100)

// 		// Create mock blocks and add them to the OracleState's ProposedBlocks
// 		mockBlock := blockOkProposal(
// 			rand.Uint64()%101,
// 			uint64(validatorIndex),
// 			"PUBKEY",
// 			big.NewInt(int64(rand.Intn(1000)+10000)),
// 			"0xaaa0000000000000000000000000000000000000",
// 		)
// 		mockOracle.AdvanceStateToNextSlot(&mockBlock)
// 		// Create mock missed blocks and add them to the OracleState's MissedBlocks
// 		mockMissedBlock := MissedBlock(
// 			rand.Uint64()%101,
// 			uint64(validatorIndex),
// 			"PUBKEY",
// 		)
// 		mockOracle.AdvanceStateToNextSlot(&mockMissedBlock)

// 		// Create mock wrong fee blocks and add them to the OracleState's WrongFeeBlocks
// 		mockWrongFeeBlock := WrongFeeBlock(
// 			rand.Uint64()%101,
// 			uint64(validatorIndex),
// 			"PUBKEY",
// 		)
// 		mockOracle.AdvanceStateToNextSlot(&mockWrongFeeBlock)
// 	}
// 	// Create an instance of your ApiService with the mock Oracle
// 	apiService := &ApiService{
// 		oracle: mockOracle,
// 		// other fields initialization...
// 	}
// 	req, err := http.NewRequest("GET", "/memory/validatorblocks", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rr := httptest.NewRecorder()
// 	// Call the handler function directly, passing in the ResponseRecorder and the Request
// 	handler := http.HandlerFunc(apiService.handleMemoryValidatorBlocks)
// 	handler.ServeHTTP(rr, req)
// 	fmt.Println(rr.Body)

// 	// handler2 := http.HandlerFunc(apiService.handleMemoryAllBlocks)
// 	// handler2.ServeHTTP(rr, req)
// 	// fmt.Print(rr.Body.String())

// 	// Perform assertions on the response
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
// 	}
// 	// Perform additional assertions on the response body or headers if needed
// 	// For example, you can check the response body for expected JSON data

// 	// Example assertion for JSON response
// 	// expectedResponse := `{"message":"success"}`
// 	// if rr.Body.String() != expectedResponse {
// 	// 	t.Errorf("handler returned unexpected body: got %v, want %v", rr.Body.String(), expectedResponse)
// 	// }
// }

// func TestHandleMemoryValidatorBlocks(t *testing.T) {

// 	// Create an instance of ApiService
// 	apiService := &ApiService{}

// 	// Create a request and response recorder
// 	req, err := http.NewRequest("GET", "/handle-memory-validator-blocks", nil)
// 	assert.NoError(t, err)
// 	res := httptest.NewRecorder()

// 	// Call the handler function
// 	apiService.handleMemoryValidatorBlocks(res, req)

// 	// Assert the response status code
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	// Assert the response body
// 	expectedJSON := `[{"validator_index":1,"proposed_blocks":[{"slot":0,"block":0,"validator_index":1,"validator_key":"","block_type":0,"reward_wei":null,"reward_type":0,"withdrawal_address":""}],"missed_blocks":[{"slot":0,"block":0,"validator_index":1,"validator_key":"","block_type":0,"reward_wei":null,"reward_type":0,"withdrawal_address":""}]},{"validator_index":2,"proposed_blocks":[{"slot":0,"block":0,"validator_index":2,"validator_key":"","block_type":0,"reward_wei":null,"reward_type":0,"withdrawal_address":""}],"wrong_fee_blocks":[{"slot":0,"block":0,"validator_index":2,"validator_key":"","block_type":0,"reward_wei":null,"reward_type":0,"withdrawal_address":""}]}]`
// 	assert.Equal(t, expectedJSON, res.Body.String())
// }

// func MissedBlock(slot uint64, valIndex uint64, pubKey string) oracle.FullBlock {
// 	return oracle.FullBlock{
// 		Validator: &v1.Validator{
// 			Index: phase0.ValidatorIndex(valIndex),
// 		},
// 	}
// }

// func WrongFeeBlock(slot uint64, valIndex uint64, pubKey string) oracle.FullBlock {
// 	return oracle.FullBlock{
// 		Validator: &v1.Validator{
// 			Index: phase0.ValidatorIndex(valIndex),
// 		},
// 	}
// }

// func blockOkProposal(slot uint64, valIndex uint64, pubKey string, reward *big.Int, withAddress string) oracle.FullBlock {
// 	// return oracle.FullBlock{
// 	// 	Slot:              slot,
// 	// 	ValidatorIndex:    valIndex,
// 	// 	ValidatorKey:      pubKey,
// 	// 	BlockType:         oracle.OkPoolProposal,
// 	// 	Reward:            reward,
// 	// 	RewardType:        oracle.MevBlock,
// 	// 	WithdrawalAddress: withAddress,
// 	// }
// 	return oracle.FullBlock{
// 		Validator: &v1.Validator{
// 			Index: phase0.ValidatorIndex(valIndex),
// 		},
// 	}
// }
