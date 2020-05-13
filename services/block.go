package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	oc "github.com/oasislabs/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasislabs/oasis-core/go/common/logging"
)

// OpTransferFrom is the first part of the Oasis Transfer operation.
const OpTransferFrom = "TransferFrom"

// OpTransferTo is the second part of the Oasis Transfer operation.
const OpTransferTo = "TransferTo"

// OpBurn is the Oasis Burn operation.
const OpBurn = "Burn"

// OpStatusOK is the OK status.
const OpStatusOK = "OK"

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

	evts, err := s.oasisClient.GetStakingEvents(ctx, height)
	if err != nil {
		loggerBlk.Error("Block: unable to get transactions",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetTxns
	}

	txns := []*types.Transaction{}
	for _, evt := range evts {
		switch {
		case evt.TransferEvent != nil:
			txns = append(txns, &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: evt.TxHash.String(),
				},
				Operations: []*types.Operation{
					&types.Operation{
						OperationIdentifier: &types.OperationIdentifier{
							Index: 0,
						},
						Type:   OpTransferFrom,
						Status: OpStatusOK,
						Account: &types.AccountIdentifier{
							Address: evt.TransferEvent.From.String(),
							SubAccount: &types.SubAccountIdentifier{
								Address: SubAccountGeneral,
							},
						},
						Amount: &types.Amount{
							Value:    "-" + evt.TransferEvent.Tokens.String(),
							Currency: OasisCurrency,
						},
					},
					&types.Operation{
						OperationIdentifier: &types.OperationIdentifier{
							Index: 1,
						},
						RelatedOperations: []*types.OperationIdentifier{
							&types.OperationIdentifier{
								Index: 0,
							},
						},
						Type:   OpTransferTo,
						Status: OpStatusOK,
						Account: &types.AccountIdentifier{
							Address: evt.TransferEvent.To.String(),
							SubAccount: &types.SubAccountIdentifier{
								Address: SubAccountGeneral,
							},
						},
						Amount: &types.Amount{
							Value:    evt.TransferEvent.Tokens.String(),
							Currency: OasisCurrency,
						},
					},
				},
			})
		case evt.BurnEvent != nil:
			txns = append(txns, &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: evt.TxHash.String(),
				},
				Operations: []*types.Operation{
					&types.Operation{
						OperationIdentifier: &types.OperationIdentifier{
							Index: 0,
						},
						Type:   OpBurn,
						Status: OpStatusOK,
						Account: &types.AccountIdentifier{
							Address: evt.BurnEvent.Owner.String(),
							SubAccount: &types.SubAccountIdentifier{
								Address: SubAccountGeneral,
							},
						},
						Amount: &types.Amount{
							Value:    "-" + evt.BurnEvent.Tokens.String(),
							Currency: OasisCurrency,
						},
					},
				},
			})
		case evt.EscrowEvent != nil:
			ee := evt.EscrowEvent
			// Note: These have been abstracted to use Transfer* and Burn
			// instead of creating new operations just for escrow accounts.
			// It should be evident based on the subaccount identifiers
			// if an operation is escrow-related or not.
			switch {
			case ee.Add != nil:
				// Owner's general account -> escrow account.
				txns = append(txns, &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{
						Hash: evt.TxHash.String(),
					},
					Operations: []*types.Operation{
						&types.Operation{
							OperationIdentifier: &types.OperationIdentifier{
								Index: 0,
							},
							Type:   OpTransferFrom,
							Status: OpStatusOK,
							Account: &types.AccountIdentifier{
								Address: ee.Add.Owner.String(),
								SubAccount: &types.SubAccountIdentifier{
									Address: SubAccountGeneral,
								},
							},
							Amount: &types.Amount{
								Value:    "-" + ee.Add.Tokens.String(),
								Currency: OasisCurrency,
							},
						},
						&types.Operation{
							OperationIdentifier: &types.OperationIdentifier{
								Index: 1,
							},
							RelatedOperations: []*types.OperationIdentifier{
								&types.OperationIdentifier{
									Index: 0,
								},
							},
							Type:   OpTransferTo,
							Status: OpStatusOK,
							Account: &types.AccountIdentifier{
								Address: ee.Add.Escrow.String(),
								SubAccount: &types.SubAccountIdentifier{
									Address: SubAccountEscrow,
								},
							},
							Amount: &types.Amount{
								Value:    ee.Add.Tokens.String(),
								Currency: OasisCurrency,
							},
						},
					},
				})
			case ee.Take != nil:
				txns = append(txns, &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{
						Hash: evt.TxHash.String(),
					},
					Operations: []*types.Operation{
						&types.Operation{
							OperationIdentifier: &types.OperationIdentifier{
								Index: 0,
							},
							Type:   OpBurn,
							Status: OpStatusOK,
							Account: &types.AccountIdentifier{
								Address: ee.Take.Owner.String(),
								SubAccount: &types.SubAccountIdentifier{
									Address: SubAccountEscrow,
								},
							},
							Amount: &types.Amount{
								Value:    "-" + ee.Take.Tokens.String(),
								Currency: OasisCurrency,
							},
						},
					},
				})
			case ee.Reclaim != nil:
				// Escrow account -> owner's general account.
				txns = append(txns, &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{
						Hash: evt.TxHash.String(),
					},
					Operations: []*types.Operation{
						&types.Operation{
							OperationIdentifier: &types.OperationIdentifier{
								Index: 0,
							},
							Type:   OpTransferFrom,
							Status: OpStatusOK,
							Account: &types.AccountIdentifier{
								Address: ee.Reclaim.Escrow.String(),
								SubAccount: &types.SubAccountIdentifier{
									Address: SubAccountEscrow,
								},
							},
							Amount: &types.Amount{
								Value:    "-" + ee.Reclaim.Tokens.String(),
								Currency: OasisCurrency,
							},
						},
						&types.Operation{
							OperationIdentifier: &types.OperationIdentifier{
								Index: 1,
							},
							RelatedOperations: []*types.OperationIdentifier{
								&types.OperationIdentifier{
									Index: 0,
								},
							},
							Type:   OpTransferTo,
							Status: OpStatusOK,
							Account: &types.AccountIdentifier{
								Address: ee.Reclaim.Owner.String(),
								SubAccount: &types.SubAccountIdentifier{
									Address: SubAccountGeneral,
								},
							},
							Amount: &types.Amount{
								Value:    ee.Reclaim.Tokens.String(),
								Currency: OasisCurrency,
							},
						},
					},
				})
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
