package oracle

import (
	"encoding/json"
	"math/big"

	api "github.com/attestantio/go-eth2-client/api/v1"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/dappnode/mev-sp-oracle/contract"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

const DefaultRoot = "0x0000000000000000000000000000000000000000000000000000000000000000"
const DefaultAddress = "0x0000000000000000000000000000000000000000"

// Types of block rewards
type RewardType uint8

const (
	UnknownRewardType RewardType = 0
	VanilaBlock       RewardType = 1
	MevBlock          RewardType = 2
)

// States of the state machine
type ValidatorStatus uint8

const (
	UnknownState  ValidatorStatus = 0
	Active        ValidatorStatus = 1
	YellowCard    ValidatorStatus = 2
	RedCard       ValidatorStatus = 3
	NotSubscribed ValidatorStatus = 4
	Banned        ValidatorStatus = 5
	Untracked     ValidatorStatus = 6
)

// Events in the state machine that trigger transition
type Event uint8

const (
	UnknownEvent       Event = 0
	ProposalOk         Event = 1
	ProposalMissed     Event = 2
	ProposalWrongFee   Event = 3
	ManualSubscription Event = 4
	AutoSubscription   Event = 5
	Unsubscribe        Event = 6
	ManualBan          Event = 7
	ManualUnban        Event = 8
)

// Block type
type BlockType uint8

const (
	UnknownBlockType      BlockType = 0
	MissedProposal        BlockType = 1
	WrongFeeRecipient     BlockType = 2
	OkPoolProposal        BlockType = 3
	OkPoolProposalBlsKeys BlockType = 4
)

// Withdrawal credentials type
type WithdrawalType uint8

const (
	BlsWithdrawal  WithdrawalType = 0
	Eth1Withdrawal WithdrawalType = 1
)

// Type of validator subscription
type SubscriptionType uint8

const (
	Manual SubscriptionType = 0
	Auto   SubscriptionType = 1
)

type Config struct {
	ConsensusEndpoint        string   `json:"consensus_endpoint"`
	ExecutionEndpoint        string   `json:"execution_endpoint"`
	Network                  string   `json:"network"`
	PoolAddress              string   `json:"pool_address"`
	DeployedSlot             uint64   `json:"deployed_slot"`
	DeployedBlock            uint64   `json:"deployed_block"`
	CheckPointSizeInSlots    uint64   `json:"checkpoint_size"`
	PoolFeesPercentOver10000 int      `json:"pool_fees_percent"` // With 2 decimals (eg 1.5% = 150)
	PoolFeesAddress          string   `json:"pool_fees_address"`
	DryRun                   bool     `json:"dry_run"`
	NumRetries               int      `json:"num_retries"`
	CollateralInWei          *big.Int `json:"collateral_in_wei"`
	UpdaterKeyPass           string   `json:"-"`
	UpdaterKeyFile           string   `json:"-"`
}

// All the events that the contract can emit
type Events struct {
	EtherReceived                []*contract.ContractEtherReceived                `json:"ether_received_events"`
	SubscribeValidator           []*contract.ContractSubscribeValidator           `json:"subscribe_validator_events"`
	ClaimRewards                 []*contract.ContractClaimRewards                 `json:"claim_rewards_events"`
	SetRewardRecipient           []*contract.ContractSetRewardRecipient           `json:"set_reward_recipient_events"`
	UnsubscribeValidator         []*contract.ContractUnsubscribeValidator         `json:"unsubscribe_validator_events"`
	InitSmoothingPool            []*contract.ContractInitSmoothingPool            `json:"init_smoothing_pool_events"`
	UpdatePoolFee                []*contract.ContractUpdatePoolFee                `json:"update_pool_fee_events"`
	PoolFeeRecipient             []*contract.ContractUpdatePoolFeeRecipient       `json:"pool_fee_recipient_events"`
	CheckpointSlotSize           []*contract.ContractUpdateCheckpointSlotSize     `json:"checkpoint_slot_size_events"`
	UpdateSubscriptionCollateral []*contract.ContractUpdateSubscriptionCollateral `json:"update_subscription_collateral_events"`
	SubmitReport                 []*contract.ContractSubmitReport                 `json:"submit_report_events"`
	ReportConsolidated           []*contract.ContractReportConsolidated           `json:"report_consolidated_events"`
	UpdateQuorum                 []*contract.ContractUpdateQuorum                 `json:"update_quorum_events"`
	AddOracleMember              []*contract.ContractAddOracleMember              `json:"add_oracle_member_events"`
	RemoveOracleMember           []*contract.ContractRemoveOracleMember           `json:"remove_oracle_member_events"`
	TransferGovernance           []*contract.ContractTransferGovernance           `json:"transfer_governance_events"`
	AcceptGovernance             []*contract.ContractAcceptGovernance             `json:"accept_governance_events"`
	BanValidator                 []*contract.ContractBanValidator                 `json:"ban_validator_events"`
	UnbanValidator               []*contract.ContractUnbanValidator               `json:"unban_validator_events"`
}

// Information of every block from the blockchain. Some fields are optional
// eg: if the block is not relevant to the pool
type FullBlock struct {

	// consensus data: duty (mandatory, who should propose the block)
	ConsensusDuty *api.ProposerDuty `json:"consensus_duty"`

	// consensus data: validator (mandatory, who should propose the block)
	Validator *v1.Validator `json:"validator"`

	// consensus data: block (optional, only when not missed)
	ConsensusBlock *spec.VersionedSignedBeaconBlock `json:"consensus_block"`

	// execution data: txs (optional, only when interested in vanila reward)
	ExecutionHeader   *types.Header    `json:"execution_header"`
	ExecutionReceipts []*types.Receipt `json:"execution_receipts"`

	// execution data: events (optional, only when the block was not missed)
	Events *Events `json:"events"`

	// Populated with the validators if there are sub/unsub events
	ValidatorsSubs   []*v1.Validator `json:"validators_subs"`
	ValidatorsUnsubs []*v1.Validator `json:"validators_unsubs"`

	ChainId uint64 `json:"chain_id"`
}

// Represents a block with information relevant for the pool, uses Fullblock
// but stores a subset of the fields (summarized). Otherwise storing everything
// in memory may be too much
type SummarizedBlock struct {
	Slot              uint64     `json:"slot"`
	Block             uint64     `json:"block"`
	ValidatorIndex    uint64     `json:"validator_index"`
	ValidatorKey      string     `json:"validator_key"`
	BlockType         BlockType  `json:"block_type"`
	Reward            *big.Int   `json:"reward_wei"`
	RewardType        RewardType `json:"reward_type"`
	WithdrawalAddress string     `json:"withdrawal_address"`
}

// Represents all the information that is stored of a validator
type ValidatorInfo struct {
	ValidatorStatus       ValidatorStatus  `json:"status"`
	AccumulatedRewardsWei *big.Int         `json:"accumulated_rewards_wei"`
	PendingRewardsWei     *big.Int         `json:"pending_rewards_wei"`
	CollateralWei         *big.Int         `json:"collateral_wei"`
	WithdrawalAddress     string           `json:"withdrawal_address"`
	ValidatorIndex        uint64           `json:"validator_index"`
	ValidatorKey          string           `json:"validator_key"`
	SubscriptionType      SubscriptionType `json:"subscription_type"`
}

// Represents the latest commited state onchain
type OnchainState struct {
	Slot       uint64                    `json:"slot"`
	TxHash     string                    `json:"tx_hash"`
	MerkleRoot string                    `json:"merkle_root"`
	Validators map[uint64]*ValidatorInfo `json:"validators"`
	Leafs      map[string]RawLeaf        `json:"leafs"`
	Proofs     map[string][]string       `json:"proofs"`
}

type OracleState struct {
	StateHash            string   `json:"state_hash"`
	LatestProcessedSlot  uint64   `json:"latest_processed_slot"`
	LatestProcessedBlock uint64   `json:"latest_processed_block"`
	NextSlotToProcess    uint64   `json:"next_slot_to_process"`
	PoolAccumulatedFees  *big.Int `json:"pool_accumulated_fees"`

	// Stores the latest state of the validators at LatestProcessedSlot
	Validators map[uint64]*ValidatorInfo `json:"validators"`

	// Contains frozen checkpoints of the validator state on different slots
	CommitedStates map[uint64]*OnchainState `json:"commited_states"`

	// Stores all events both valid and invalid
	SubscriptionEvents   []*contract.ContractSubscribeValidator   `json:"subscriptions_events"`
	UnsubscriptionEvents []*contract.ContractUnsubscribeValidator `json:"unsubscriptions_events"`

	// Stored all EtherReceived events (MEV rewards and donations)
	EtherReceivedEvents []*contract.ContractEtherReceived `json:"ether_received_events"`

	// Stores a subset of EtherReceived events, just donations
	Donations []*contract.ContractEtherReceived `json:"donations"`

	// Summarized versions of the blocks
	ProposedBlocks []SummarizedBlock `json:"proposed_blocks"`
	MissedBlocks   []SummarizedBlock `json:"missed_blocks"`
	WrongFeeBlocks []SummarizedBlock `json:"wrong_fee_blocks"`

	// Config parameters
	PoolFeesPercentOver10000 int      `json:"pool_fees_percent_over_10000"`
	PoolAddress              string   `json:"pool_address"`
	Network                  string   `json:"network"`
	PoolFeesAddress          string   `json:"pool_fees_address"`
	CheckPointSizeInSlots    uint64   `json:"check_point_size_in_slots"`
	DeployedBlock            uint64   `json:"deployed_block"`
	DeployedSlot             uint64   `json:"deployed_slot"`
	CollateralInWei          *big.Int `json:"collateral_in_wei"`
}

type RawLeaf struct {
	WithdrawalAddress     string   `json:"withdrawal_address"`
	AccumulatedBalanceWei *big.Int `json:"accumulated_balance_wei"`
}

func (r *RewardType) String() string {
	if *r == VanilaBlock {
		return "vanila"
	} else if *r == MevBlock {
		return "mev"
	} else if *r == UnknownRewardType {
		return "unknownrewardtype"
	}
	return ""
}

func (s *RewardType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *RewardType) UnmarshalJSON(b []byte) error {
	var rewardtype string
	if err := json.Unmarshal(b, &rewardtype); err != nil {
		return errors.Wrap(err, "unmarshaling reward type")
	}

	if rewardtype == "vanila" {
		*s = VanilaBlock
	} else if rewardtype == "mev" {
		*s = MevBlock
	} else if rewardtype == "unknownrewardtype" {
		*s = UnknownRewardType
	} else {
		return errors.New("unknown reward type")
	}
	return nil
}

func (v ValidatorStatus) String() string {
	if v == Active {
		return "active"
	} else if v == YellowCard {
		return "yellowcard"
	} else if v == RedCard {
		return "redcard"
	} else if v == NotSubscribed {
		return "notsubscribed"
	} else if v == Banned {
		return "banned"
	} else if v == Untracked {
		return "untracked"
	} else if v == UnknownState {
		return "unknownstate"
	}
	return ""
}
func (s *ValidatorStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *ValidatorStatus) UnmarshalJSON(b []byte) error {
	var status string
	if err := json.Unmarshal(b, &status); err != nil {
		return errors.Wrap(err, "unmarshaling validator status")
	}

	if status == "active" {
		*s = Active
	} else if status == "yellowcard" {
		*s = YellowCard
	} else if status == "redcard" {
		*s = RedCard
	} else if status == "notsubscribed" {
		*s = NotSubscribed
	} else if status == "banned" {
		*s = Banned
	} else if status == "untracked" {
		*s = Untracked
	} else if status == "unknownstate" {
		*s = UnknownState
	} else {
		return errors.New("unknown validator status")
	}
	return nil
}

func (b *BlockType) String() string {
	if *b == MissedProposal {
		return "missedproposal"
	} else if *b == WrongFeeRecipient {
		return "wrongfeerecipient"
	} else if *b == OkPoolProposal {
		return "okpoolproposal"
	} else if *b == OkPoolProposalBlsKeys {
		return "okpoolproposalblskeys"
	} else if *b == UnknownBlockType {
		return "unknownblocktype"
	}
	return ""
}

func (s *BlockType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *BlockType) UnmarshalJSON(b []byte) error {
	var blocktype string
	if err := json.Unmarshal(b, &blocktype); err != nil {
		return errors.Wrap(err, "unmarshaling block type")
	}

	if blocktype == "missedproposal" {
		*s = MissedProposal
	} else if blocktype == "wrongfeerecipient" {
		*s = WrongFeeRecipient
	} else if blocktype == "okpoolproposal" {
		*s = OkPoolProposal
	} else if blocktype == "okpoolproposalblskeys" {
		*s = OkPoolProposalBlsKeys
	} else if blocktype == "unknownblocktype" {
		*s = UnknownBlockType
	} else {
		return errors.New("unknown block type")
	}
	return nil
}

func (b *SubscriptionType) String() string {
	if *b == Auto {
		return "auto"
	} else if *b == Manual {
		return "manual"
	}
	return ""
}

func (s *SubscriptionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *SubscriptionType) UnmarshalJSON(b []byte) error {
	var subType string
	if err := json.Unmarshal(b, &subType); err != nil {
		return errors.Wrap(err, "unmarshaling subscription type")
	}

	if subType == "auto" {
		*s = Auto
	} else if subType == "manual" {
		*s = Manual
	} else {
		return errors.New("unknown subscription type")
	}
	return nil
}
