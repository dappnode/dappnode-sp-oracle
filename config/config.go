package config

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

type CliConfig struct {
	DryRun              bool
	UpdaterKeyFile      string
	UpdaterKeyPass      string
	NumRetries          int
	ConsensusEndpoint   string
	ExecutionEndpoint   string
	PoolAddress         string
	EigenManagerAddress string
	LogLevel            string
	ApiPort             int
	MetricsPort         int
	CheckPointSyncUrl   string
	RelayersEndpoints   []string
}

// By default the release is a custom build. CI takes care of upgrading it with
// go build -v -ldflags="-X 'github.com/dappnode/mev-sp-oracle/config.ReleaseVersion=x.y.z'"
var ReleaseVersion = "custom-build-your-own-risk"

func NewCliConfig() (*CliConfig, error) {
	// Optional flags:
	var version = flag.Bool("version", false, "Prints the release version and exits")
	var dryRun = flag.Bool("dry-run", false, "If enabled, the pool contract will not be updated")
	var updaterKeystoreFile = flag.String("updater-keystore-file", "", "Password protected keystore file of the updater")
	var updaterKeystorePass = flag.String("updater-keystore-pass", "", "Password of the updater keystore file")
	var numRetries = flag.Int("num-retries", 0, "Number of retries for each interaction (consensus, execution): 0 infinite")
	var logLevel = flag.String("log-level", "info", "Logging verbosity (trace, debug, info=default, warn, error, fatal, panic)")
	var apiPort = flag.Int("api-port", 7300, "Port for the API server")
	var metricsPort = flag.Int("metrics-port", 8008, "Port for the metrics server")
	var checkPointSyncUrl = flag.String("checkpoint-sync-url", "", "URL for the checkpoint sync server: http://url:port/state")

	// Mandatory flags:
	var consensusEndpoint = flag.String("consensus-endpoint", "", "Ethereum consensus endpoint")
	var executionEndpoint = flag.String("execution-endpoint", "", "Ethereum execution endpoint")
	var poolAddress = flag.String("pool-address", "", "Address of the smoothing pool contract")
	var eigenManagerAddress = flag.String("eigen-manager-address", "", "Address of the eigen manager contract")
	var relayersEndpointsStr = flag.String("relayers-endpoints", "", "Comma-separated list of relayers endpoints")

	flag.Parse()

	if *version {
		log.Info("Version: ", ReleaseVersion)
		os.Exit(0)
	}

	// Some simple cli argument validation

	if !*dryRun && *updaterKeystoreFile == "" {
		return nil, errors.New("you must provide a keystore file to update the contract root")
	}

	if !*dryRun && *updaterKeystorePass == "" {
		return nil, errors.New("you must provide a password for the keystore file")
	}

	if *dryRun && *updaterKeystoreFile != "" {
		return nil, errors.New("you can't provide a keystore file in dry run mode")
	}

	if *dryRun && *updaterKeystorePass != "" {
		return nil, errors.New("you can't provide a password for the keystore file in dry run mode")
	}

	if !common.IsHexAddress(*poolAddress) {
		return nil, errors.New("pool-address: " + *poolAddress + " is not a valid address")
	}

	if !common.IsHexAddress(*eigenManagerAddress) {
		return nil, errors.New("eigen-manager-address: " + *eigenManagerAddress + " is not a valid address")
	}

	// Post process the relayers endpoints, make it a slice
	relayersEndpoints := strings.Split(*relayersEndpointsStr, ",")

	if len(relayersEndpoints) == 0 || (len(relayersEndpoints) == 1 && relayersEndpoints[0] == "") {
		return nil, errors.New("relayers-endpoints is a mandatory flag and cant be empty")
	}

	// Validate the relayers endpoints, they must be valid URLs, not empty and start with https://.
	for _, endpoint := range relayersEndpoints {
		if endpoint == "" {
			return nil, errors.New("relayer endpoint URL cannot be empty")
		}
		if !strings.HasPrefix(endpoint, "https://") {
			return nil, errors.New("relayer endpoint URL must start with 'https://'")
		}
		if _, err := url.Parse(endpoint); err != nil {
			return nil, errors.New("invalid relayer endpoint URL: " + endpoint)
		}
	}

	cliConf := &CliConfig{
		DryRun:              *dryRun,
		UpdaterKeyFile:      *updaterKeystoreFile,
		UpdaterKeyPass:      *updaterKeystorePass,
		NumRetries:          *numRetries,
		ConsensusEndpoint:   *consensusEndpoint,
		ExecutionEndpoint:   *executionEndpoint,
		PoolAddress:         *poolAddress,
		EigenManagerAddress: *eigenManagerAddress,
		LogLevel:            *logLevel,
		ApiPort:             *apiPort,
		MetricsPort:         *metricsPort,
		CheckPointSyncUrl:   *checkPointSyncUrl,
		RelayersEndpoints:   relayersEndpoints,
	}
	logConfig(cliConf)
	return cliConf, nil
}

func logConfig(cfg *CliConfig) {
	log.WithFields(log.Fields{
		"DryRun":              cfg.DryRun,
		"UpdaterKeyFile":      cfg.UpdaterKeyFile,
		"UpdaterKeyPass":      "hidden",
		"NumRetries":          cfg.NumRetries,
		"ConsensusEndpoint":   cfg.ConsensusEndpoint,
		"ExecutionEndpoint":   cfg.ExecutionEndpoint,
		"PoolAddress":         cfg.PoolAddress,
		"EigenManagerAddress": cfg.EigenManagerAddress,
		"LogLevel":            cfg.LogLevel,
		"ApiPort":             cfg.ApiPort,
		"MetricsPort":         cfg.MetricsPort,
		"CheckPointSyncUrl":   cfg.CheckPointSyncUrl,
		"RelayersEndpoints":   cfg.RelayersEndpoints,
	}).Info("Cli Config:")
}
