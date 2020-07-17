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
	"github.com/oasisprotocol/oasis-core/go/common/logging"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
)

// GatewayPortEnvVar is the name of the environment variable that specifies
// which port the Oasis Rosetta gateway should run on.
const GatewayPortEnvVar = "OASIS_ROSETTA_GATEWAY_PORT"

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

	return server.NewRouter(networkAPIController, accountAPIController, blockAPIController, constructionAPIController), nil
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
	oasisClient, err := oasis_client.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to prepare Oasis gRPC client: %v\n", err)
		os.Exit(1)
	}

	// Make a test request using the client to see if the node works.
	cid, err := oasisClient.GetChainID(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Node connectivity error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging.
	err = logging.Initialize(os.Stdout, logging.FmtLogfmt, logging.LevelDebug, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to initialize logging: %v\n", err)
		os.Exit(1)
	}

	logger.Info("connected to Oasis node", "chain_context", cid)

	// Start the server.
	router, err := NewBlockchainRouter(oasisClient)
	if err != nil {
		logger.Error("unable to create Rosetta blockchain router", "err", err)
		os.Exit(1)
	}
	logger.Info("Oasis Rosetta Gateway listening", "port", nPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", nPort), router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oasis Rosetta Gateway server exited with error: %v\n", err)
		os.Exit(1)
	}
}
