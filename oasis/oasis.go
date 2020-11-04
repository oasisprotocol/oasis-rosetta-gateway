package oasis

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	cmnGrpc "github.com/oasisprotocol/oasis-core/go/common/grpc"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	control "github.com/oasisprotocol/oasis-core/go/control/api"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// LatestHeight can be used as the height in queries to specify the latest
// height.
const LatestHeight = consensus.HeightLatest

// GrpcAddrEnvVar is the name of the environment variable that specifies the
// gRPC host address of the Oasis node that the client should connect to.
const GrpcAddrEnvVar = "OASIS_NODE_GRPC_ADDR"

var logger = logging.GetLogger("oasis")

// Client can be used to query an Oasis node for information and to submit
// transactions.
type Client interface {
	// GetChainID returns the network chain context, derived from the genesis
	// document.
	GetChainID(ctx context.Context) (string, error)

	// GetBlock returns the Oasis block at given height.
	GetBlock(ctx context.Context, height int64) (*Block, error)

	// GetLatestBlock returns latest Oasis block.
	GetLatestBlock(ctx context.Context) (*Block, error)

	// GetGenesisBlock returns the Oasis genesis block.
	GetGenesisBlock(ctx context.Context) (*Block, error)

	// GetAccount returns the Oasis staking account for given owner address
	// at given height.
	GetAccount(ctx context.Context, height int64, owner staking.Address) (*staking.Account, error)

	// GetDelegations returns the staking active delegations where the given
	// owner address is the delegator, as of given height.
	GetDelegations(
		ctx context.Context, height int64, owner staking.Address,
	) (map[staking.Address]*staking.Delegation, error)

	// GetDebondingDelegations returns the staking debonding delegations where
	// the given owner address is the delegator, as of given height.
	GetDebondingDelegations(
		ctx context.Context, height int64, owner staking.Address,
	) (map[staking.Address][]*staking.DebondingDelegation, error)

	// GetTransactions returns Oasis consensus transactions at given height.
	GetTransactionsWithResults(
		ctx context.Context, height int64,
	) (*consensus.TransactionsWithResults, error)

	// GetUnconfirmedTransactions returns a list of transactions currently in
	// the local node's mempool. These have not yet been included in a block.
	GetUnconfirmedTransactions(ctx context.Context) ([][]byte, error)

	// GetStakingEvents returns Oasis staking events at given height.
	GetStakingEvents(ctx context.Context, height int64) ([]*staking.Event, error)

	// SubmitTxNoWait submits the given signed transaction to the node.
	SubmitTxNoWait(ctx context.Context, tx *transaction.SignedTransaction) error

	// GetNextNonce returns the nonce that should be used when signing the next
	// transaction for the given account address at given height.
	GetNextNonce(ctx context.Context, addr staking.Address, height int64) (uint64, error)

	// GetStatus returns the status overview of the node.
	GetStatus(ctx context.Context) (*control.Status, error)
}

// Block is a representation of the Oasis block metadata, converted to be more
// compatible with the Rosetta API.
type Block struct {
	Height       int64  // Block height.
	Hash         string // Block hash.
	Timestamp    int64  // UNIX time, converted to milliseconds.
	ParentHeight int64  // Height of parent block.
	ParentHash   string // Hash of parent block.
}

// grpcClient is an implementation of Client using gRPC.
type grpcClient struct {
	sync.RWMutex

	// Connection to an Oasis node's internal socket.
	grpcConn *grpc.ClientConn

	// Cached chain ID.
	chainID string

	// Cached genesis height.
	genesisHeight int64
}

// connect() returns a gRPC connection to Oasis node via its internal socket.
// The connection is cached in the grpcClient struct and re-established
// automatically by this method if it gets shut down.
func (c *grpcClient) connect(ctx context.Context) (*grpc.ClientConn, error) {
	c.Lock()
	defer c.Unlock()

	// Check if the existing connection is good.
	if c.grpcConn != nil && c.grpcConn.GetState() != connectivity.Shutdown {
		// Return existing connection.
		return c.grpcConn, nil
	}

	// Connection needs to be re-established.
	c.grpcConn = nil

	// Get gRPC host address from environment variable.
	grpcAddr := os.Getenv(GrpcAddrEnvVar)
	if grpcAddr == "" {
		return nil, fmt.Errorf("%s environment variable not specified", GrpcAddrEnvVar)
	}

	// Establish new gRPC connection.
	var err error
	logger.Debug("Establishing connection", "grpc_addr", grpcAddr)
	c.grpcConn, err = cmnGrpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		logger.Debug("Failed to establish connection",
			"grpc_addr", grpcAddr,
			"err", err,
		)
		return nil, fmt.Errorf("failed to dial gRPC connection to '%s': %v", grpcAddr, err)
	}

	// Cache genesis height.
	status, err := control.NewNodeControllerClient(c.grpcConn).GetStatus(ctx)
	if err != nil {
		logger.Debug("Failed to get status from node", "err", err)
		return nil, fmt.Errorf("failed to get status from node: %v", err)
	}
	c.genesisHeight = status.Consensus.GenesisHeight

	return c.grpcConn, nil
}

func (c *grpcClient) GetChainID(ctx context.Context) (string, error) {
	// Return cached chain ID if we already have it.
	c.RLock()
	cid := c.chainID
	c.RUnlock()
	if cid != "" {
		return cid, nil
	}

	conn, err := c.connect(ctx)
	if err != nil {
		return "", err
	}

	c.Lock()
	defer c.Unlock()

	client := consensus.NewConsensusClient(conn)
	genesis, err := client.GetGenesisDocument(ctx)
	if err != nil {
		logger.Debug("GetChainID: failed to get genesis document", "err", err)
		return "", err
	}
	c.chainID = genesis.ChainContext()
	return c.chainID, nil
}

func (c *grpcClient) GetBlock(ctx context.Context, height int64) (*Block, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := consensus.NewConsensusClient(conn)
	blk, err := client.GetBlock(ctx, height)
	if err != nil {
		logger.Debug("GetBlock: failed to get block",
			"height", height,
			"err", err,
		)
		return nil, err
	}

	parentHeight := blk.Height - 1
	var parentHash []byte
	if parentHeight <= 0 {
		parentHeight = 1
	}
	if parentHeight < c.genesisHeight {
		parentHeight = c.genesisHeight
	}

	parentBlk, err := client.GetBlock(ctx, parentHeight)
	if err != nil {
		return nil, err
	}
	parentHeight = parentBlk.Height
	parentHash = parentBlk.Hash

	return &Block{
		Height:       blk.Height,
		Hash:         hex.EncodeToString(blk.Hash),
		Timestamp:    blk.Time.UnixNano() / 1000000, // ms
		ParentHeight: parentHeight,
		ParentHash:   hex.EncodeToString(parentHash),
	}, nil
}

func (c *grpcClient) GetLatestBlock(ctx context.Context) (*Block, error) {
	return c.GetBlock(ctx, consensus.HeightLatest)
}

func (c *grpcClient) GetGenesisBlock(ctx context.Context) (*Block, error) {
	return c.GetBlock(ctx, c.genesisHeight)
}

func (c *grpcClient) GetAccount(ctx context.Context, height int64, owner staking.Address) (*staking.Account, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	return client.Account(ctx, &staking.OwnerQuery{
		Height: height,
		Owner:  owner,
	})
}

func (c *grpcClient) GetDelegations(
	ctx context.Context,
	height int64,
	owner staking.Address,
) (map[staking.Address]*staking.Delegation, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	return client.Delegations(ctx, &staking.OwnerQuery{
		Height: height,
		Owner:  owner,
	})
}

func (c *grpcClient) GetDebondingDelegations(
	ctx context.Context,
	height int64,
	owner staking.Address,
) (map[staking.Address][]*staking.DebondingDelegation, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	return client.DebondingDelegations(ctx, &staking.OwnerQuery{
		Height: height,
		Owner:  owner,
	})
}

func (c *grpcClient) GetTransactionsWithResults(
	ctx context.Context,
	height int64,
) (*consensus.TransactionsWithResults, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := consensus.NewConsensusClient(conn)
	return client.GetTransactionsWithResults(ctx, height)
}

func (c *grpcClient) GetUnconfirmedTransactions(ctx context.Context) ([][]byte, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := consensus.NewConsensusClient(conn)
	return client.GetUnconfirmedTransactions(ctx)
}

func (c *grpcClient) GetStakingEvents(ctx context.Context, height int64) ([]*staking.Event, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	return client.GetEvents(ctx, height)
}

func (c *grpcClient) SubmitTxNoWait(ctx context.Context, tx *transaction.SignedTransaction) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	client := consensus.NewConsensusClient(conn)
	return client.SubmitTxNoWait(ctx, tx)
}

func (c *grpcClient) GetNextNonce(ctx context.Context, addr staking.Address, height int64) (uint64, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return 0, err
	}
	client := consensus.NewConsensusClient(conn)
	return client.GetSignerNonce(ctx, &consensus.GetSignerNonceRequest{
		AccountAddress: addr,
		Height:         height,
	})
}

func (c *grpcClient) GetStatus(ctx context.Context) (*control.Status, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := control.NewNodeControllerClient(conn)
	status, err := client.GetStatus(ctx)
	if err != nil {
		c.genesisHeight = status.Consensus.GenesisHeight
	}
	return status, err
}

// New creates a new Oasis gRPC client.
func New() (Client, error) {
	return &grpcClient{}, nil
}
