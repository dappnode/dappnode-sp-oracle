package oracle

import (
	"encoding/hex"
	"mev-sp-oracle/config" // TODO: Change when pushed "github.com/dappnode/mev-sp-oracle/config"
	"mev-sp-oracle/contract"
	"mev-sp-oracle/postgres"

	log "github.com/sirupsen/logrus"
)

type Oracle struct {
	fetcher    *Fetcher
	cfg        *config.Config
	State      *OracleState
	Operations *contract.Operations
	Postgres   *postgres.Postgresql
}

func NewOracle(
	cfg *config.Config,
	fetcher *Fetcher) *Oracle {
	state := NewOracleState(cfg)

	postgres, err := postgres.New(cfg.PostgresEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	operations := contract.NewOperations(cfg)
	oracle := &Oracle{
		cfg:        cfg,
		fetcher:    fetcher,
		State:      state,
		Postgres:   postgres,
		Operations: operations,
	}

	return oracle
}

// Returns the validator index that should propose the block at a given slot, followed
// by wether the block was missed or not (true = ok proposal) and the block if it was not missed
func (or *Oracle) GetBlockIfAny(slot uint64) (uint64, string, bool, *VersionedSignedBeaconBlock) {
	// Gets the duty, indicating which validator should propose the block at this slot
	slotDuty, err := or.fetcher.GetProposalDuty(slot)
	if err != nil {
		log.Fatal("could not get proposal duty: ", err)
	}

	// The validator that should propose the block
	valIndexDuty := uint64(slotDuty.ValidatorIndex)
	valPublicKey := "0x" + hex.EncodeToString(slotDuty.PubKey[:])

	proposedBlock, err := or.fetcher.GetBlockAtSlot(slot)
	if err != nil {
		log.Fatal("could not get block at slot:", err)
	}

	// A nil block means that the validator did not propose a block (missed proposal)
	if proposedBlock == nil {
		return valIndexDuty, valPublicKey, false, nil
	}
	return valIndexDuty, valPublicKey, true, &VersionedSignedBeaconBlock{proposedBlock}
}

func (or *Oracle) UpdateSubscriptions(block VersionedSignedBeaconBlock) {
	// TODO: Listen events from the smart contract
	// Detect manual subscriptions
	/*
		proposerIndex := uint64(block.GetProposerIndex())

		reward, correctFeeRec, rewardType, err := block.GetSentRewardAndType(or.cfg.PoolAddress, *or.fetcher)
		if err != nil {
			log.Fatal(err)
		}*/
}

// TODO: Remove the slot from the input, makes no sense
func (or *Oracle) GetDepositAddressOfValidator(validatorPubKey string, slot uint64) string {
	depositAddress, err := or.Postgres.GetDepositAddressOfValidatorKey(validatorPubKey)
	if err == nil {
		return depositAddress
	}
	log.Warn("Deposit key not found for ", validatorPubKey, ". Expected in goerli. Using a default one. err: ", err)

	// TODO: Remove this in production. Used in goerli for testing with differenet addresses
	someDepositAddresses := []string{
		"0x001eDa52592fE2f8a28dA25E8033C263744b1b6E",
		"0x0029a125E6A3f058628Bd619C91f481e4470D673",
		"0x003718fb88964A1F167eCf205c7f04B25FF46B8E",
		"0x004b1EaBc3ea60331a01fFfC3D63E5F6B3aB88B3",
		"0x005CD1608e40d1e775a97d12e4f594029567C071",
		"0x0069c9017BDd6753467c138449eF98320be1a4E4",
		"0x007cF0936ACa64Ef22C0019A616801Bec7FCCECF",
	}
	//Just pick a "random" one to not always the same
	return someDepositAddresses[slot%7]
}

func (or *Oracle) AdvanceStateToNextSlot() error {
	// TODO: Get block and listen for new subscriptions

	// TODO: Ensure somehow that we dont process a slot twice.
	slotToProcess := or.State.LatestSlot

	// Get the block if any and who proposed it (or should have proposed it)
	proposerIndex, proposerKey, proposedOk, block := or.GetBlockIfAny(slotToProcess)

	// TODO: Update subscriptions with the info from this block (fee rec) + listening to the smart contract
	// this also updates the deposit address and all the information of the validator.

	if proposedOk {
		// If the block was proposed ok
		reward, correctFeeRec, rewardType, err := block.GetSentRewardAndType(or.cfg.PoolAddress, *or.fetcher)
		_ = rewardType
		if err != nil {
			log.Fatal(err)
		}

		// Manual subscription. If feeRec is ok, means the reward was sent to the pool
		if correctFeeRec {
			depositAddress := or.GetDepositAddressOfValidator(proposerKey, slotToProcess)
			or.State.AddSubscriptionIfNotAlready(proposerIndex, depositAddress, proposerKey)
			or.State.AdvanceStateMachine(proposerIndex, ProposalWithCorrectFee)
			or.State.IncreaseAllPendingRewards(reward)
			or.State.ConsolidateBalance(proposerIndex)
			or.State.Validators[proposerIndex].ProposedBlocksSlots = append(or.State.Validators[proposerIndex].ProposedBlocksSlots, slotToProcess)
		}
		// If the validator was subscribed but the fee recipient was wrong
		// we ban the validator as it is not following the protocol rules
		if !correctFeeRec && or.State.IsValidatorSubscribed(proposerIndex) {
			or.State.AdvanceStateMachine(proposerIndex, ProposalWithWrongFee)
			or.State.IncreaseAllPendingRewards(or.State.Validators[proposerIndex].PendingRewardsWei)
			or.State.ResetPendingRewards(proposerIndex)
			or.State.Validators[proposerIndex].WrongFeeBlocksSlots = append(or.State.Validators[proposerIndex].WrongFeeBlocksSlots, slotToProcess)
		}
	}

	// If the validator was not subscribed and missed proposed the block in this slot
	if !proposedOk && or.State.IsValidatorSubscribed(proposerIndex) {
		// If the validator missed a block, just advance the state machine
		// there are no rewards to share, but validator state slighly changes
		//or.State.AdvanceStateMachine(proposerIndex, MissedBlock)
		or.State.Validators[proposerIndex].MissedBlocksSlots = append(or.State.Validators[proposerIndex].MissedBlocksSlots, slotToProcess)
	}

	// TODO: Enable this back.
	/*
		rewardTypeString := "unset"
		if rewardType == VanilaBlock {
			rewardTypeString = "vanila"
		} else if rewardType == MevBlock {
			rewardTypeString = "mev"
		}
		err = or.Postgres.StoreBlockInDb(
			"TODO",
			uint64(slotDuty.Slot),
			pubKey,
			valIndexDuty,
			rewardTypeString,
			*reward,
			1, // TODO
		)
		if err != nil {
			log.Fatal(err)
		}*/

	or.State.LatestSlot = slotToProcess + 1
	return nil
}
