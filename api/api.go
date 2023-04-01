package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/dappnode/mev-sp-oracle/config"
	"github.com/dappnode/mev-sp-oracle/oracle"
	"github.com/dappnode/mev-sp-oracle/postgres"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Note that the api has no paging, so it is not suitable for large queries, but
// it should be able to scale to a few thousand subscribed validators without any problem

// Important: These are the retry options when an api call involves external call to
// the beacon node or execution client. The idea is to try once, and fail fast.
// Use this for all onchain calls, otherwise defaultRetryOpts will be aplied
var apiRetryOpts = []retry.Option{
	retry.Attempts(1),
}

const (
	// Available endpoints
	pathStatus                = "/status"
	pathValidatorRelayers     = "/registeredrelays/{valpubkey}"
	pathDepositAddressByIndex = "/depositaddress/{valindex}"

	// Memory endpoints: what the oracle knows
	pathMemoryValidators          = "/memory/validators"
	pathMemoryValidatorByIndex    = "/memory/validator/{valindex}"
	pathMemoryValidatorsByDeposit = "/memory/validators/{depositaddress}"
	pathMemoryFeesInfo            = "/memory/feesinfo"
	pathMemorySubscriptions       = "/memory/subscriptions"   // TODO
	pathMemoryUnsubscriptions     = "/memory/unsubscriptions" // TODO
	pathMemoryProposedBlocks      = "/memory/proposedblocks"
	pathMemoryMissedBlocks        = "/memory/missedblocks"
	pathMemoryWrongFeeBlocks      = "/memory/wrongfeeblocks"
	pathMemoryDonations           = "/memory/donations"
	pathMemoryPoolStatistics      = "/memory/statistics"

	// Onchain endpoints: what is submitted to the contract
	pathOnchainValidators          = "/onchain/validators"                  // TODO
	pathOnchainValidatorByIndex    = "/onchain/validator/{valindex}"        // TODO
	pathOnchainValidatorsByDeposit = "/onchain/validators/{depositaddress}" // TODO
	pathOnchainFeesInfo            = "/onchain/proof/fees"                  // TODO
	pathOnchainMerkleRoot          = "/onchain/merkleroot"                  // TODO
	pathOnchainMerkleProof         = "/onchain/proof/{depositaddress}"      // TODO
	pathOnchainLatestCheckpoint    = "/onchain/latestcheckpoint"            // TODO
)

type httpErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type httpOkStatus struct {
	IsConsensusInSync         bool   `json:"is_consensus_in_sync"`
	IsExecutionInSync         bool   `json:"is_execution_in_sync"`
	OracleLatestProcessedSlot uint64 `json:"oracle_latest_processed_slot"`
	ChainFinalizedSlot        uint64 `json:"chain_head_slot"`
	OracleHeadDistance        uint64 `json:"oracle_head_distance"`
	ChainId                   string `json:"chainid"`
	DepositContact            string `json:"depositcontract"`
}

type httpOkRelayersState struct {
	Registered           bool     `json:"registered"`
	RegisteredRelayers   []string `json:"registered_relayers"`
	UnRegisteredRelayers []string `json:"unregistered_relayers"`
}

type httpOkDepositAddress struct {
	DepositAddress   string `json:"deposit_address"`
	ValidatorIndex   uint64 `json:"validator_index"`
	ValidatorAddress string `json:"validator_address"`
}

type httpOkLatestCheckpoint struct {
	MerkleRoot     string `json:"merkleroot"`
	CheckpointSlot uint64 `json:"checkpointslot"`
}

type httpOkMerkleRoot struct {
	MerkleRoot string `json:"merkle_root"`
}

type httpOkMemoryStatistics struct {
	TotalSubscribed    uint64 `json:"total_subscribed_validators"`
	TotalActive        uint64 `json:"total_active_validators"`
	TotalYellowCard    uint64 `json:"total_yellowcard_validators"`
	TotalRedCard       uint64 `json:"total_redcard_validators"`
	TotalBanned        uint64 `json:"total_banned_validators"`
	TotalNotSubscribed uint64 `json:"total_notsubscribed_validators"`

	LatestCheckpointSlot       uint64   `json:"latest_checkpoint_slot"`
	NextCheckpointSlot         uint64   `json:"next_checkpoint_slot"`
	TotalAccumulatedRewardsWei *big.Int `json:"total_accumulated_rewards_wei"`
	TotalPendingRewaradsWei    *big.Int `json:"total_pending_rewards_wei"`

	TotalRewardsSentWei *big.Int `json:"total_rewards_sent_wei"`
	TotalDonationsWei   *big.Int `json:"total_donations_wei"`
	AvgBlockRewardWei   *big.Int `json:"avg_block_reward_wei"`

	// TODO: Split Proposed in Vanila/Mev
	//TotalVanilaBlocks   uint64
	//TotalMevBlocks      uint64
	TotalProposedBlocks uint64 `json:"total_proposed_blocks"`
	TotalMissedBlocks   uint64 `json:"total_missed_blocks"`
	TotalWrongFeeBlocks uint64 `json:"total_wrongfee_blocks"`
}

type httpOkValidatorState struct {
	ValidatorStatus       string   `json:"status"`
	AccumulatedRewardsWei *big.Int `json:"accumulated_rewards_wei"`
	PendingRewardsWei     *big.Int `json:"pending_rewards_wei"`
	CollateralWei         *big.Int `json:"collateral_rewards_wei"`
	DepositAddress        string   `json:"deposit_address"`
	ValidatorIndex        uint64   `json:"validator_index"`
	ValidatorKey          string   `json:"validator_key"`
	//ValidatorProposedBlocks   []BlockState
	//ValidatorMissedBlocks     []BlockState
	//ValidatorWrongFeeBlocks   []BlockState

	// TODO: Include ClaimedSoFar from the smart contract for reconciliation
}

type httpOkProofs struct {
	LeafDepositAddress     string   `json:"leaf_deposit_address"`
	LeafAccumulatedBalance *big.Int `json:"leaf_accumulated_balance"`
	MerkleRoot             string   `json:"merkleroot"`
	CheckpointSlot         uint64   `json:"checkpoint_slot"`
	Proofs                 []string `json:"merkle_proofs"`
	RegisteredValidators   []uint64 `json:"registered_validators"`
}

type ApiService struct {
	srv           *http.Server
	Postgres      *postgres.Postgresql
	OracleState   *oracle.OracleState
	Onchain       *oracle.Onchain
	ApiListenAddr string
	Network       string
}

func NewApiService(cfg config.Config, state *oracle.OracleState, onchain *oracle.Onchain) *ApiService {
	postgres, err := postgres.New(cfg.PostgresEndpoint, cfg.NumRetries)
	if err != nil {
		// TODO: Return error instead of fatal
		log.Fatal(err)
	}

	return &ApiService{
		// TODO: configure, add cli flag
		ApiListenAddr: "0.0.0.0:7300",
		Postgres:      postgres,
		OracleState:   state,
		Onchain:       onchain,
		Network:       cfg.Network,
	}
}

func (m *ApiService) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := httpErrorResp{code, message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.WithField("response", resp).WithError(err).Error("Couldn't write error response")
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (m *ApiService) respondOK(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithField("response", response).WithError(err).Error("Couldn't write OK response")
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (m *ApiService) getRouter() http.Handler {
	r := mux.NewRouter()

	// Map endpoints and their handlers
	r.HandleFunc("/", m.handleRoot).Methods(http.MethodGet)

	r.HandleFunc(pathStatus, m.handleStatus).Methods(http.MethodGet)
	r.HandleFunc(pathValidatorRelayers, m.handleValidatorRelayers).Methods(http.MethodGet)
	r.HandleFunc(pathDepositAddressByIndex, m.handleDepositAddressByIndex).Methods(http.MethodGet)

	r.HandleFunc(pathMemoryValidators, m.handleMemoryValidators).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryValidatorByIndex, m.handleMemoryValidatorInfo).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryValidatorsByDeposit, m.handleMemoryValidatorsByDeposit).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryFeesInfo, m.handleMemoryFeesInfo).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryPoolStatistics, m.handleMemoryStatistics).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryProposedBlocks, m.handleMemoryProposedBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryMissedBlocks, m.handleMemoryMissedBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryWrongFeeBlocks, m.handleMemoryWrongFeeBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryDonations, m.handleMemoryDonations).Methods(http.MethodGet)

	//r.HandleFunc(pathLatestMerkleProof, m.handleLatestMerkleProof)
	//r.HandleFunc(pathLatestMerkleRoot, m.handleLatestMerkleRoot)
	//r.HandleFunc(pathLatestCheckpoint, m.handleLatestCheckpoint)

	//r.HandleFunc(pathValidatorOnchainStateByIndex, m.handleValidatorOnchainStateByIndex)
	//r.HandleFunc(pathValidatorOffchainStateByIndex, m.handleValidatorOffchainStateByIndex)
	//r.HandleFunc(pathDonations, m.handleDonations)

	//r.Use(mux.CORSMethodMiddleware(r))

	// TODO: Add logging
	return r
}

func (m *ApiService) StartHTTPServer() error {
	log.Info("Starting HTTP server on ", m.ApiListenAddr)
	if m.srv != nil {
		return errors.New("server already running")
	}

	//go m.startBidCacheCleanupTask()

	m.srv = &http.Server{
		Addr:    m.ApiListenAddr,
		Handler: m.getRouter(),

		//ReadTimeout:       time.Duration(config.ServerReadTimeoutMs) * time.Millisecond,
		//ReadHeaderTimeout: time.Duration(config.ServerReadHeaderTimeoutMs) * time.Millisecond,
		//WriteTimeout:      time.Duration(config.ServerWriteTimeoutMs) * time.Millisecond,
		//IdleTimeout:       time.Duration(config.ServerIdleTimeoutMs) * time.Millisecond,

		//MaxHeaderBytes: config.ServerMaxHeaderBytes,
	}

	err := m.srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (m *ApiService) handleRoot(w http.ResponseWriter, req *http.Request) {
	m.respondOK(w, "see api doc for available endpoints")
}

func (m *ApiService) handleMemoryStatistics(w http.ResponseWriter, req *http.Request) {
	totalSubscribed := uint64(0)
	totalActive := uint64(0)
	totalYellowCard := uint64(0)
	totalRedCard := uint64(0)
	totalBanned := uint64(0)
	totalNotSubscribed := uint64(0)

	totalAccumulatedRewards := big.NewInt(0)
	totalPendingRewards := big.NewInt(0)

	// TODO: Would be nice to divice en MEV and non-MEV blocks
	//totalVanilaBlocks := 0
	//totalMevBlocks := 0

	for _, validator := range m.OracleState.Validators {
		if validator.ValidatorStatus == oracle.Active {
			totalActive++
		} else if validator.ValidatorStatus == oracle.YellowCard {
			totalYellowCard++
		} else if validator.ValidatorStatus == oracle.RedCard {
			totalRedCard++
		} else if validator.ValidatorStatus == oracle.Banned {
			totalBanned++
		} else if validator.ValidatorStatus == oracle.NotSubscribed {
			totalNotSubscribed++
		}
		totalAccumulatedRewards.Add(totalAccumulatedRewards, validator.AccumulatedRewardsWei)
		totalPendingRewards.Add(totalPendingRewards, validator.PendingRewardsWei)
	}

	totalSubscribed = totalActive + totalYellowCard + totalRedCard

	totalRewardsSentWei := big.NewInt(0)
	for _, block := range m.OracleState.ProposedBlocks {
		totalRewardsSentWei.Add(totalRewardsSentWei, block.Reward)
	}
	totalDonationsWei := big.NewInt(0)
	for _, donation := range m.OracleState.Donations {
		totalDonationsWei.Add(totalDonationsWei, donation.AmountWei)
	}

	totalProposedBlocks := uint64(len(m.OracleState.ProposedBlocks))
	avgBlockRewardWei := big.NewInt(0)

	// Avoid division by zero
	if totalProposedBlocks != 0 {
		avgBlockRewardWei = big.NewInt(0).Div(totalRewardsSentWei, big.NewInt(0).SetUint64(uint64(len(m.OracleState.ProposedBlocks))))
	}

	m.respondOK(w, httpOkMemoryStatistics{
		TotalSubscribed:            totalSubscribed,
		TotalActive:                totalActive,
		TotalYellowCard:            totalYellowCard,
		TotalRedCard:               totalRedCard,
		TotalBanned:                totalBanned,
		TotalNotSubscribed:         totalNotSubscribed,
		LatestCheckpointSlot:       m.OracleState.LatestSlot,                                       // This is wrong. TODO: convert date
		NextCheckpointSlot:         m.OracleState.LatestSlot + m.Onchain.Cfg.CheckPointSizeInSlots, // TODO: Also wrong. convert to date
		TotalAccumulatedRewardsWei: totalAccumulatedRewards,
		TotalPendingRewaradsWei:    totalPendingRewards,
		TotalRewardsSentWei:        totalRewardsSentWei,
		TotalDonationsWei:          totalDonationsWei,
		AvgBlockRewardWei:          avgBlockRewardWei,
		TotalProposedBlocks:        totalProposedBlocks,
		TotalMissedBlocks:          uint64(len(m.OracleState.MissedBlocks)),
		TotalWrongFeeBlocks:        uint64(len(m.OracleState.WrongFeeBlocks)),
	})
}

func (m *ApiService) handleStatus(w http.ResponseWriter, req *http.Request) {
	chainId, err := m.Onchain.ExecutionClient.ChainID(context.Background())
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get exex chainid: "+err.Error())
	}

	depositContract, err := m.Onchain.ConsensusClient.DepositContract(context.Background())
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get deposit contract: "+err.Error())
	}

	execSync, err := m.Onchain.ExecutionClient.SyncProgress(context.Background())
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get exec sync progress: "+err.Error())
	}

	// Seems that if nil means its in sync
	execInSync := false
	if execSync == nil {
		execInSync = true
	}

	consSync, err := m.Onchain.ConsensusClient.NodeSyncing(context.Background())
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get consensus sync progress: "+err.Error())
	}

	// Allow some slots to avoid jitter
	consInSync := false
	if uint64(consSync.SyncDistance) < 2 {
		consInSync = true
	}

	finality, err := m.Onchain.ConsensusClient.Finality(context.Background(), "finalized")
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get consensus latest finalized slot: "+err.Error())
	}

	SlotsInEpoch := uint64(32)
	finalizedEpoch := uint64(finality.Finalized.Epoch)
	finalizedSlot := finalizedEpoch * SlotsInEpoch

	status := httpOkStatus{
		IsConsensusInSync:         consInSync,
		IsExecutionInSync:         execInSync,
		OracleLatestProcessedSlot: m.OracleState.LatestSlot,
		ChainFinalizedSlot:        finalizedSlot,
		OracleHeadDistance:        finalizedSlot - m.OracleState.LatestSlot,
		ChainId:                   chainId.String(),
		DepositContact:            depositContract.String(),
	}

	m.respondOK(w, status)
}

func (m *ApiService) handleMemoryValidators(w http.ResponseWriter, req *http.Request) {
	// Perhaps a bit dangerours to access this directly without getters.
	m.respondOK(w, m.OracleState.Validators)
}

func (m *ApiService) handleMemoryValidatorInfo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	valIndexStr := vars["valindex"]
	valIndex, ok := IsValidIndex(valIndexStr)

	if !ok {
		m.respondError(w, http.StatusBadRequest, "invalid validator index: "+valIndexStr)
		return
	}

	validator, found := m.OracleState.Validators[valIndex]
	if !found {
		m.respondError(w, http.StatusBadRequest, fmt.Sprint("could not find validator with index: ", valIndex))
		return
	}

	m.respondOK(w, validator)
}

func (m *ApiService) handleMemoryValidatorsByDeposit(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	depositAddress := vars["depositaddress"]

	if !IsValidAddress(depositAddress) {
		m.respondError(w, http.StatusBadRequest, "invalid depositAddress: "+depositAddress)
		return
	}

	// Use always lowercase
	depositAddress = strings.ToLower(depositAddress)
	validatorsByDeposit := make([]*oracle.ValidatorInfo, 0)

	// Get the validators that have the requested deposit address
	for _, validator := range m.OracleState.Validators {
		if validator.DepositAddress == depositAddress {
			validatorsByDeposit = append(validatorsByDeposit, validator)
		}
	}

	m.respondOK(w, validatorsByDeposit)
}

func (m *ApiService) handleMemoryFeesInfo(w http.ResponseWriter, req *http.Request) {
	type httpOkMemoryFeesInfo struct {
		PoolFeesPercent     int      `json:"pool_fee_percent"`
		PoolFeesAddress     string   `json:"pool_fee_address"`
		PoolAccumulatedFees *big.Int `json:"pool_accumulated_fees"`
	}

	m.respondOK(w, httpOkMemoryFeesInfo{
		PoolFeesPercent:     m.OracleState.PoolFeesPercent,
		PoolFeesAddress:     m.OracleState.PoolFeesAddress,
		PoolAccumulatedFees: m.OracleState.PoolAccumulatedFees,
	})
}

func (m *ApiService) handleMemoryProposedBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO: Use getter, since its safer and dont make this fields public
	m.respondOK(w, m.OracleState.ProposedBlocks)
}

func (m *ApiService) handleMemoryMissedBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO: Use getter, since its safer and dont make this fields public
	m.respondOK(w, m.OracleState.MissedBlocks)
}

func (m *ApiService) handleMemoryWrongFeeBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO: Use getter, since its safer and dont make this fields public
	m.respondOK(w, m.OracleState.WrongFeeBlocks)
}

func (m *ApiService) handleMemoryDonations(w http.ResponseWriter, req *http.Request) {
	// TODO: Use getter, since its safer and dont make this fields public
	m.respondOK(w, m.OracleState.Donations)
}

func (m *ApiService) handleLatestMerkleProof(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	depositAddress := vars["depositaddress"]

	if !IsValidAddress(depositAddress) {
		m.respondError(w, http.StatusBadRequest, "invalid depositAddress: "+depositAddress)
		return
	}

	// Use always lowercase
	depositAddress = strings.ToLower(depositAddress)

	// Get the proofs of this deposit address (to be used onchain to claim rewards)
	proofs, proofFound := m.OracleState.LatestCommitedState.Proofs[depositAddress]
	if !proofFound {
		m.respondError(w, http.StatusBadRequest, "could not find proof for depositAddress: "+depositAddress)
		return
	}

	// Get the leafs of this deposit address (to be used onchain to claim rewards)
	leafs, leafsFound := m.OracleState.LatestCommitedState.Leafs[depositAddress]
	if !leafsFound {
		m.respondError(w, http.StatusBadRequest, "could not find leafs for depositAddress: "+depositAddress)
		return
	}

	// Get validators that are registered to this deposit address in the pool
	registeredValidators := make([]uint64, 0)
	for valIndex, validator := range m.OracleState.LatestCommitedState.Validators {
		if validator.DepositAddress == depositAddress {
			registeredValidators = append(registeredValidators, valIndex)
		}
	}

	m.respondOK(w, httpOkProofs{
		LeafDepositAddress:     leafs.DepositAddress,
		LeafAccumulatedBalance: leafs.AccumulatedBalance,
		MerkleRoot:             m.OracleState.LatestCommitedState.MerkleRoot,
		CheckpointSlot:         m.OracleState.LatestCommitedState.Slot,
		Proofs:                 proofs,
		RegisteredValidators:   registeredValidators,
	})
}

func (m *ApiService) handleLatestMerkleRoot(w http.ResponseWriter, req *http.Request) {
	// This is the latest merkle root tracked from the oracle.
	//oracleMerkleRoot := "0x" + m.OracleState.LatestCommitedState.MerkleRoot

	contractMerkleRoot, err := m.Onchain.GetContractMerkleRoot(apiRetryOpts...)
	if err != nil {
		m.respondError(w, http.StatusBadRequest, "could not get latest merkle root from chain")
		return
	}
	m.respondOK(w, httpOkMerkleRoot{
		MerkleRoot: contractMerkleRoot,
	})
}

func (m *ApiService) handleDepositAddressByIndex(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	valIndex, err := strconv.ParseUint(vars["valindex"], 10, 64)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not parse valIndex: "+err.Error())
		return
	}

	valInfo, err := m.Onchain.ConsensusClient.Validators(context.Background(), "finalized", []phase0.ValidatorIndex{phase0.ValidatorIndex(valIndex)})
	valPubKeyByte := valInfo[phase0.ValidatorIndex(valIndex)].Validator.PublicKey
	valPubKeyStr := "0x" + hex.EncodeToString(valPubKeyByte[:])

	depositAddress, err := m.Postgres.GetDepositAddressOfValidatorKey(valPubKeyStr)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get deposit address for valindex: "+err.Error())
		return
	}

	m.respondOK(w, httpOkDepositAddress{
		DepositAddress:   depositAddress,
		ValidatorIndex:   valIndex,
		ValidatorAddress: valPubKeyStr,
	})
}

func (m *ApiService) handleValidatorOnchainStateByIndex(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	valIndex, err := strconv.ParseUint(vars["valindex"], 10, 64)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not parse valIndex: "+err.Error())
		return
	}

	// We look into the LatestCommitedState, since its whats its onchain
	valState, found := m.OracleState.LatestCommitedState.Validators[uint64(valIndex)]
	if !found {
		m.respondError(w, http.StatusInternalServerError, fmt.Sprintf("validator index not tracked in the oracle: %d", valIndex))
		return
	}
	m.respondOK(w, httpOkValidatorState{
		ValidatorStatus:       oracle.ValidatorStateToString(valState.ValidatorStatus),
		AccumulatedRewardsWei: valState.AccumulatedRewardsWei,
		PendingRewardsWei:     valState.PendingRewardsWei,
		CollateralWei:         valState.CollateralWei,
		DepositAddress:        valState.DepositAddress,
		ValidatorIndex:        valState.ValidatorIndex,
		ValidatorKey:          valState.ValidatorKey,
		// TODO: Missing blocks fields
	})
}

func (m *ApiService) handleValidatorRelayers(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	valPubKey := vars["valpubkey"]
	if !IsValidPubkey(valPubKey) {
		m.respondError(w, http.StatusInternalServerError, fmt.Sprintf("invalid validator pubkey format"))
		return
	}
	var registeredRelays []string
	var unregisteredRelays []string
	registered := false
	var relays []string

	if m.Network == "mainnet" {
		relays = config.MainRelays

	} else {
		relays = config.TestRelays
	}

	for _, relay := range relays {
		url := fmt.Sprintf("https://%s/relay/v1/data/validator_registration?pubkey=%s", relay, valPubKey)
		resp, err := http.Get(url)
		if err != nil {
			m.respondError(w, http.StatusInternalServerError, "could not call relayer endpoint: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			registeredRelays = append(registeredRelays, relay)
		} else {
			unregisteredRelays = append(unregisteredRelays, relay)
		}
	}

	if len(registeredRelays) != 0 {
		registered = true
	}

	m.respondOK(w, httpOkRelayersState{
		Registered:           registered,
		RegisteredRelayers:   registeredRelays,
		UnRegisteredRelayers: unregisteredRelays,
	})
}

func IsValidIndex(v string) (uint64, bool) {
	//re := regexp.MustCompile("^[0-9]+$")
	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

func IsValidAddress(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(v)
}

func IsValidPubkey(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-f]{96}$")
	return re.MatchString(v)
}
