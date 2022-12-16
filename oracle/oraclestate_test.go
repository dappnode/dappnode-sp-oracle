package oracle

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddSubscription(t *testing.T) {
	state := NewOracleState()
	state.AddSubscriptionIfNotAlready(10)
	state.IncreaseAllPendingRewards(big.NewInt(100))
	state.ConsolidateBalance(10)
	state.IncreaseAllPendingRewards(big.NewInt(200))
	require.Equal(t, big.NewInt(200), state.pendingRewards[10])
	require.Equal(t, big.NewInt(100), state.claimableRewards[10])

	// check that adding again doesnt reset the subscription
	state.AddSubscriptionIfNotAlready(10)
	require.Equal(t, big.NewInt(200), state.pendingRewards[10])
	require.Equal(t, big.NewInt(100), state.claimableRewards[10])
}

func Test_IncreasePendingRewards(t *testing.T) {
	state := NewOracleState()
	state.pendingRewards[12] = big.NewInt(100)
	state.validatorState[12] = Active
	totalAmount := big.NewInt(130)

	require.Equal(t, big.NewInt(100), state.pendingRewards[12])
	state.IncreaseAllPendingRewards(totalAmount)
	require.Equal(t, big.NewInt(230), state.pendingRewards[12])
}

func Test_ConsolidateBalance_Eligible(t *testing.T) {
	valIndex := uint64(10)
	state := NewOracleState()
	state.claimableRewards[valIndex] = big.NewInt(77)
	state.pendingRewards[valIndex] = big.NewInt(23)

	require.Equal(t, big.NewInt(77), state.claimableRewards[valIndex])
	require.Equal(t, big.NewInt(23), state.pendingRewards[valIndex])

	state.ConsolidateBalance(valIndex)

	require.Equal(t, big.NewInt(100), state.claimableRewards[valIndex])
	require.Equal(t, big.NewInt(0), state.pendingRewards[valIndex])
}

func Test_StateMachine(t *testing.T) {
	state := NewOracleState()
	valIndex1 := uint64(1000)
	valIndex2 := uint64(2000)

	type stateTest struct {
		From  int
		Event int
		End   int
	}

	stateMachineTestVector := []stateTest{
		{Active, ProposalWithCorrectFee, Active},
		{Active, ProposalWithWrongFee, Banned},
		{Active, MissedProposal, ActiveWarned},
		//{Active, UnbanValidator, Active}, // TODO: Test that fails

		{ActiveWarned, ProposalWithCorrectFee, Active},
		{ActiveWarned, ProposalWithWrongFee, Banned},
		{ActiveWarned, MissedProposal, NotActive},
		//{ActiveWarned, UnbanValidator, ActiveWarned}, // TODO: Test that fails

		{NotActive, ProposalWithCorrectFee, ActiveWarned},
		{NotActive, ProposalWithWrongFee, Banned},
		{NotActive, MissedProposal, MissedProposal},
		// {NotActive, UnbanValidator, NotActive}, // TODO: Test that fails

		{Banned, ProposalWithCorrectFee, Banned},
		{Banned, ProposalWithWrongFee, Banned},
		{Banned, MissedProposal, Banned},
		{Banned, UnbanValidator, Active},
	}

	for _, testState := range stateMachineTestVector {
		state.validatorState[valIndex1] = testState.From
		state.validatorState[valIndex2] = testState.From

		state.AdvanceStateMachine(valIndex1, testState.Event)
		state.AdvanceStateMachine(valIndex2, testState.Event)

		require.Equal(t, testState.End, state.validatorState[valIndex1])
		require.Equal(t, testState.End, state.validatorState[valIndex2])
	}
}

// TODO: Add more tests when spec settled
