package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	builderApiV1 "github.com/attestantio/go-builder-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/avast/retry-go/v4"
	"github.com/dappnode/mev-sp-oracle/config"
	"github.com/dappnode/mev-sp-oracle/constants"
	"github.com/dappnode/mev-sp-oracle/metrics"
	"github.com/dappnode/mev-sp-oracle/oracle"
	"github.com/dappnode/mev-sp-oracle/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Oracle does not serve some endpoint if not in sync to latest finalized epoch. Some
// slots behind are allowed, since its normal that when a new epoch is finalized, some
// slots are still pending to be processed. This is the max number of slots allowed
var MaxSlotsBehind = uint64(64)

// Note that the api has no paging, so it is not suitable for large queries, but
// it should be able to scale to a few thousand subscribed validators without any problem

// Important: These are the retry options when an api call involves external call to
// the beacon node or execution client. The idea is to try once, and fail fast.
// Use this for all onchain calls, otherwise defaultRetryOpts will be aplied
var apiRetryOpts = []retry.Option{
	retry.Attempts(1),
}

const defaultMerkleRoot = "0x0000000000000000000000000000000000000000000000000000000000000000"

// 30 days/month * 24 hours/day * 3600 seconds/day / 12 seconds/slot
var SlotsInOneMonth = uint64(216000)

const (
	// Available endpoints
	pathStatus            = "/status"
	pathConfig            = "/config"
	pathValidatorRelayers = "/registeredrelays/{valpubkey}"
	pathState             = "/state"

	// Memory endpoints: what the oracle knows
	pathMemoryValidators             = "/memory/validators"
	pathMemoryValidatorByIndex       = "/memory/validator/{valindex}"
	pathMemoryValidatorsByWithdrawal = "/memory/validators/{withdrawalAddress}"
	pathMemoryFeesInfo               = "/memory/feesinfo"
	pathMemoryAllBlocks              = "/memory/allblocks"
	pathMemoryProposedBlocks         = "/memory/proposedblocks"
	pathMemoryMissedBlocks           = "/memory/missedblocks"
	pathMemoryWrongFeeBlocks         = "/memory/wrongfeeblocks"
	pathMemoryDonations              = "/memory/donations"
	pathMemoryPoolStatistics         = "/memory/statistics"

	// Onchain endpoints: what is submitted to the contract
	pathOnchainMerkleProof = "/onchain/proof/{withdrawalAddress}"
)

type ApiService struct {
	srv           *http.Server
	cfg           *oracle.Config
	cliCfg        *config.CliConfig
	Onchain       *oracle.Onchain
	oracle        *oracle.Oracle
	ApiListenAddr string
	Network       string
}

func NewApiService(
	cfg *oracle.Config,
	cliCfg *config.CliConfig,
	oracle *oracle.Oracle,
	onchain *oracle.Onchain) *ApiService {

	return &ApiService{
		ApiListenAddr: fmt.Sprintf("0.0.0.0:%d", cliCfg.ApiPort),
		cfg:           cfg,
		cliCfg:        cliCfg,
		oracle:        oracle,
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

type responseWriterDelegator struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func sanitizeMethod(m string) string {
	return strings.ToLower(m)
}

func sanitizeCode(s int) string {
	return strconv.Itoa(s)
}

// Prometheus middleware to track http requests count and latency. Inspired by
// https://github.com/albertogviana/prometheus-middleware
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()

		delegate := &responseWriterDelegator{ResponseWriter: w}
		rw := delegate

		next.ServeHTTP(rw, r)

		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		code := sanitizeCode(delegate.status)
		method := sanitizeMethod(r.Method)

		go metrics.HttpRequestsTotal.WithLabelValues(
			code,
			method,
			path,
		).Inc()

		go metrics.HttpRequestsLatency.WithLabelValues(
			code,
			method,
			path,
		).Observe(float64(time.Since(begin)) / float64(time.Second))
	})
}

func (m *ApiService) getRouter() http.Handler {
	r := mux.NewRouter()

	// Map endpoints and their handlers
	r.HandleFunc("/", m.handleRoot).Methods(http.MethodGet)

	// General endpoints
	r.HandleFunc(pathStatus, m.handleStatus).Methods(http.MethodGet)
	r.HandleFunc(pathConfig, m.handleConfig).Methods(http.MethodGet)
	r.HandleFunc(pathValidatorRelayers, m.handleValidatorRelayers).Methods(http.MethodGet)
	r.HandleFunc(pathState, m.handleState).Methods(http.MethodGet)

	// Memory endpoints
	r.HandleFunc(pathMemoryValidators, m.handleMemoryValidators).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryValidatorByIndex, m.handleMemoryValidatorInfo).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryValidatorsByWithdrawal, m.handleMemoryValidatorsByWithdrawal).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryFeesInfo, m.handleMemoryFeesInfo).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryPoolStatistics, m.handleMemoryStatistics).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryAllBlocks, m.handleMemoryAllBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryProposedBlocks, m.handleMemoryProposedBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryMissedBlocks, m.handleMemoryMissedBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryWrongFeeBlocks, m.handleMemoryWrongFeeBlocks).Methods(http.MethodGet)
	r.HandleFunc(pathMemoryDonations, m.handleMemoryDonations).Methods(http.MethodGet)

	// Onchain endpoints
	r.HandleFunc(pathOnchainMerkleProof, m.handleOnchainMerkleProof).Methods(http.MethodGet)

	// Not strictly necessary but good to have
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(prometheusMiddleware)

	return r
}

func (m *ApiService) StartHTTPServer() {
	log.Info("Starting HTTP server on ", m.ApiListenAddr)
	if m.srv != nil {
		log.Fatal("HTTP server already started")
	}

	m.srv = &http.Server{
		Addr: m.ApiListenAddr,
		//wrap handler with corsMiddleware, it passes execution to router handler when finished
		Handler: corsMiddleware(m.getRouter()),

		//ReadTimeout:       time.Duration(config.ServerReadTimeoutMs) * time.Millisecond,
		//ReadHeaderTimeout: time.Duration(config.ServerReadHeaderTimeoutMs) * time.Millisecond,
		//WriteTimeout:      time.Duration(config.ServerWriteTimeoutMs) * time.Millisecond,
		//IdleTimeout:       time.Duration(config.ServerIdleTimeoutMs) * time.Millisecond,

		//MaxHeaderBytes: config.ServerMaxHeaderBytes,
	}

	err := m.srv.ListenAndServe()
	if err != nil {
		log.Fatal("could not start http server: ", err)
	}
}

// Checks Origin header of the request and only allows from the desired origin or "" origin.
// Also adds CORS headers to the HTTP response so that the server indicates which origins and methods are allowed.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the CORS headers for all requests so that the browser allows the request
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// If the request method is OPTIONS, return a response with the allowed methods, headers, and origin
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *ApiService) handleRoot(w http.ResponseWriter, req *http.Request) {
	m.respondOK(w, "see api doc for available endpoints")
}

func (m *ApiService) handleMemoryStatistics(w http.ResponseWriter, req *http.Request) {
	// Ensure the oracle is ready
	if !m.OracleReady(uint64(64)) {
		m.respondError(w, http.StatusServiceUnavailable, "Oracle node is currently syncing and not serving requests")
		return
	}

	totalSubscribed := uint64(0)
	totalActive := uint64(0)
	totalYellowCard := uint64(0)
	totalRedCard := uint64(0)
	totalBanned := uint64(0)
	totalNotSubscribed := uint64(0)

	totalAccumulatedRewards := big.NewInt(0)
	totalPendingRewards := big.NewInt(0)

	for _, validator := range m.oracle.State().Validators {
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
	totalRewardsSent30DaysWei := big.NewInt(0)
	totalDonationsWei := big.NewInt(0)

	// Prevent underflow
	if uint64(m.oracle.State().LatestProcessedSlot) < SlotsInOneMonth {
		m.respondError(w, http.StatusInternalServerError, "head slot is lower than slots in a month, this should not happen")
	}

	if uint64(m.oracle.State().LatestProcessedBlock) < SlotsInOneMonth {
		m.respondError(w, http.StatusInternalServerError, "head block is lower than slots in a month, this should not happen")
	}

	// Only consider blocks in the last 30 days
	limitSlot := uint64(m.oracle.State().LatestProcessedSlot) - SlotsInOneMonth
	limitBlock := uint64(m.oracle.State().LatestProcessedBlock) - SlotsInOneMonth

	// Note that in a month we have SlotsInOneMonth slots, but not exactly that amount of blocks. If blocks
	// are missed we can have less. If blocks are missed we will take into account a time window
	// slightly higher than 30 days.

	for _, block := range m.oracle.State().ProposedBlocks {
		totalRewardsSentWei.Add(totalRewardsSentWei, block.Reward)

		// Filter blocks in the last 30 days
		if block.Slot > limitSlot {
			totalRewardsSent30DaysWei.Add(totalRewardsSent30DaysWei, block.Reward)
		}
	}

	for _, donation := range m.oracle.State().Donations {
		totalDonationsWei.Add(totalDonationsWei, donation.DonationAmount)

		// Note that rewards also take donations into account
		totalRewardsSentWei.Add(totalRewardsSentWei, donation.DonationAmount)

		if donation.Raw.BlockNumber > limitBlock {
			totalRewardsSent30DaysWei.Add(totalRewardsSent30DaysWei, donation.DonationAmount)
		}
	}

	rewardsPerValidatorPer30Days := big.NewInt(0)
	totalValidatorsEarning := totalActive + totalYellowCard
	if totalSubscribed != 0 {
		// This metric can be biased if multiple validators exit at once within the month
		rewardsPerValidatorPer30Days.Div(totalRewardsSent30DaysWei, big.NewInt(0).SetUint64(totalValidatorsEarning))
	}

	totalProposedBlocks := uint64(len(m.oracle.State().ProposedBlocks))
	avgBlockRewardWei := big.NewInt(0)

	// Avoid division by zero
	if totalProposedBlocks != 0 {
		avgBlockRewardWei = big.NewInt(0).Div(totalRewardsSentWei, big.NewInt(0).SetUint64(uint64(len(m.oracle.State().ProposedBlocks))))
	}

	m.respondOK(w, httpOkMemoryStatistics{
		TotalSubscribed:              totalSubscribed,
		TotalActive:                  totalActive,
		TotalYellowCard:              totalYellowCard,
		TotalRedCard:                 totalRedCard,
		TotalBanned:                  totalBanned,
		TotalNotSubscribed:           totalNotSubscribed,
		LatestCheckpointSlot:         m.oracle.State().LatestProcessedSlot,
		NextCheckpointSlot:           m.oracle.State().LatestProcessedSlot + m.cfg.CheckPointSizeInSlots,
		TotalAccumulatedRewardsWei:   totalAccumulatedRewards.String(),
		TotalPendingRewaradsWei:      totalPendingRewards.String(),
		TotalRewardsSentWei:          totalRewardsSentWei.String(),
		TotalDonationsWei:            totalDonationsWei.String(),
		AvgBlockRewardWei:            avgBlockRewardWei.String(),
		TotalRewardsSent30DaysWei:    totalRewardsSent30DaysWei.String(),
		RewardsPerValidatorPer30Days: rewardsPerValidatorPer30Days.String(),
		TotalProposedBlocks:          totalProposedBlocks,
		TotalMissedBlocks:            uint64(len(m.oracle.State().MissedBlocks)),
		TotalWrongFeeBlocks:          uint64(len(m.oracle.State().WrongFeeBlocks)),
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

	finalizedEpoch := uint64(finality.Finalized.Epoch)
	finalizedSlot := finalizedEpoch * constants.SlotsInEpoch

	oracleSync := false
	if m.oracle.State().LatestProcessedSlot-finalizedSlot == 0 {
		oracleSync = true
	}

	_, onchainSlot, err := m.Onchain.GetOnchainSlotAndRoot(apiRetryOpts...)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get onchain slot and root: "+err.Error())
	}

	// If the oracle is not in sync, we cant really calculate the slots till the next checkpoint
	// because we are behind. So we just set it to 0
	nextCheckpointInSlots := uint64(0)
	if finalizedSlot < (onchainSlot + m.cfg.CheckPointSizeInSlots) {
		nextCheckpointInSlots = onchainSlot + m.cfg.CheckPointSizeInSlots - finalizedSlot
	}

	status := httpOkStatus{
		IsConsensusInSync:           consInSync,
		IsExecutionInSync:           execInSync,
		IsOracleInSync:              oracleSync,
		LatestProcessedSlot:         m.oracle.State().LatestProcessedSlot,
		LatestProcessedBlock:        m.oracle.State().LatestProcessedBlock,
		LatestFinalizedEpoch:        finalizedEpoch,
		LatestFinalizedSlot:         finalizedSlot,
		OracleHeadDistance:          finalizedSlot - m.oracle.State().LatestProcessedSlot,
		NextCheckpointSlot:          onchainSlot + m.cfg.CheckPointSizeInSlots,
		NextCheckpointTime:          "", // TODO:
		NextCheckpointRemaining:     utils.SlotsToTime(nextCheckpointInSlots, constants.SecondsInSlot),
		NextCheckpointRemainingUnix: nextCheckpointInSlots * constants.SecondsInSlot,
		PreviousCheckpointSlot:      onchainSlot,
		PreviousCheckpointTime:      "", // TODO:
		PreviousCheckpointAge:       utils.SlotsToTime(finalizedSlot-onchainSlot, constants.SecondsInSlot),
		PreviousCheckpointAgeUnix:   (finalizedSlot - onchainSlot) * constants.SecondsInSlot,
		ExecutionChainId:            chainId.String(),
		ConsensusChainId:            strconv.FormatUint(depositContract.ChainID, 10),
		DepositContact:              hexutil.Encode(depositContract.Address[:]),
	}

	m.respondOK(w, status)
}

func (m *ApiService) handleConfig(w http.ResponseWriter, req *http.Request) {
	if m.cfg == nil {
		m.respondError(w, http.StatusInternalServerError, "no config loaded, nil value")
		return
	}
	m.respondOK(w, httpOkConfig{
		Network:                  m.cfg.Network,
		PoolAddress:              m.cfg.PoolAddress,
		DeployedSlot:             m.cfg.DeployedSlot,
		CheckPointSizeInSlots:    m.cfg.CheckPointSizeInSlots,
		PoolFeesPercentOver10000: m.cfg.PoolFeesPercentOver10000,
		PoolFeesAddress:          m.cfg.PoolFeesAddress,
		DryRun:                   m.cfg.DryRun,
		CollateralInWei:          m.cfg.CollateralInWei.String(),
	})
}

func (m *ApiService) handleMemoryValidators(w http.ResponseWriter, req *http.Request) {
	if !m.OracleReady(uint64(64)) {
		m.respondError(w, http.StatusServiceUnavailable, "Oracle node is currently syncing and not serving requests")
		return
	}
	validators := maps.Values(m.oracle.State().Validators)

	// Order by index
	sort.Slice(validators, func(i, j int) bool { return validators[i].ValidatorIndex < validators[j].ValidatorIndex })

	validatorsResp := make([]httpOkValidatorInfo, 0)
	for _, v := range validators {
		validatorsResp = append(validatorsResp, httpOkValidatorInfo{
			ValidatorStatus:       v.ValidatorStatus.String(),
			AccumulatedRewardsWei: v.AccumulatedRewardsWei.String(),
			PendingRewardsWei:     v.PendingRewardsWei.String(),
			CollateralWei:         v.CollateralWei.String(),
			WithdrawalAddress:     v.WithdrawalAddress,
			ValidatorIndex:        v.ValidatorIndex,
			ValidatorKey:          v.ValidatorKey,
		})
	}

	m.respondOK(w, validatorsResp)
}

func (m *ApiService) handleMemoryValidatorInfo(w http.ResponseWriter, req *http.Request) {
	if !m.OracleReady(MaxSlotsBehind) {
		m.respondError(w, http.StatusServiceUnavailable, "Oracle node is currently syncing and not serving requests")
		return
	}

	vars := mux.Vars(req)
	valIndexStr := vars["valindex"]
	valIndex, ok := IsValidIndex(valIndexStr)

	if !ok {
		m.respondError(w, http.StatusBadRequest, "invalid validator index: "+valIndexStr)
		return
	}

	validator, found := m.oracle.State().Validators[valIndex]
	if !found {
		m.respondError(w, http.StatusBadRequest, fmt.Sprint("could not find validator with index: ", valIndex))
		return
	}

	// TODO: Temporal, remove in production.
	if validator.ValidatorIndex != valIndex {
		validator.ValidatorIndex = valIndex
	}

	m.respondOK(w, validator)
}

func (m *ApiService) handleMemoryValidatorsByWithdrawal(w http.ResponseWriter, req *http.Request) {
	if !m.OracleReady(MaxSlotsBehind) {
		m.respondError(w, http.StatusServiceUnavailable, "Oracle node is currently syncing and not serving requests")
		return
	}

	vars := mux.Vars(req)
	withdrawalAddress := vars["withdrawalAddress"]

	// Use always lowercase
	withdrawalAddress = strings.ToLower(withdrawalAddress)

	if !IsValidAddress(withdrawalAddress) {
		m.respondError(w, http.StatusBadRequest, "invalid withdrawalAddress: "+withdrawalAddress)
		return
	}

	if m.Onchain.Validators() == nil {
		m.respondError(w, http.StatusInternalServerError, "finalized validators not loaded yet, try again later")
		return
	}

	// We return
	// 1) validators using this withdrawal address but not tracked by the oracle
	// 2) validators using this withdrawal address and tracked by the oracle (eg already subscribed)
	requestedValidators := make(map[uint64]*oracle.ValidatorInfo, 0)

	// 1) Get all onchain validators for that withdrawal address (untracked)
	for valIndex, validator := range m.Onchain.Validators() {

		// Check if the withdrawal address matches the requested one
		credStr := hex.EncodeToString(validator.Validator.WithdrawalCredentials)
		eth1Add, err := utils.GetEth1Address(credStr) // TODO: Use the new function

		// Skip validators without non eth withdrawal address (bls address)
		if err != nil {
			continue
		}

		// Skip if the address does not match with the requested
		if !AreAddressEqual(eth1Add, withdrawalAddress) {
			continue
		}

		// Skip validators that cannot be subscribed
		if !oracle.CanValidatorSubscribeToPool(validator) {
			continue
		}

		requestedValidators[uint64(valIndex)] = &oracle.ValidatorInfo{
			ValidatorStatus:       oracle.Untracked,
			AccumulatedRewardsWei: big.NewInt(0),
			PendingRewardsWei:     big.NewInt(0),
			CollateralWei:         big.NewInt(0),
			WithdrawalAddress:     eth1Add,
			ValidatorIndex:        uint64(validator.Index),
			ValidatorKey:          hexutil.Encode(validator.Validator.PublicKey[:]),
		}
	}

	// 2) Get all tracked validators for that withdrawal address (tracked)
	validatorsCopy := make(map[uint64]*oracle.ValidatorInfo)

	// Imporant! This is a deep copy, otherwise we will modify the state
	utils.DeepCopy(m.oracle.State().Validators, &validatorsCopy)
	for valIndex, validator := range validatorsCopy {
		// Just overwrite the untracked validators with oracle state
		if AreAddressEqual(validator.WithdrawalAddress, withdrawalAddress) {
			requestedValidators[valIndex] = validator

			// TODO: Temporal, remove in production.
			if validator.ValidatorIndex != valIndex {
				validator.ValidatorIndex = valIndex
			}
		}
	}

	// If at this point we have no validators, just return empty to avoid more processing
	// TODO: Cant i return earlier? after 2)?
	if len(requestedValidators) == 0 {
		m.respondOK(w, make([]httpOkValidatorInfo, 0))
		return
	}

	// Now we apply the state transition to these validators, based on what we have seen
	// onchain since the latest finalized slot util head. This is neccesary because the
	// oracle runs all calculations on finalized states, but the api must report to the
	// users without this 15 minutes-ish delay.
	// This applies a non-finalized state to the validators, creating a virtual state
	// only used for the api.

	if m.oracle.State().LatestProcessedBlock == 0 {
		m.respondError(w, http.StatusInternalServerError, "latest processed block is 0, try again later")
		return
	}

	firstNotProcessedBlock := m.oracle.State().LatestProcessedBlock + 1

	// TODO: Cache this, very inneficient to get it every time
	allSubsTillHead, err := m.GetSubscriptionsTillHead(firstNotProcessedBlock)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get subscriptions: "+err.Error())
		return
	}
	allUnsubsTillHead, err := m.GetUnsubscriptionsTillHead(firstNotProcessedBlock)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get unsubscriptions: "+err.Error())
		return
	}

	// Apply latest seen events to the existing state. This is a "virtual" state, just for the api
	// so that users are aware of the latest events, without waiting for the next finalized state.
	m.ApplyNonFinalizedState(
		allSubsTillHead,
		allUnsubsTillHead,
		requestedValidators)

	// Sort by index
	values := maps.Values(requestedValidators)
	sort.Slice(values, func(i, j int) bool { return values[i].ValidatorIndex < values[j].ValidatorIndex })

	validatorsResp := make([]httpOkValidatorInfo, 0)
	for _, v := range values {
		validatorsResp = append(validatorsResp, httpOkValidatorInfo{
			ValidatorStatus:       v.ValidatorStatus.String(),
			AccumulatedRewardsWei: v.AccumulatedRewardsWei.String(),
			PendingRewardsWei:     v.PendingRewardsWei.String(),
			CollateralWei:         v.CollateralWei.String(),
			WithdrawalAddress:     v.WithdrawalAddress,
			ValidatorIndex:        v.ValidatorIndex,
			ValidatorKey:          v.ValidatorKey,
			SubscriptionType:      v.SubscriptionType.String(),
		})
	}
	m.respondOK(w, validatorsResp)
}

func (m *ApiService) handleMemoryFeesInfo(w http.ResponseWriter, req *http.Request) {
	m.respondOK(w, httpOkMemoryFeesInfo{
		PoolFeesPercentOver10000: m.oracle.State().PoolFeesPercentOver10000,
		PoolFeesAddress:          m.oracle.State().PoolFeesAddress,
		PoolAccumulatedFees:      m.oracle.State().PoolAccumulatedFees.String(),
	})
}

func (m *ApiService) handleMemoryAllBlocks(w http.ResponseWriter, req *http.Request) {
	// Concat all the blocks, order is not guaranteed
	allBlocks := make([]httpOkBlock, 0)

	for _, block := range m.oracle.State().ProposedBlocks {
		allBlocks = append(allBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}

	for _, block := range m.oracle.State().MissedBlocks {
		allBlocks = append(allBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}

	for _, block := range m.oracle.State().WrongFeeBlocks {
		allBlocks = append(allBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}

	m.respondOK(w, allBlocks)
}

func (m *ApiService) handleMemoryProposedBlocks(w http.ResponseWriter, req *http.Request) {
	proposedBlocks := make([]httpOkBlock, 0)
	for _, block := range m.oracle.State().ProposedBlocks {
		proposedBlocks = append(proposedBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}
	m.respondOK(w, proposedBlocks)
}

func (m *ApiService) handleMemoryMissedBlocks(w http.ResponseWriter, req *http.Request) {
	missedBlocks := make([]httpOkBlock, 0)
	for _, block := range m.oracle.State().MissedBlocks {
		missedBlocks = append(missedBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}
	m.respondOK(w, missedBlocks)
}

func (m *ApiService) handleMemoryWrongFeeBlocks(w http.ResponseWriter, req *http.Request) {
	wrongFeeBlocks := make([]httpOkBlock, 0)
	for _, block := range m.oracle.State().WrongFeeBlocks {
		wrongFeeBlocks = append(wrongFeeBlocks, httpOkBlock{
			Slot:              block.Slot,
			Block:             block.Block,
			ValidatorIndex:    block.ValidatorIndex,
			ValidatorKey:      block.ValidatorKey,
			BlockType:         block.BlockType.String(),
			Reward:            block.Reward.String(),
			RewardType:        block.RewardType.String(),
			WithdrawalAddress: block.WithdrawalAddress,
		})
	}
	m.respondOK(w, wrongFeeBlocks)
}

func (m *ApiService) handleMemoryDonations(w http.ResponseWriter, req *http.Request) {
	donations := make([]httpOkDonation, 0)
	for _, donation := range m.oracle.State().Donations {
		donations = append(donations, httpOkDonation{
			AmountWei: donation.DonationAmount.String(),
			Block:     donation.Raw.BlockNumber,
			TxHash:    donation.Raw.TxHash.String(),
			Sender:    donation.Sender.String(),
		})
	}
	m.respondOK(w, donations)
}

func (m *ApiService) handleOnchainMerkleProof(w http.ResponseWriter, req *http.Request) {
	if !m.OracleReady(MaxSlotsBehind) {
		m.respondError(w, http.StatusServiceUnavailable, "Oracle node is currently syncing and not serving requests")
		return
	}

	vars := mux.Vars(req)
	withdrawalAddress := vars["withdrawalAddress"]

	if !IsValidAddress(withdrawalAddress) {
		m.respondError(w, http.StatusBadRequest, "invalid WithdrawalAddress: "+withdrawalAddress)
		return
	}

	// Use always lowercase
	withdrawalAddress = strings.ToLower(withdrawalAddress)

	contractRoot, contractSlot, err := m.Onchain.GetOnchainSlotAndRoot(apiRetryOpts...)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get onchain slot and root: "+err.Error())
		return
	}

	_, found := m.oracle.State().CommitedStates[contractSlot]
	if !found {
		m.respondError(w, http.StatusInternalServerError, "could not find onchain slot in oracle state: "+strconv.FormatUint(contractSlot, 10))
		return
	}

	// Check if the oracle root matches the one offchain
	if contractRoot != m.oracle.State().CommitedStates[contractSlot].MerkleRoot {
		m.respondError(w, http.StatusInternalServerError,
			"contract merkle root does not match oracle state: "+
				contractRoot+" vs "+m.oracle.State().CommitedStates[contractSlot].MerkleRoot)
		return
	}

	// Get the proofs of this withdrawal address (to be used onchain to claim rewards)
	proofs, proofFound := m.oracle.State().CommitedStates[contractSlot].Proofs[withdrawalAddress]
	if !proofFound {
		m.respondError(w, http.StatusBadRequest, "could not find proof for WithdrawalAddress: "+withdrawalAddress)
		return
	}

	// Get the leafs of this withdrawal address (to be used onchain to claim rewards)
	leafs, leafsFound := m.oracle.State().CommitedStates[contractSlot].Leafs[withdrawalAddress]
	if !leafsFound {
		m.respondError(w, http.StatusBadRequest, "could not find leafs for WithdrawalAddress: "+withdrawalAddress)
		return
	}

	// Get validators that are registered to this withdrawal address in the pool
	registeredValidators := make([]uint64, 0)
	for valIndex, validator := range m.oracle.State().CommitedStates[contractSlot].Validators {
		if strings.ToLower(validator.WithdrawalAddress) == strings.ToLower(withdrawalAddress) {
			registeredValidators = append(registeredValidators, valIndex)
		}
	}

	claimed, err := m.Onchain.GetContractClaimedBalance(withdrawalAddress, nil, apiRetryOpts...)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get claimed balance so far from contract: "+err.Error())
		return
	}

	totalPending := big.NewInt(0)

	for _, validator := range m.oracle.State().CommitedStates[contractSlot].Validators {
		if strings.ToLower(validator.WithdrawalAddress) == strings.ToLower(withdrawalAddress) {
			totalPending.Add(totalPending, validator.PendingRewardsWei)
		}
	}

	m.respondOK(w, httpOkProofs{
		LeafWithdrawalAddress:      leafs.WithdrawalAddress,
		LeafAccumulatedBalance:     leafs.AccumulatedBalanceWei.String(),
		MerkleRoot:                 m.oracle.State().CommitedStates[contractSlot].MerkleRoot,
		CheckpointSlot:             m.oracle.State().CommitedStates[contractSlot].Slot,
		Proofs:                     proofs,
		RegisteredValidators:       registeredValidators,
		TotalAccumulatedRewardsWei: leafs.AccumulatedBalanceWei.String(),
		ClaimableRewardsWei:        new(big.Int).Sub(leafs.AccumulatedBalanceWei, claimed).String(),
		AlreadyClaimedRewardsWei:   claimed.String(),
		PendingRewardsWei:          totalPending.String(),
	})
}

func (m *ApiService) handleValidatorRelayers(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	valPubKeys := vars["valpubkey"]
	if valPubKeys == "" {
		m.respondError(w, http.StatusBadRequest, "No validator pubkey provided!")
		return
	}

	keys := strings.Split(valPubKeys, ",")
	if len(keys) > 50 {
		m.respondError(w, http.StatusBadRequest, "Maximum number of pubkeys exceeded (max: 50)")
		return
	}

	for _, key := range keys {
		if !IsValidPubkey(key) {
			m.respondError(w, http.StatusBadRequest, "Invalid validator pubkey format: "+key)
			return
		}
	}

	results, allValid, err := m.processValidatorsConcurrently(keys)
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.respondOK(w, httpOkMultiRelayersState{
		Validators:           results,
		AllValidatorsCorrect: allValid,
		IncorrectValidators:  m.extractIncorrectValidators(results),
	})
}

func (m *ApiService) processValidatorsConcurrently(keys []string) ([]httpOkRelayersState, bool, error) {
	var results []httpOkRelayersState
	allValidatorsRegisteredCorrectFee := true
	resultsChan := make(chan validatorRelayResult)

	for _, key := range keys {
		go m.processSingleValidator(key, resultsChan)
	}

	for range keys {
		res := <-resultsChan
		if res.Err != nil {
			return nil, false, res.Err
		}
		if !res.IsValidatorValid {
			allValidatorsRegisteredCorrectFee = false
		}
		results = append(results, res.ValidatorResult)
	}
	close(resultsChan)

	return results, allValidatorsRegisteredCorrectFee, nil
}

func (m *ApiService) processSingleValidator(valPubKey string, resultsChan chan validatorRelayResult) {
	var correctFeeRelays []httpRelay
	var wrongFeeRelays []httpRelay
	var unregisteredRelays []httpRelay
	registeredCorrectFee := false

	relays := m.cliCfg.RelayersEndpoints

	for _, relay := range relays {
		url := fmt.Sprintf("%s/relay/v1/data/validator_registration?pubkey=%s", relay, valPubKey)
		resp, err := http.Get(url)
		if err != nil {
			resultsChan <- validatorRelayResult{
				Err: fmt.Errorf("error calling relayer %s for validator %s: %v", relay, valPubKey, err),
			}
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			resultsChan <- validatorRelayResult{
				Err: fmt.Errorf("error reading response from relayer %s for validator %s: %v", relay, valPubKey, err),
			}
			return
		}

		// If the validator is or has been registered, the relayer will return a 200 message with the signed registration message.
		// https://flashbots.github.io/relay-specs/#/Data/getValidatorRegistration
		if resp.StatusCode == http.StatusOK {
			signedRegistration := &builderApiV1.SignedValidatorRegistration{}

			if err = json.Unmarshal(bodyBytes, signedRegistration); err != nil {
				resultsChan <- validatorRelayResult{
					Err: fmt.Errorf("error unmarshalling relay response from relayer %s for validator %s: %v", relay, valPubKey, err),
				}
				return
			}

			relayRegistration := httpRelay{
				RelayAddress: relay,
				FeeRecipient: signedRegistration.Message.FeeRecipient.String(),
				Timestamp:    fmt.Sprintf("%d", signedRegistration.Message.Timestamp.UnixNano()),
			}

			// If the fee recipient matches the pool address, the relayer is registered with the correct fee recipient (the smoothing pool one)
			if utils.Equals(signedRegistration.Message.FeeRecipient.String(), m.Onchain.PoolAddress) {
				correctFeeRelays = append(correctFeeRelays, relayRegistration)
			} else {
				// if the fee recipient does not match the pool address, the relayer is registered but with the wrong fee recipient
				wrongFeeRelays = append(wrongFeeRelays, relayRegistration)
			}

		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
			// If the validator has never been registered, the relayer will return code 400 or 404 (depending on the relay)
			unregisteredRelays = append(unregisteredRelays, httpRelay{RelayAddress: relay})
		} else {
			// If we get here, the relayer had an internal server error, so we couldnt check if the validator is/was registered with the correct
			// fee recipient. We return an error.
			resultsChan <- validatorRelayResult{
				Err: fmt.Errorf("error calling relayer %s for validator %s: %v", relay, valPubKey, string(bodyBytes)),
			}
		}
	}

	// If there are no wrong fee relays and there are correct fee relays, the validator is registered with the correct fee recipient
	// we do not accept validators that have not registered to any relay
	if len(wrongFeeRelays) == 0 && len(correctFeeRelays) > 0 {
		registeredCorrectFee = true
	}

	resultsChan <- validatorRelayResult{
		ValidatorResult: httpOkRelayersState{
			ValPubKey:            valPubKey,
			CorrectFeeRecipients: registeredCorrectFee,
			CorrectFeeRelays:     correctFeeRelays,
			WrongFeeRelays:       wrongFeeRelays,
			UnregisteredRelays:   unregisteredRelays,
		},
		IsValidatorValid: registeredCorrectFee,
	}
}

func (m *ApiService) extractIncorrectValidators(results []httpOkRelayersState) []string {
	var incorrectValidators []string
	for _, result := range results {
		if !result.CorrectFeeRecipients {
			incorrectValidators = append(incorrectValidators, result.ValPubKey)
		}
	}
	return incorrectValidators
}

func (m *ApiService) handleState(w http.ResponseWriter, req *http.Request) {
	// Just dump the whole known state of the oracle. This is useful for debugging. Note that
	// if the state becomes too big, we may need to page it here. This use the same type
	// as the oracle state type.
	state, err := m.oracle.StateWithHash()
	if err != nil {
		m.respondError(w, http.StatusInternalServerError, "could not get state: "+err.Error())
		return
	}
	m.respondOK(w, state)
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

// The validator's BLS public key, uniquely identifying them. 48-bytes, hex encoded with 0x prefix, case insensitive.
// example: example: 0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7445a6d1a2753e5f3e8b1cfe39b46f43611ef74a
func IsValidPubkey(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-f]{96}$")
	return re.MatchString(v)
}

// Copied from oracle/utils. Cant import due to circular dependency
// TODO: Move to utils package
func AreAddressEqual(address1 string, address2 string) bool {
	if len(address1) != len(address2) {
		log.Fatal("address length mismatch: ",
			"add1: ", address1,
			"add2: ", address2)
	}
	if strings.ToLower(address1) == strings.ToLower(address2) {
		return true
	}
	return false
}

// TODO: unsure if move this somewhere else
func (m *ApiService) GetSubscriptionsTillHead(latestProcessedBlock uint64) ([]Subscription, error) {
	// TODO: add check here to ensure its a reasonable amount of blocks. should be around 15-20 minutes in blocks
	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: latestProcessedBlock, End: nil}

	// Note that this event can be both donations and mev rewards
	itrSubs, err := m.Onchain.Contract.FilterSubscribeValidator(filterOpts)
	if err != nil {
		return nil, errors.Wrap(err, "could not subscribe to validator events")
	}

	// Loop over all found events. Super inneficient. just Proof of concept
	blockSubscriptions := make([]Subscription, 0)
	for itrSubs.Next() {
		sub := Subscription{
			Event:     itrSubs.Event,
			Validator: m.Onchain.Validators()[phase0.ValidatorIndex(itrSubs.Event.ValidatorID)],
		}
		blockSubscriptions = append(blockSubscriptions, sub)
	}
	err = itrSubs.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close subscription iterator")
	}
	return blockSubscriptions, nil
}

func (m *ApiService) GetUnsubscriptionsTillHead(latestProcessedBlock uint64) ([]Unsubscription, error) {
	// TODO: add check here to ensure its a reasonable amount of blocks. should be around 15-20 minutes in blocks
	filterOpts := &bind.FilterOpts{Context: context.Background(), Start: latestProcessedBlock, End: nil}
	// Note that this event can be both donations and mev rewards
	itrUnsubs, err := m.Onchain.Contract.FilterUnsubscribeValidator(filterOpts)
	if err != nil {
		return nil, errors.Wrap(err, "could not subscribe to validator events")
	}

	// Loop over all found events, TODO: inneficient. only finter events of this validator.
	blockUnsubscriptions := make([]Unsubscription, 0)
	for itrUnsubs.Next() {
		unsub := Unsubscription{
			Event:     itrUnsubs.Event,
			Validator: m.Onchain.Validators()[phase0.ValidatorIndex(itrUnsubs.Event.ValidatorID)],
		}
		blockUnsubscriptions = append(blockUnsubscriptions, unsub)
	}
	err = itrUnsubs.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not close subscription iterator")
	}
	return blockUnsubscriptions, nil
}

func (m *ApiService) ApplyNonFinalizedState(
	subs []Subscription,
	unsubs []Unsubscription,
	validators map[uint64]*oracle.ValidatorInfo) {

	eventsBlocksList := make([]uint64, 0)

	for _, sub := range subs {
		block := sub.Event.Raw.BlockNumber
		found := false
		for _, b := range eventsBlocksList {
			if b == block {
				found = true
			}
		}
		if !found {
			eventsBlocksList = append(eventsBlocksList, block)
		}
	}
	for _, unsub := range unsubs {
		block := unsub.Event.Raw.BlockNumber
		found := false
		for _, b := range eventsBlocksList {
			if b == block {
				found = true
			}
		}
		if !found {
			eventsBlocksList = append(eventsBlocksList, block)
		}
	}

	sort.Slice(eventsBlocksList, func(i, j int) bool { return eventsBlocksList[i] < eventsBlocksList[j] })

	for _, block := range eventsBlocksList {
		blockSub := GetSubInBlock(subs, block)
		blockUnsub := GetUnsubInBlock(unsubs, block)

		for _, subInBlock := range blockSub {
			valIndex := subInBlock.Event.ValidatorID
			val, found := validators[valIndex]
			if found {
				valWithdrawalAddress := val.WithdrawalAddress
				eventAddress := subInBlock.Event.Sender.String()
				if AreAddressEqual(valWithdrawalAddress, eventAddress) {
					if subInBlock.Event.SubscriptionCollateral.Cmp(m.cfg.CollateralInWei) >= 0 {
						if oracle.CanValidatorSubscribeToPool(subInBlock.Validator) {
							if val.ValidatorStatus == oracle.Untracked || val.ValidatorStatus == oracle.NotSubscribed {
								validators[valIndex].ValidatorStatus = oracle.Active
								validators[valIndex].PendingRewardsWei.Add(validators[valIndex].PendingRewardsWei, subInBlock.Event.SubscriptionCollateral)
								// Accumulated is not updated, since that has to be done onchain
							}
						}
					}
				}
			}
		}

		for _, unsubInBlock := range blockUnsub {
			valIndex := unsubInBlock.Event.ValidatorID
			val, found := validators[valIndex]
			if found {
				valWithdrawalAddress := val.WithdrawalAddress
				eventAddress := unsubInBlock.Event.Sender.String()
				if AreAddressEqual(valWithdrawalAddress, eventAddress) {
					if val.ValidatorStatus == oracle.Active ||
						val.ValidatorStatus == oracle.YellowCard ||
						val.ValidatorStatus == oracle.RedCard {
						validators[valIndex].ValidatorStatus = oracle.NotSubscribed
						validators[valIndex].PendingRewardsWei = big.NewInt(0)
						// Accumulated is not updated, since that has to be done onchain
					}
				}
			}
		}
	}
}

func (m *ApiService) OracleReady(maxSlotsBehind uint64) bool {
	// Allow 3 epochs 32*3 slots out of sync (behind latest finalized). This allows to always serve requests since
	// otherwise the oracle wont be able to reply, since from time to time its normal that it fall behind sync
	// since it has to process the new epochs that keep arriving.

	finality, err := m.Onchain.ConsensusClient.Finality(context.Background(), "finalized")
	if err != nil {
		return false
	}

	finalizedSlot := uint64(finality.Finalized.Epoch) * constants.SlotsInEpoch
	slotsFromFinalized := finalizedSlot - m.oracle.State().LatestProcessedSlot

	// Use this if we want full in sync to latest finalized
	/*oracleInSync := false
	if slotsFromFinalized == 0 {
		oracleInSync = true
	}
	_ = oracleInSync*/

	if slotsFromFinalized > maxSlotsBehind {
		return false
	}
	return true
}

func GetSubInBlock(subs []Subscription, block uint64) []Subscription {
	filteredSubs := make([]Subscription, 0)
	for _, sub := range subs {
		if sub.Event.Raw.BlockNumber == block {
			filteredSubs = append(filteredSubs, sub)
		}
	}
	return filteredSubs
}

func GetUnsubInBlock(subs []Unsubscription, block uint64) []Unsubscription {
	filteredUnsubs := make([]Unsubscription, 0)
	for _, unsub := range subs {
		if unsub.Event.Raw.BlockNumber == block {
			filteredUnsubs = append(filteredUnsubs, unsub)
		}
	}
	return filteredUnsubs
}
