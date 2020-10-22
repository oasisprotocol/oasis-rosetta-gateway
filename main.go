package main

import (
	"context"
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

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
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

var logger = logging.GetLogger("oasis-rosetta-gateway")

// NewBlockchainRouter returns a Mux http.Handler from a collection of
// Rosetta service controllers.
func NewBlockchainRouter(oasisClient oasis_client.OasisClient) (http.Handler, error) {
	chainID, err := oasisClient.GetChainID(context.Background())
	if err != nil {
		return nil, err
	}

	asserter, err := asserter.NewServer(
		services.SupportedOperationTypes,
		true,
		[]*types.NetworkIdentifier{
			&types.NetworkIdentifier{
				Blockchain: services.OasisBlockchainName,
				Network:    chainID,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	networkAPIController := server.NewNetworkAPIController(services.NewNetworkAPIService(oasisClient), asserter)
	accountAPIController := server.NewAccountAPIController(services.NewAccountAPIService(oasisClient), asserter)
	blockAPIController := server.NewBlockAPIController(services.NewBlockAPIService(oasisClient), asserter)
	constructionAPIController := server.NewConstructionAPIController(services.NewConstructionAPIService(oasisClient), asserter)
	mempoolAPIController := server.NewMempoolAPIController(services.NewMempoolAPIService(oasisClient), asserter)

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
			&types.NetworkIdentifier{
				Blockchain: services.OasisBlockchainName,
				Network:    chainID,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	constructionAPIController := server.NewConstructionAPIController(services.NewConstructionAPIService(nil), asserter)

	return server.NewRouter(constructionAPIController), nil
}

func main() {
	// Get server port from environment variable or use the default.
	port := os.Getenv(GatewayPortEnvVar)
	if port == "" {
		port = "8080"
	}
	nPort, err := strconv.Atoi(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Malformed %s environment variable: %v\n", GatewayPortEnvVar, err)
		os.Exit(1)
	}

	var chainID string
	var oasisClient oasis_client.OasisClient

	// Check if we should run in offline mode.
	var offlineMode bool
	if os.Getenv(OfflineModeEnvVar) != "" {
		offlineMode = true

		// Also get chain ID.
		chainID = os.Getenv(services.OfflineModeChainIDEnvVar)
		if chainID == "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s environment variable missing\n", services.OfflineModeChainIDEnvVar)
			os.Exit(1)
		}
	}

	if !offlineMode {
		// Wait for the node socket to appear.
		addr := os.Getenv(oasis_client.GrpcAddrEnvVar)
		if addr == "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s environment variable missing\n", oasis_client.GrpcAddrEnvVar)
			os.Exit(1)
		}
		if strings.HasPrefix(addr, "unix:") {
			sock := strings.Split(addr, ":")[1]
			_, grr := os.Stat(sock)
			for os.IsNotExist(grr) {
				logger.Info("waiting for node socket to appear...", "socket_path", sock)
				time.Sleep(1 * time.Second)
				_, grr = os.Stat(sock)
			}
		}

		// Prepare a new Oasis gRPC client.
		oasisClient, err = oasis_client.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to prepare Oasis gRPC client: %v\n", err)
			os.Exit(1)
		}

		// Make a test request using the client to see if the node works.
		chainID, err = oasisClient.GetChainID(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Node connectivity error: %v\n", err)
			os.Exit(1)
		}
	}

	// Set the chain context for preparing signing payloads.
	signature.SetChainContext(chainID)

	// Initialize logging.
	err = logging.Initialize(os.Stdout, logging.FmtLogfmt, logging.LevelDebug, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to initialize logging: %v\n", err)
		os.Exit(1)
	}

	var router http.Handler
	switch {
	case offlineMode:
		logger.Info("running in offline mode", "chain_context", chainID)
		router, err = NewOfflineBlockchainRouter(chainID)
	default:
		logger.Info("connected to Oasis node", "chain_context", chainID)
		router, err = NewBlockchainRouter(oasisClient)
	}

	if err != nil {
		logger.Error("unable to create Rosetta blockchain router", "err", err)
		os.Exit(1)
	}

	// Start the server.
	logger.Info("Oasis Rosetta Gateway listening", "port", nPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", nPort), router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oasis Rosetta Gateway server exited with error: %v\n", err)
		os.Exit(1)
	}
}
