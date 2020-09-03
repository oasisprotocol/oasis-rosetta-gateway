package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/logging"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
)

var loggerBlk = logging.GetLogger("services/block")

type blockAPIService struct {
	oasisClient oc.OasisClient
}

// NewBlockAPIService creates a new instance of an AccountAPIService.
func NewBlockAPIService(oasisClient oc.OasisClient) server.BlockAPIServicer {
	return &blockAPIService{
		oasisClient: oasisClient,
	}
}

// Block implements the /block endpoint.
func (s *blockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerBlk.Error("Block: network validation failed", "err", terr.Message)
		return nil, terr
	}

	height := oc.LatestHeight

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index != nil {
			height = *request.BlockIdentifier.Index
		} else if request.BlockIdentifier.Hash != nil {
			loggerBlk.Error("Block: must query block by index")
			return nil, ErrMustQueryByIndex
		}
	}

	blk, err := s.oasisClient.GetBlock(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get block",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetBlk
	}

	td := newTransactionsDecoder()

	txsWithRes, err := s.oasisClient.GetTransactionsWithResults(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get transactions",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}
	for i, res := range txsWithRes.Results {
		rawTx := txsWithRes.Transactions[i]

		if err := td.DecodeTx(rawTx, res); err != nil {
			loggerBlk.Warn("Block: malformed transaction",
				"height", height,
				"index", i,
				"raw_tx", rawTx,
				"err", err,
			)
			continue
		}
	}

	evts, err := s.oasisClient.GetStakingEvents(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get staking events",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}

	var blkHash hash.Hash
	_ = blkHash.UnmarshalHex(blk.Hash)

	if err = td.DecodeBlock(blkHash, evts); err != nil {
		loggerBlk.Error("Block: unable to decode block events",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}

	tblk := &types.Block{
		BlockIdentifier: &types.BlockIdentifier{
			Index: blk.Height,
			Hash:  blk.Hash,
		},
		ParentBlockIdentifier: &types.BlockIdentifier{
			Index: blk.ParentHeight,
			Hash:  blk.ParentHash,
		},
		Timestamp:    blk.Timestamp,
		Transactions: td.Transactions(),
	}

	resp := &types.BlockResponse{
		Block: tblk,
	}

	jr, _ := json.Marshal(resp)
	loggerBlk.Debug("Block OK", "response", jr)

	return resp, nil
}

// BlockTransaction implements the /block/transaction endpoint.
// Note: we don't implement this, since we already return all transactions
// in the /block endpoint reponse above.
func (s *blockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	loggerBlk.Error("BlockTransaction: not implemented")
	return nil, ErrNotImplemented
}
