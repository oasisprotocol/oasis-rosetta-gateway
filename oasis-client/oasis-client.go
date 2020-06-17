package oasis_client

import (
	"context"
	"encoding/hex"
	"encoding/json"
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

// LatestHeight can be used as the height in queries to specify the latest height.
const LatestHeight = consensus.HeightLatest

// GenesisHeight is the height of the genesis block.
const GenesisHeight = int64(1)

// GrpcAddrEnvVar is the name of the environment variable that specifies
// the gRPC host address of the Oasis node that the client should connect to.
const GrpcAddrEnvVar = "OASIS_NODE_GRPC_ADDR"

var logger = logging.GetLogger("oasis-client")

// OasisClient can be used to query an Oasis node for information
// and to submit transactions.
type OasisClient interface {
	// GetChainID returns the network chain context, derived from the
	// genesis document.
	GetChainID(ctx context.Context) (string, error)

	// GetBlock returns the Oasis block at given height.
	GetBlock(ctx context.Context, height int64) (*OasisBlock, error)

	// GetLatestBlock returns latest Oasis block.
	GetLatestBlock(ctx context.Context) (*OasisBlock, error)

	// GetGenesisBlock returns the Oasis genesis block.
	GetGenesisBlock(ctx context.Context) (*OasisBlock, error)

	// GetAccount returns the Oasis staking account for given owner address
	// at given height.
	GetAccount(ctx context.Context, height int64, owner staking.Address) (*staking.Account, error)

	// GetStakingEvents returns Oasis staking events at given height.
	GetStakingEvents(ctx context.Context, height int64) ([]staking.Event, error)

	// SubmitTx submits the given JSON-encoded transaction to the node.
	SubmitTx(ctx context.Context, txRaw string) error

	// GetNextNonce returns the nonce that should be used when signing the
	// next transaction for the given account address at given height.
	GetNextNonce(ctx context.Context, addr staking.Address, height int64) (uint64, error)

	// GetStatus returns the status overview of the node.
	GetStatus(ctx context.Context) (*control.Status, error)
}

// OasisBlock is a representation of the Oasis block metadata,
// converted to be more compatible with the Rosetta API.
type OasisBlock struct {
	Height       int64  // Block height.
	Hash         string // Block hash.
	Timestamp    int64  // UNIX time, converted to milliseconds.
	ParentHeight int64  // Height of parent block.
	ParentHash   string // Hash of parent block.
}

// grpcOasisClient is an implementation of OasisClient using gRPC.
type grpcOasisClient struct {
	sync.RWMutex

	// Connection to an Oasis node's internal socket.
	grpcConn *grpc.ClientConn

	// Cached chain ID.
	chainID string
}

// connect() returns a gRPC connection to Oasis node via its internal socket.
// The connection is cached in the grpcOasisClient struct and re-established
// automatically by this method if it gets shut down.
func (oc *grpcOasisClient) connect(ctx context.Context) (*grpc.ClientConn, error) {
	oc.Lock()
	defer oc.Unlock()

	// Check if the existing connection is good.
	if oc.grpcConn != nil && oc.grpcConn.GetState() != connectivity.Shutdown {
		// Return existing connection.
		return oc.grpcConn, nil
	} else {
		// Connection needs to be re-established.
		oc.grpcConn = nil
	}

	// Get gRPC host address from environment variable.
	grpcAddr := os.Getenv(GrpcAddrEnvVar)
	if grpcAddr == "" {
		return nil, fmt.Errorf("%s environment variable not specified", GrpcAddrEnvVar)
	}

	// Establish new gRPC connection.
	var err error
	logger.Debug("Establishing connection", "grpc_addr", grpcAddr)
	oc.grpcConn, err = cmnGrpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		logger.Debug("Failed to establish connection",
			"grpc_addr", grpcAddr,
			"err", err,
		)
		return nil, fmt.Errorf("failed to dial gRPC connection to '%s': %v", grpcAddr, err)
	}
	return oc.grpcConn, nil
}

func (oc *grpcOasisClient) GetChainID(ctx context.Context) (string, error) {
	// Return cached chain ID if we already have it.
	oc.RLock()
	cid := oc.chainID
	oc.RUnlock()
	if cid != "" {
		return cid, nil
	}

	conn, err := oc.connect(ctx)
	if err != nil {
		return "", err
	}

	oc.Lock()
	defer oc.Unlock()

	client := consensus.NewConsensusClient(conn)
	genesis, err := client.GetGenesisDocument(ctx)
	if err != nil {
		logger.Debug("GetChainID: failed to get genesis document", "err", err)
		return "", err
	}
	oc.chainID = genesis.ChainContext()
	return oc.chainID, nil
}

func (oc *grpcOasisClient) GetBlock(ctx context.Context, height int64) (*OasisBlock, error) {
	conn, err := oc.connect(ctx)
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
		parentHeight = GenesisHeight
	}

	parentBlk, err := client.GetBlock(ctx, parentHeight)
	if err != nil {
		return nil, err
	}
	parentHeight = parentBlk.Height
	parentHash = parentBlk.Hash

	return &OasisBlock{
		Height:       blk.Height,
		Hash:         hex.EncodeToString(blk.Hash),
		Timestamp:    blk.Time.UnixNano() / 1000000, // ms
		ParentHeight: parentHeight,
		ParentHash:   hex.EncodeToString(parentHash),
	}, nil
}

func (oc *grpcOasisClient) GetLatestBlock(ctx context.Context) (*OasisBlock, error) {
	return oc.GetBlock(ctx, consensus.HeightLatest)
}

func (oc *grpcOasisClient) GetGenesisBlock(ctx context.Context) (*OasisBlock, error) {
	return oc.GetBlock(ctx, GenesisHeight)
}

func (oc *grpcOasisClient) GetAccount(ctx context.Context, height int64, owner staking.Address) (*staking.Account, error) {
	conn, err := oc.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	return client.Account(ctx, &staking.OwnerQuery{
		Height: height,
		Owner:  owner,
	})
}

func (oc *grpcOasisClient) GetStakingEvents(ctx context.Context, height int64) ([]staking.Event, error) {
	conn, err := oc.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := staking.NewStakingClient(conn)
	evts, err := client.GetEvents(ctx, height)
	if err != nil {
		return nil, err
	}
	// Change empty hashes to block hashes, as they belong to block events.
	var gotBlkHash bool
	var blkHash []byte
	for i := range evts {
		e := &evts[i]
		if e.TxHash.IsEmpty() {
			if !gotBlkHash {
				// First time, need to fetch the block hash.
				conClient := consensus.NewConsensusClient(conn)
				blk, err := conClient.GetBlock(ctx, height)
				if err != nil {
					return nil, err
				}
				blkHash = blk.Hash
				gotBlkHash = true
			}
			copy(e.TxHash[:], blkHash)
		}
	}
	return evts, nil
}

func (oc *grpcOasisClient) SubmitTx(ctx context.Context, txRaw string) error {
	conn, err := oc.connect(ctx)
	if err != nil {
		return err
	}
	client := consensus.NewConsensusClient(conn)
	var tx *transaction.SignedTransaction
	if err := json.Unmarshal([]byte(txRaw), &tx); err != nil {
		logger.Debug("SubmitTx: failed to unmarshal raw transaction", "err", err)
		return err
	}
	return client.SubmitTx(ctx, tx)
}

func (oc *grpcOasisClient) GetNextNonce(ctx context.Context, addr staking.Address, height int64) (uint64, error) {
	conn, err := oc.connect(ctx)
	if err != nil {
		return 0, err
	}
	client := consensus.NewConsensusClient(conn)
	return client.GetSignerNonce(ctx, &consensus.GetSignerNonceRequest{
		AccountAddress: addr,
		Height:         height,
	})
}

func (oc *grpcOasisClient) GetStatus(ctx context.Context) (*control.Status, error) {
	conn, err := oc.connect(ctx)
	if err != nil {
		return nil, err
	}
	client := control.NewNodeControllerClient(conn)
	return client.GetStatus(ctx)
}

// New creates a new Oasis gRPC client.
func New() (OasisClient, error) {
	return &grpcOasisClient{}, nil
}
