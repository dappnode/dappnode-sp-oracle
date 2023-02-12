package main

import (
	// TODO: Change when pushed
	//"github.com/dappnode/mev-sp-oracle/config"
	//"github.com/dappnode/mev-sp-oracle/oracle"
	"context"
	"mev-sp-oracle/api"
	"mev-sp-oracle/config"
	"mev-sp-oracle/oracle"
	"mev-sp-oracle/postgres"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// Hardcoded for Ethereum
var SlotsInEpoch = uint64(32)

// Examples:
// Goerli/Prater
// ./mev-sp-oracle --consensus-endpoint="http://127.0.0.1:5051" --execution-endpoint="http://127.0.0.1:8545" --deployed-slot=4500000 --pool-address="0x455e5aa18469bc6ccef49594645666c587a3a71b" --checkpoint-size=10
func main() {
	log.Info("mev-sp-oracle")
	cfg, err := config.NewCliConfig()
	if err != nil {
		log.Fatal(err)
	}

	fetcher := oracle.NewFetcher(*cfg)
	oracle := oracle.NewOracle(cfg, fetcher)
	api := api.NewApiService(*cfg)
	go api.StartHTTPServer()

	// Preparae the database
	// TODO: Dirty, to be safe. Clean db at startup until we can safely resume. The idea is
	// to resume from the last checkpoint.
	_, err = oracle.Postgres.Db.Exec(context.Background(), "drop table if exists t_oracle_validator_balances")
	if err != nil {
		log.Fatal("error cleaning table t_oracle_validator_balances at startup: ", err)
	}

	_, err = oracle.Postgres.Db.Exec(context.Background(), "drop table if exists t_pool_blocks")
	if err != nil {
		log.Fatal("error cleaning table t_pool_blocks at startup: ", err)
	}

	_, err = oracle.Postgres.Db.Exec(context.Background(), "drop table if exists t_oracle_depositaddress_rewards")
	if err != nil {
		log.Fatal("error cleaning table t_pool_blocks at startup: ", err)
	}

	if _, err := oracle.Postgres.Db.Exec(
		context.Background(),
		postgres.CreateRewardsTable); err != nil {
		log.Fatal("error creating table t_oracle_validator_balances: ", err)
	}

	if _, err := oracle.Postgres.Db.Exec(
		context.Background(),
		postgres.CreateDepositAddressRewardsTable); err != nil {
		log.Fatal("error creating table t_oracle_depositaddress_rewards: ", err)
	}

	if _, err := oracle.Postgres.Db.Exec(
		context.Background(),
		postgres.CreateBlocksTable); err != nil {
		log.Fatal("error creating table t_pool_blocks ", err)
	}

	go mainLoop(oracle, fetcher, cfg)

	// Wait for signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	for {
		sig := <-sigCh
		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == os.Interrupt || sig == os.Kill {
			break
		}
	}

	// TODO: Dump to file before stopping
	log.Info("Stopping mev-sp-oracle")
}

func mainLoop(oracle *oracle.Oracle, fetcher *oracle.Fetcher, cfg *config.Config) {
	/*
		syncProgress, err := fetcher.ExecutionClient.SyncProgress(context.Background())
		if err != nil {
			log.Error(err)
		}
	*/

	// TODO: resume from file
	log.Info("Starting to process from slot: ", oracle.State.Slot)

	for {

		headSlot, err := fetcher.ConsensusClient.NodeSyncing(context.Background())
		if err != nil {
			log.Error("Could not get node sync status:", err)
			time.Sleep(15 * time.Second)
			continue
		}

		if headSlot.IsSyncing {
			log.Error("Node is not in sync")
			time.Sleep(15 * time.Second)
			continue
		}

		finality, err := fetcher.ConsensusClient.Finality(context.Background(), "finalized")
		if err != nil {
			log.Error("Could not get finalized status:", err)
			time.Sleep(15 * time.Second)
			continue
		}

		finalizedEpoch := uint64(finality.Finalized.Epoch)
		finalizedSlot := finalizedEpoch * SlotsInEpoch

		if finalizedSlot > oracle.State.Slot {
			err = oracle.AdvanceStateToNextEpoch()
			if err != nil {
				log.Fatal(err)
			}
			log.Info("[", oracle.State.Slot, "/", finalizedSlot, "] Done processing slot. Remaining slots: ", finalizedSlot-oracle.State.Slot)
		} else {
			log.Info("Waiting for new finalized slot")
			time.Sleep(15 * time.Second)
		}

		// TODO: Rethink this a bit. Do not run in the first block we process, and think about edge cases
		if (oracle.State.Slot-cfg.DeployedSlot)%cfg.CheckPointSizeInSlots == 0 {
			log.Info("Checkpoint reached, slot: ", oracle.State.Slot)
			err, mRoot := oracle.State.DumpOracleStateToDatabase()
			if !cfg.DryRun {
				oracle.Operations.UpdateContractMerkleRoot(mRoot)
			}
			// TODO: By now just panic
			if err != nil {
				log.Fatal("Failed dumping oracle state to db: ", err)
			}
			oracle.State.LogClaimableBalances()
			oracle.State.LogPendingBalances()
		}
	}
}
