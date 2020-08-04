package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// OpTransfer is the Transfer operation.
const OpTransfer = "Transfer"

// OpBurn is the Burn operation.
const OpBurn = "Burn"

// OpStatusOK is the OK status.
const OpStatusOK = "OK"

// SupportedOperationTypes is a list of the supported operations.
var SupportedOperationTypes = []string{
	OpTransfer,
	OpBurn,
}

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

// Helper for making ops in a succinct way.
func appendOp(ops []*types.Operation, kind string, acct string, subacct *types.SubAccountIdentifier, amt string) []*types.Operation {
	opidx := int64(len(ops))
	op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: opidx,
		},
		Type:   kind,
		Status: OpStatusOK,
		Account: &types.AccountIdentifier{
			Address:    acct,
			SubAccount: subacct,
		},
		Amount: &types.Amount{
			Value:    amt,
			Currency: OasisCurrency,
		},
	}

	// Add related operation if it exists.
	if opidx >= 1 {
		op.RelatedOperations = []*types.OperationIdentifier{
			&types.OperationIdentifier{
				Index: opidx - 1,
			},
		}
	}

	return append(ops, op)
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

	evts, err := s.oasisClient.GetStakingEvents(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get transactions",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}

	// We group transactions by hash and number the operations within a
	// transaction in the order as they appear.
	txns := []*types.Transaction{}
	txnmap := make(map[string]uint) // hash -> index into txns
	for _, evt := range evts {
		// Get index of existing transaction if it already exists or
		// create a new one and return its index if it doesn't.
		var txidx uint
		txhash := evt.TxHash.String()
		if idx, exists := txnmap[txhash]; exists {
			txidx = idx
		} else {
			txidx = uint(len(txns))
			txnmap[txhash] = txidx
			txns = append(txns, &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: txhash,
				},
				Operations: []*types.Operation{},
			})
		}

		switch {
		case evt.Transfer != nil:
			txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, evt.Transfer.From.String(), nil, "-"+evt.Transfer.Tokens.String())
			txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, evt.Transfer.To.String(), nil, evt.Transfer.Tokens.String())
		case evt.Burn != nil:
			txns[txidx].Operations = appendOp(txns[txidx].Operations, OpBurn, evt.Burn.Owner.String(), nil, "-"+evt.Burn.Tokens.String())
		case evt.Escrow != nil:
			ee := evt.Escrow
			switch {
			case ee.Add != nil:
				// Owner's general account -> escrow account.
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, ee.Add.Owner.String(), nil, "-"+ee.Add.Tokens.String())
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, ee.Add.Escrow.String(), &types.SubAccountIdentifier{Address: SubAccountEscrow}, ee.Add.Tokens.String())
			case ee.Take != nil:
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, ee.Take.Owner.String(), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Take.Tokens.String())
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, staking.CommonPoolAddress.String(), nil, ee.Take.Tokens.String())
			case ee.Reclaim != nil:
				// Escrow account -> owner's general account.
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, ee.Reclaim.Escrow.String(), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Reclaim.Tokens.String())
				txns[txidx].Operations = appendOp(txns[txidx].Operations, OpTransfer, ee.Reclaim.Owner.String(), nil, ee.Reclaim.Tokens.String())
			}
		}
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
		Transactions: txns,
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
