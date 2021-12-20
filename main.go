package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/logging"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/common"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
)

// GatewayPortEnvVar is the name of the environment variable that specifies
// which port the Oasis Rosetta gateway should run on.
const GatewayPortEnvVar = "OASIS_ROSETTA_GATEWAY_PORT"

// OfflineModeEnvVar is the name of the environment variable that specifies
// that the gateway should run in offline mode (without a connection to an
// Oasis node).  Note that only parts of the Construction API are available
// in this mode and nothing else.
// Don't forget to set services.OfflineModeChainIDEnvVar as well.
const OfflineModeEnvVar = "OASIS_ROSETTA_GATEWAY_OFFLINE_MODE"

var (
	logger = logging.GetLogger("oasis-rosetta-gateway")

	versionFlag = flag.Bool("version", false, "Print version and exit")
)

// NewBlockchainRouter returns a Mux http.Handler from a collection of
// Rosetta service controllers.
func NewBlockchainRouter(oasisClient oasis.Client) (http.Handler, error) {
	chainID, err := oasisClient.GetChainID(context.Background())
	if err != nil {
		return nil, err
	}

	asserter, err := asserter.NewServer(
		services.SupportedOperationTypes,
		true,
		[]*types.NetworkIdentifier{
			{
				Blockchain: services.OasisBlockchainName,
				Network:    chainID,
			},
		},
		nil,
		false,
		"",
	)
	if err != nil {
		return nil, err
	}

	networkAPIController := server.NewNetworkAPIController(
		services.NewNetworkAPIService(oasisClient), asserter,
	)
	accountAPIController := server.NewAccountAPIController(
		services.NewAccountAPIService(oasisClient), asserter,
	)
	blockAPIController := server.NewBlockAPIController(
		services.NewBlockAPIService(oasisClient), asserter,
	)
	constructionAPIController := server.NewConstructionAPIController(
		services.NewConstructionAPIService(oasisClient), asserter,
	)
	mempoolAPIController := server.NewMempoolAPIController(
		services.NewMempoolAPIService(oasisClient), asserter,
	)

	return server.NewRouter(
		networkAPIController,
		accountAPIController,
		blockAPIController,
		constructionAPIController,
		mempoolAPIController,
	), nil
}

// NewOfflineBlockchainRouter is the same as above, but for offline mode.
func NewOfflineBlockchainRouter(chainID string) (http.Handler, error) {
	asserter, err := asserter.NewServer(
		services.SupportedOperationTypes,
		true,
		[]*types.NetworkIdentifier{
			{
				Blockchain: services.OasisBlockchainName,
				Network:    chainID,
			},
		},
		nil,
		false,
		"",
	)
	if err != nil {
		return nil, err
	}

	constructionAPIController := server.NewConstructionAPIController(services.NewConstructionAPIService(nil), asserter)

	return server.NewRouter(constructionAPIController), nil
}

// Return the value of the given environment variable or exit if it is
// empty (or unset).
func getEnvVarOrExit(name string) string {
	value := os.Getenv(name)
	if value == "" {
		logger.Error("environment variable missing",
			"name", name,
		)
		os.Exit(1)
	}
	return value
}

// Return the server port that should be used or exit if it is malformed.
func getPortOrExit() int {
	portStr := os.Getenv(GatewayPortEnvVar)
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Error("malformed environment variable",
			"err", err,
			"name", GatewayPortEnvVar,
		)
		os.Exit(1)
	}
	return port
}

// Print version information.
func printVersionInfo() {
	fmt.Printf("Software version: %s\n", common.SoftwareVersion)
	fmt.Printf("Oasis Core:\n")
	fmt.Printf("  Software version: %s\n", common.GetOasisCoreVersion())
	fmt.Printf("Rosetta API version: %s\n", common.RosettaAPIVersion)
	fmt.Printf("Go toolchain version: %s\n", common.ToolchainVersion)
}

func main() {
	// Initialize logging.
	if err := logging.Initialize(os.Stdout, logging.FmtLogfmt, logging.LevelDebug, nil); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to initialize logging: %v\n", err)
		os.Exit(1)
	}

	// Print version info if -version flag is passed.
	flag.Parse()
	if *versionFlag {
		printVersionInfo()
		return
	}

	// Get server port.
	port := getPortOrExit()

	var chainID string
	var oasisClient oasis.Client
	var err error

	// Check if we should run in offline mode.
	offlineMode := os.Getenv(OfflineModeEnvVar) != ""

	switch offlineMode {
	case true:
		// Get chain ID.
		chainID = getEnvVarOrExit(services.OfflineModeChainIDEnvVar)

	case false:
		// Get node's Unix socket.
		addr := getEnvVarOrExit(oasis.GrpcAddrEnvVar)

		if strings.HasPrefix(addr, "unix:") {
			sock := strings.Split(addr, ":")[1]
			// Wait for node's Unix socket to appear.
			_, err2 := os.Stat(sock)
			for os.IsNotExist(err2) {
				logger.Info("waiting for node socket to appear...", "socket_path", sock)
				time.Sleep(1 * time.Second)
				_, err2 = os.Stat(sock)
			}
		}

		// Prepare a new Oasis gRPC client.
		oasisClient, err = oasis.New()
		if err != nil {
			logger.Error("failed to create Oasis gRPC client",
				"err", err,
			)
			os.Exit(1)
		}

		// Get chain ID.
		chainID, err = oasisClient.GetChainID(context.Background())
		if err != nil {
			logger.Error("failed to obtain chain ID from Oasis node",
				"err", err,
			)
			os.Exit(1)
		}
	}

	// Set the chain context for preparing signing payloads.
	signature.SetChainContext(chainID)

	var router http.Handler
	switch offlineMode {
	case true:
		logger.Info("running in offline mode", "chain_context", chainID)
		router, err = NewOfflineBlockchainRouter(chainID)
	case false:
		logger.Info("connected to Oasis node", "chain_context", chainID)
		router, err = NewBlockchainRouter(oasisClient)
	}
	if err != nil {
		logger.Error("unable to create Rosetta blockchain router", "err", err)
		os.Exit(1)
	}

	// Start the server.
	logger.Info("Oasis Rosetta Gateway listening", "port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		logger.Error("Oasis Rosetta Gateway server exited",
			"err", err,
		)
		os.Exit(1)
	}
}
