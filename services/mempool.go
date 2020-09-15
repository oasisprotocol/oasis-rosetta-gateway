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

var loggerMempool = logging.GetLogger("services/mempool")

type mempoolAPIService struct {
	oasisClient oc.OasisClient
}

// NewMempoolAPIService creates a new instance of a NetworkAPIService.
func NewMempoolAPIService(oasisClient oc.OasisClient) server.MempoolAPIServicer {
	return &mempoolAPIService{
		oasisClient: oasisClient,
	}
}

// Mempool implements the /mempool endpoint.
func (s *mempoolAPIService) Mempool(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerMempool.Error("Mempool: network validation failed", "err", terr.Message)
		return nil, terr
	}

	txs, err := s.oasisClient.GetUnconfirmedTransactions(ctx)
	if err != nil {
		loggerMempool.Error("Mempool: unable to get unconfirmed transactions", "err", err)
		return nil, ErrUnableToGetTxns
	}

	var tids []*types.TransactionIdentifier
	for _, tx := range txs {
		tids = append(tids, &types.TransactionIdentifier{
			Hash: hash.NewFromBytes(tx).String(),
		})
	}

	resp := &types.MempoolResponse{
		TransactionIdentifiers: tids,
	}

	jr, _ := json.Marshal(resp)
	loggerMempool.Debug("Mempool OK", "response", jr)

	return resp, nil
}

// MempoolTransaction implements the /mempool/transaction endpoint.
func (s *mempoolAPIService) MempoolTransaction(
	ctx context.Context,
	request *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerMempool.Error("MempoolTransaction: network validation failed", "err", terr.Message)
		return nil, terr
	}

	txs, err := s.oasisClient.GetUnconfirmedTransactions(ctx)
	if err != nil {
		loggerMempool.Error("MempoolTransaction: unable to get unconfirmed transactions", "err", err)
		return nil, ErrUnableToGetTxns
	}

	var foundTx []byte
	for _, tx := range txs {
		if hash.NewFromBytes(tx).String() == request.TransactionIdentifier.Hash {
			foundTx = tx
			break
		}
	}
	if foundTx == nil {
		return nil, ErrTransactionNotFound
	}

	td := newTransactionsDecoder()
	if err = td.DecodeTx(foundTx, nil); err != nil {
		loggerMempool.Error("MempoolTransaction: unable to decode unconfirmed transaction", "err", err)
		return nil, ErrUnableToGetTxns
	}

	resp := &types.MempoolTransactionResponse{
		Transaction: td.Transactions()[0],
	}

	jr, _ := json.Marshal(resp)
	loggerMempool.Debug("MempoolTransaction OK", "response", jr)

	return resp, nil
}
