package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"

	"github.com/oasisprotocol/oasis-core/go/common/logging"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
)

// OpTransfer is the Transfer operation.
const OpTransfer = "Transfer"

// OpBurn is the Burn operation.
const OpBurn = "Burn"

// OpReclaimEscrow is the Burn operation.
const OpReclaimEscrow = "ReclaimEscrow"

// OpStatusOK is the OK status.
const OpStatusOK = "OK"

// SupportedOperationTypes is a list of the supported operations.
var SupportedOperationTypes = []string{
	OpTransfer,
	OpBurn,
	OpReclaimEscrow,
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

	// We group transactions by hash and number the operations within a
	// transaction in the order as they appear.
	txns := []*types.Transaction{}
	txnmap := make(map[string]*types.Transaction) // hash -> transaction in txns
	// getTxn get existing transaction if it already exists or
	// create a new one and return it if it doesn't.
	getTxn := func(txhash string) *types.Transaction {
		if txn, exists := txnmap[txhash]; exists {
			return txn
		}
		tx := &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: txhash,
			},
			Operations: []*types.Operation{},
		}
		txns = append(txns, tx)
		txnmap[txhash] = tx
		return tx
	}

	txsWithRes, err := s.oasisClient.GetTransactionsWithResults(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get transactions",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}
	for i, res := range txsWithRes.Results {
		if !res.IsSuccess() {
			continue
		}
		rawTx := txsWithRes.Transactions[i]
		var sigTx transaction.SignedTransaction
		if err := cbor.Unmarshal(rawTx, &sigTx); err != nil {
			loggerBlk.Warn("Block: malformed transaction",
				"height", height,
				"index", i,
				"raw_tx", rawTx,
				"err", err,
			)
			continue
		}
		var tx transaction.Transaction
		if err := sigTx.Open(&tx); err != nil {
			loggerBlk.Warn("Block: invalid transaction signature",
				"height", height,
				"index", i,
				"sig_tx", sigTx,
				"err", err,
			)
			continue
		}
		switch tx.Method {
		case staking.MethodReclaimEscrow:
			var reclaim staking.ReclaimEscrow
			if err := cbor.Unmarshal(tx.Body, &reclaim); err != nil {
				loggerBlk.Warn("Block: malformed reclaim escrow",
					"height", height,
					"index", i,
					"body", tx.Body,
					"err", err,
				)
				continue
			}
			// Emit the reclaim escrow intent.
			txn := getTxn(hash.NewFromBytes(rawTx).String())
			opidx := int64(len(txn.Operations))
			txn.Operations = append(txn.Operations,
				&types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: opidx,
					},
					Type:   OpReclaimEscrow,
					Status: OpStatusOK,
					Account: &types.AccountIdentifier{
						Address: StringFromAddress(staking.NewAddress(sigTx.Signature.PublicKey)),
					},
					Metadata: nil,
				},
				&types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: opidx + 1,
					},
					Type:   OpReclaimEscrow,
					Status: OpStatusOK,
					Account: &types.AccountIdentifier{
						Address: StringFromAddress(reclaim.Account),
						SubAccount: &types.SubAccountIdentifier{
							Address: SubAccountEscrow,
						},
					},
					Metadata: map[string]interface{}{
						ReclaimEscrowSharesKey: reclaim.Shares.String(),
					},
				},
			)
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

	for _, evt := range evts {
		txn := getTxn(evt.TxHash.String())

		switch {
		case evt.Transfer != nil:
			txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(evt.Transfer.From), nil, "-"+evt.Transfer.Amount.String())
			txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(evt.Transfer.To), nil, evt.Transfer.Amount.String())
		case evt.Burn != nil:
			txn.Operations = appendOp(txn.Operations, OpBurn, StringFromAddress(evt.Burn.Owner), nil, "-"+evt.Burn.Amount.String())
		case evt.Escrow != nil:
			ee := evt.Escrow
			switch {
			case ee.Add != nil:
				// Owner's general account -> escrow account.
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(ee.Add.Owner), nil, "-"+ee.Add.Amount.String())
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(ee.Add.Escrow), &types.SubAccountIdentifier{Address: SubAccountEscrow}, ee.Add.Amount.String())
			case ee.Take != nil:
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(ee.Take.Owner), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Take.Amount.String())
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(staking.CommonPoolAddress), nil, ee.Take.Amount.String())
			case ee.Reclaim != nil:
				// Escrow account -> owner's general account.
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(ee.Reclaim.Escrow), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Reclaim.Amount.String())
				txn.Operations = appendOp(txn.Operations, OpTransfer, StringFromAddress(ee.Reclaim.Owner), nil, ee.Reclaim.Amount.String())
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
