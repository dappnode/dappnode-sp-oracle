package main

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/dappnode/mev-sp-oracle/api"
	"github.com/dappnode/mev-sp-oracle/config"
	"github.com/dappnode/mev-sp-oracle/constants"
	"github.com/dappnode/mev-sp-oracle/metrics"
	"github.com/dappnode/mev-sp-oracle/oracle"
	"github.com/dappnode/mev-sp-oracle/utils"
	"github.com/ethereum/go-ethereum/common"

	log "github.com/sirupsen/logrus"
)

// How often onchain validators are reloaded: 1200 slots is 4 hours
var UpdateValidatorsIntervalSlots = uint64(1200)

// logs file and path
const LogsName = "logs.txt"
const LogsFolder = "oracle-logs"

// How often in hours we run onchain reconciliation
const ReconciliationEveryHours = int64(3)

func main() {
	// Load config from cli
	cliCfg, err := config.NewCliConfig()
	if err != nil {
		log.Fatal("error parsing the cli config: ", err)
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.DisableColors = true
	log.SetFormatter(customFormatter)
	//log.SetReportCaller(true)

	//file is created if not exists, otherwise it appends errors to the existing file
	//0666 means permisions to read and write to all users, but not execute
	err = os.MkdirAll(LogsFolder, os.ModePerm)
	if err != nil {
		log.Fatal("error creating the oracleLogs.txt folder: ", err)
	}
	file, err := os.OpenFile(filepath.Join(LogsFolder, LogsName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("error opening or creating the oracleLogs.txt file: ", err)
	}
	log.Info("Persisting logs in ", file.Name())
	defer file.Close()

	// Create a MultiWriter with file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	// Set log output to the MultiWriter
	log.SetOutput(multiWriter)

	log.Info("Starting smoothing pool oracle")
	log.Info("Version: ", config.ReleaseVersion)
	metrics.Version.WithLabelValues(config.ReleaseVersion).Set(1)

	// Set log-level
	logLevel, err := log.ParseLevel(cliCfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)

	// Load key with rights to update the oracle (if not dry run)
	var updaterKey *ecdsa.PrivateKey
	var updaterAddress common.Address
	if !cliCfg.DryRun {
		keystore, err := utils.DecryptKey(cliCfg)
		if err != nil {
			log.Fatal("Could not decrypt updater key: ", err)
		}
		updaterKey = keystore.PrivateKey
		updaterAddress = keystore.Address
		log.Info("Oracle contract will be updated with new roots using address: ", updaterAddress.String())
	}

	// Instance of the onchain object to handle onchain interactions
	onchain, err := oracle.NewOnchain(cliCfg, updaterKey)
	if err != nil {
		log.Fatal("Could not create new onchain object: ", err)
	}

	if !cliCfg.DryRun {
		log.Info("Checking if configured address ", updaterAddress.String(), " is whitelisted to update the contract")
		isWhitelisted, err := onchain.IsAddressWhitelisted(updaterAddress)
		if err != nil {
			log.Fatal("Could not get whitelist status: " + err.Error())
		}
		if !isWhitelisted {
			log.Fatal("Pool address is not whitelisted, please run the 'whitelist' command first")
		}
		log.Info("Ok ", updaterAddress.String(), " is whitelisted")

		// Check the updater address has some Eth balance
		balance, err := onchain.GetAddressEthBalance(updaterAddress)
		if err != nil {
			log.Fatal("Could not get updater address balance: ", err)
		}

		// Ensure it has balance, otherwise it wont be able to pay tx fees
		if balance.Cmp(big.NewInt(0)) == 0 {
			log.Fatal("Updater address: ", updaterAddress.String(), " has no balance, please send some Eth to it")
		} else {
			log.Info("Updater address: ", updaterAddress.String(), " has balance: ", utils.WeiToEther(balance), "Eth, ensure its enough to cover txs during some time")
		}
	}

	// Populate config, most of the parameters are loaded from the smart contract
	cfg := onchain.GetConfigFromContract(cliCfg)

	// Create the oracle instance
	oracleInstance := oracle.NewOracle(cfg)
	oracleInstance.SetGetSetOfValidatorsFunc(onchain.GetSetOfValidators)
	oracleInstance.SetOwnerToPodFunc(onchain.GetOwnerToPod)
	oracleInstance.SetValidatorRestakedEventsFunc(onchain.GetValidatorRestakedEvents)

	// If checkpoint sync url is provided, load state from it
	if cliCfg.CheckPointSyncUrl != "" {
		log.Warn("Checkpoint sync url provided, loading state from it. Ensure you trust the provider: ", cliCfg.CheckPointSyncUrl)
		resp, err := http.Get(cliCfg.CheckPointSyncUrl)
		if err != nil {
			log.Fatal("Could not call checkpoint sync endpoint: ", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal("Got error: ", resp.Status, " from checkpoint sync url: ", cliCfg.CheckPointSyncUrl)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Could not read response body: ", err)
		}

		_, err = oracleInstance.LoadFromBytes(bodyBytes)
		if err != nil {
			log.Fatal("Critical error loading state from checkpoint: ", err)
		}

	} else {
		found, err := oracleInstance.LoadFromJson()
		if err != nil {
			log.Fatal("Critical error loading state from json: ", err)
		}
		if !found {
			log.Warn("Previous state not found or could not be loaded, syncing from the begining slot=", oracleInstance.State().DeployedSlot)
		} else {
			log.Info("Found previous state to continue syncing")
		}
	}

	// Get onchain root and slot
	_, onchainSlot, err := onchain.GetOnchainSlotAndRoot()
	if err != nil {
		log.Fatal("Could not get onchain slot and root: ", err)
	}

	// Get latest commited slot locally
	latestCommited, _ := oracleInstance.LatestCommitedSlot()

	// Check that the oracle has not synced beyond the onchain slot. Only if not dry run
	if !cfg.DryRun {
		if latestCommited > onchainSlot {
			log.Warn("The loaded state goes beyond the onchain slot, trying to reload an old state. onchainSlot=",
				onchainSlot, " latestCommited=", latestCommited)
			found, err := oracleInstance.LoadGivenState(onchainSlot)
			if err != nil {
				log.Fatal("Critical error loading state from json: ", err)
			}
			if !found {
				log.Fatal("Could not find a save state for slot ", onchainSlot)
			}
		}
	}

	api := api.NewApiService(cfg, cliCfg, oracleInstance, onchain)

	metrics.RunMetrics(cliCfg.MetricsPort)
	go api.StartHTTPServer()
	go mainLoop(oracleInstance, onchain, cfg)

	// Wait for signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	for {
		sig := <-sigCh

		// Save state in SIGINT or SIGTERM
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			err := oracleInstance.SaveToJson(false)
			if err != nil {
				log.Error("Could not save state to json: ", err)
			} else {
				log.Info("State saved to json")
			}
		}

		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == os.Interrupt || sig == os.Kill {
			break
		}
	}

	log.Info("Oracle gracefully stopped")
}

func mainLoop(oracleInstance *oracle.Oracle, onchain *oracle.Onchain, cfg *oracle.Config) {

	lastReconciliationTime := int64(0)

	// Load all the validators from the beacon chain
	onchain.RefreshBeaconValidators()

	log.WithFields(log.Fields{
		"LatestProcessedSlot": oracleInstance.State().LatestProcessedSlot,
		"NextSlotToProcess":   oracleInstance.State().NextSlotToProcess,
	}).Info("Processing, see api for progress")

	for {
		// Ensure that the nodes we are using are in sync with the blockchain (consensus + execution)
		inSync, err := onchain.AreNodesInSync()
		if err != nil {
			log.Fatal("Could not get nodes in sync status:", err)
		}
		if !inSync {
			log.Warn("Nodes are not in sync, skipping until in sync")
			time.Sleep(15 * time.Second)
			continue
		}

		finalizedBlockHeader, err := onchain.FinalizedBeaconBlockHeader()
		if err != nil {
			log.Error("Could not get finalized status, sleeping and retrying:", err)
			time.Sleep(15 * time.Second)
			continue
		}

		finalizedSlot := uint64(finalizedBlockHeader.Header.Message.Slot)

		if finalizedSlot >= oracleInstance.State().NextSlotToProcess {

			// Fetch block information
			fullBlock := onchain.FetchFullBlock(oracleInstance.State().NextSlotToProcess, oracleInstance)

			// Process the block
			processedSlot, err := oracleInstance.AdvanceStateToNextSlot(fullBlock)
			if err != nil {
				log.Fatal(err)
			}
			slotToLatestFinalized := finalizedSlot - oracleInstance.State().LatestProcessedSlot

			// Update metrics
			metrics.DistanceFromFinalizedSlot.Set(float64(slotToLatestFinalized))
			metrics.LatestProcessedSlot.Set(float64(oracleInstance.State().LatestProcessedSlot))
			metrics.LatestProcessedBlock.Set(float64(oracleInstance.State().LatestProcessedBlock))

			log.Debug("[", processedSlot, "/", finalizedSlot, "] Processed until slot, remaining: ",
				slotToLatestFinalized, " (", utils.SlotsToTime(slotToLatestFinalized, constants.SecondsInSlot), " ago)")

		} else {
			// We are in sync, no new finalized slot, wait a bit
			log.WithFields(log.Fields{
				"ChainFinalizedSlot": finalizedSlot,
				"OracleStateSlot":    oracleInstance.State().LatestProcessedSlot,
			}).Debug("Waiting for new finalized slot")

			// From time to time we do onchain reconciliation to ensure our assets match our liablities
			if time.Now().Unix()-lastReconciliationTime > (ReconciliationEveryHours * 3600) {
				log.Info("Running onchain reconciliation. Last one was: ", lastReconciliationTime)

				// If we are up to date, this is the latest finalized block. Run this only in the last block, otherwise
				// a non archival node will error "missing trie node". Non archival nodes don't store much before last
				// finalized block.
				finalizedBlock := big.NewInt(0).SetUint64(oracleInstance.State().LatestProcessedBlock)
				uniqueAddresses := oracleInstance.GetUniqueWithdrawalAddresses()

				// If EL is not in archival mode, this wont work in longs periods of non finality.
				retryOption := retry.Attempts(1)
				poolEthBalanceWei, err1 := onchain.GetPoolEthBalance(finalizedBlock, retryOption)
				claimedPerAccount, err2 := onchain.GetClaimedPerWithdrawalAddress(uniqueAddresses, finalizedBlock, retryOption)
				if err1 != nil || err2 != nil {
					log.Warn("Could not get pool eth balance for reconciliation, normal when no finality: ", err1, err2)
				} else {
					// If we could fetch the data, run onchain reconciliation
					err = oracleInstance.RunOnchainReconciliation(poolEthBalanceWei, claimedPerAccount)
					if err != nil {
						log.Fatal("Reconciliation failed, state was not commited: ", err)
					}
				}
				lastReconciliationTime = time.Now().Unix()
			}

			time.Sleep(1 * time.Minute)
			continue
		}

		// Every X slots we update the onchain validators and cleanup any stranded oracle validators
		if oracleInstance.State().LatestProcessedSlot%UpdateValidatorsIntervalSlots == 0 {
			onchain.RefreshBeaconValidators()
		}

		// Every CheckPointSizeInSlots we commit the state given some conditions, starting from
		// when the contract was deployed.
		isCheckpoint, err := oracleInstance.IsCheckpoint()
		if err != nil {
			log.Fatal("Could not check if we are at a checkpoint: ", err)
		}
		if isCheckpoint {
			log.WithFields(log.Fields{
				"LatestProcessedSlot":   oracleInstance.State().LatestProcessedSlot,
				"CheckPointSizeInSlots": cfg.CheckPointSizeInSlots,
				"DeployedSlot":          oracleInstance.State().DeployedSlot,
			}).Info("Checkpoint reached")

			// TODO: Temporally commented. Doesn't work well with non archival node
			//log.Info("Running onchain reconciliation. Last one was: ", lastReconciliationTime)
			//lastReconciliationTime = time.Now().Unix()

			// If we are up to date, this is the latest finalized block. Run this only in the last block, otherwise
			// a non archival node will error "missing trie node". Non archival nodes don't store much before last
			// finalized block.
			//finalizedBlock := big.NewInt(0).SetUint64(oracleInstance.State().LatestProcessedBlock)
			//uniqueAddresses := oracleInstance.GetUniqueWithdrawalAddresses()
			//poolEthBalanceWei, err := onchain.GetPoolEthBalance(finalizedBlock)
			//if err != nil {
			//	log.Fatal("Could not get pool eth balance for reconciliation: ", err)
			//}

			//claimedPerAccount := onchain.GetClaimedPerWithdrawalAddress(uniqueAddresses, finalizedBlock)
			//err = oracleInstance.RunOnchainReconciliation(poolEthBalanceWei, claimedPerAccount)
			//if err != nil {
			//	log.Fatal("Reconciliation failed, state was not commited: ", err)
			//}

			err = oracleInstance.RunOffchainReconciliation()
			if err != nil {
				log.Fatal("Offchain reconciliation failed, cant freeze checkpoint: ", err)
			}

			// Freeze state
			enoughData := oracleInstance.FreezeCheckpoint()
			if !enoughData {
				log.Warn("Not enough data to create a merkle tree and hence update the contract. Skipping till next checkpoint")
				continue
			}

			// Get new state
			newState := oracleInstance.LatestCommitedState()

			// Update metrics
			metrics.KnownRootAndSlot.WithLabelValues(
				fmt.Sprintf("%d", newState.Slot),
				newState.MerkleRoot).Set(1)

			// Ensure we haven't already voted for this checkpoint. Could happen if the oracle
			// restarts before the checkpoint is consolidated. Wait while pending
			for {
				report, err := onchain.GetAddressToVotedReport(onchain.UpdaterAddress)
				if err != nil {
					log.Fatal("Could not get address to voted report: ", err)
				}

				quorum, err := onchain.GetQuorum()
				if err != nil {
					log.Fatal("Could not get quorum: ", err)
				}

				latestSlot, found := oracleInstance.LatestCommitedSlot()
				if !found {
					break
				}

				if report.Slot != latestSlot {
					break
				}

				log.Info("This oracle submited a checkpoint that is not yet consolidated, waiting to be consolidate")
				log.Info("Waiting for quorum, votes=", report.Votes, " quorum=", quorum, " on slot ", report.Slot)
				time.Sleep(1 * time.Minute)
			}

			// If so we are ready to update the contract, but multiple oracles will be racing here.
			// Lets say we have m oracles with a quorum on n (n/m). The oracles will be racing to update the root
			// and only n txs will go through and (m-n) will be reverted, as the new state will be consolidated.
			// In order to avoid txs being reverted (which costs gas), we add a random sleep between 0 and 15 minutes
			// to avoid a collision. This is not perfect, but it should be good enough. Statistically, it would be
			// very improbable that n+1 oracles will wait the same amount of time producing a collision.
			if !cfg.DryRun && enoughData {
				// Get onchain root and slot
				_, onchainSlot, err := onchain.GetOnchainSlotAndRoot()
				if err != nil {
					log.Fatal("Could not get onchain slot and root: ", err)
				}
				if newState.Slot == (onchainSlot + cfg.CheckPointSizeInSlots) {
					r := rand.Intn(16 * 60)
					log.Info("Waiting ", r, " seconds before updating the contract")
					time.Sleep(time.Duration(r) * time.Second)
				}
			}

			// Get onchain root and slot
			onchainRoot, onchainSlot, err := onchain.GetOnchainSlotAndRoot()
			if err != nil {
				log.Fatal("Could not get onchain slot and root: ", err)
			}

			log.WithFields(log.Fields{
				"OnchainRoot": onchainRoot,
				"OnchainSlot": onchainSlot,
				"OracleRoot":  newState.MerkleRoot,
				"OracleSlot":  newState.Slot,
			}).Info("Local vs onchain roots and slots")

			// If the oracle has permission to update the contract root (!dryRun), we have enough data
			// to construct a merkle tree.
			if !cfg.DryRun && enoughData {
				// If the new state is the one onchain + checkpoint size then its time to update the root
				// Then we can update the new merkle root. onchainSlot == 0 is an special case when the
				// contract was just initialized and there is no root yet.
				if newState.Slot == onchainSlot+cfg.CheckPointSizeInSlots {
					log.WithFields(log.Fields{
						"Root": newState.MerkleRoot,
						"Slot": newState.Slot,
					}).Info("Updating contract with parameters")
					err := onchain.UpdateContractMerkleRoot(newState.Slot, newState.MerkleRoot)
					if err != nil {
						// There is a very improbable case that this tx is expected to fail. If quorum is n for
						// m oracles, if n+1 oracles submit the tx at the same time, the last tx will revert.
						// In this case it would be expected to fail, but note that the above delay should
						// prevent this from happening.
						// Example: "DappnodeSmoothingPool::submitReport: Slot number invalid"
						if strings.Contains(err.Error(), "DappnodeSmoothingPool::submitReport: Slot number invalid") {
							log.WithFields(log.Fields{
								"Error": err,
								"Root":  newState.MerkleRoot,
								"Slot":  newState.Slot,
							}).Warn("Could not update contract merkle root. Expected if the state was just consolidated: ", err)
						} else {
							log.Fatal("Could not update contract merkle root: ", err)
						}
					}

					// Wait until the state we submitted is consolidated in the contract
					for {
						onchainRoot, onchainSlot, err = onchain.GetOnchainSlotAndRoot()
						if err != nil {
							log.Fatal("Could not get onchain slot and root: ", err)
						}

						if onchainRoot == newState.MerkleRoot && onchainSlot == newState.Slot {
							log.WithFields(log.Fields{
								"OnchainRoot": onchainRoot,
								"OnchainSlot": onchainSlot,
								"OracleRoot":  newState.MerkleRoot,
								"OracleSlot":  newState.Slot,
							}).Info("The submitted state is now consolidated in the contract")
							break
						} else {
							log.Info("Submitted merkle root is not consolidated, waiting for other oracles to update it")
							time.Sleep(1 * time.Minute)
						}
					}
				} else {
					log.WithFields(log.Fields{
						"Root": newState.MerkleRoot,
						"Slot": newState.Slot,
					}).Info("Skipping submitting new merkle root, already consolidated")
				}
			}

			// Persist new state in file only if everything went fine
			err = oracleInstance.SaveToJson(true)
			if err != nil {
				log.Error("Could not save state to json: ", err)
			} else {
				log.Info("State saved to json")
			}
		}
	}
}
