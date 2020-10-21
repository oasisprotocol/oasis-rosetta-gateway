// https://djr6hkgq2tjcs.cloudfront.net/docs/construction_api_introduction.html
package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
)

// OptionsIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the account ID.
const OptionsIDKey = "id"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "nonce"

// UnsignedTransaction is a transaction with the account that would sign it.
type UnsignedTransaction struct {
	Tx     cbor.RawMessage `json:"tx"`
	Signer string          `json:"signer"`
}

var loggerCons = logging.GetLogger("services/construction")

type constructionAPIService struct {
	oasisClient oc.OasisClient
}

// NewConstructionAPIService creates a new instance of an ConstructionAPIService.
func NewConstructionAPIService(oasisClient oc.OasisClient) server.ConstructionAPIServicer {
	return &constructionAPIService{
		oasisClient: oasisClient,
	}
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *constructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if s.oasisClient == nil {
		loggerCons.Error("ConstructionMetadata: not available in offline mode")
		return nil, ErrNotAvailableInOfflineMode
	}

	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionMetadata: network validation failed", "err", terr.Message)
		return nil, terr
	}

	// Get the account ID field from the Options object.
	if request.Options == nil {
		loggerCons.Error("ConstructionMetadata: missing options")
		return nil, ErrInvalidAccountAddress
	}
	idRaw, ok := request.Options[OptionsIDKey]
	if !ok {
		loggerCons.Error("ConstructionMetadata: account ID field not given")
		return nil, ErrInvalidAccountAddress
	}
	idString, ok := idRaw.(string)
	if !ok {
		loggerCons.Error("ConstructionMetadata: malformed account ID field")
		return nil, ErrInvalidAccountAddress
	}

	// Convert the byte value of the ID to account address.
	var owner staking.Address
	err := owner.UnmarshalText([]byte(idString))
	if err != nil {
		loggerCons.Error("ConstructionMetadata: invalid account ID", "err", err)
		return nil, ErrInvalidAccountAddress
	}

	nonce, err := s.oasisClient.GetNextNonce(ctx, owner, oc.LatestHeight)
	if err != nil {
		loggerCons.Error("ConstructionMetadata: unable to get next nonce",
			"account_id", owner.String(),
			"err", err,
		)
		return nil, ErrUnableToGetNextNonce
	}

	// Return next nonce that should be used to sign transactions for given account.
	md := make(map[string]interface{})
	md[NonceKey] = nonce

	resp := &types.ConstructionMetadataResponse{
		Metadata: md,
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionMetadata OK", "response", jr)

	return resp, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (s *constructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	if s.oasisClient == nil {
		loggerCons.Error("ConstructionSubmit: not available in offline mode")
		return nil, ErrNotAvailableInOfflineMode
	}

	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionSubmit: network validation failed", "err", terr.Message)
		return nil, terr
	}

	tx, err := DecodeSignedTransaction(request.SignedTransaction)
	if err != nil {
		loggerCons.Error("ConstructionSubmit: failed to unmarshal signed transaction",
			"err", err,
			"signed_tx", request.SignedTransaction,
		)
		return nil, ErrMalformedValue
	}

	if err := s.oasisClient.SubmitTxNoWait(ctx, tx); err != nil {
		loggerCons.Error("ConstructionSubmit: SubmitTxNoWait failed", "err", err)
		if errors.Is(err, consensus.ErrDuplicateTx) {
			loggerCons.Info("ConstructionSubmit: treating ErrDuplicateTx as success")
		} else {
			return nil, NewDetailedError(ErrUnableToSubmitTx, err)
		}
	}

	resp := &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: tx.Hash().String(),
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionSubmit OK", "response", jr)

	return resp, nil
}

// ConstructionHash implements the /construction/hash endpoint.
func (s *constructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionHash: network validation failed", "err", terr.Message)
		return nil, terr
	}

	tx, err := DecodeSignedTransaction(request.SignedTransaction)
	if err != nil {
		loggerCons.Error("ConstructionSubmit: failed to unmarshal signed transaction",
			"err", err,
			"signed_tx", request.SignedTransaction,
		)
		return nil, ErrMalformedValue
	}

	resp := &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: tx.Hash().String(),
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionHash OK", "response", jr)

	return resp, nil
}

// ConstructionDerive implements the /construction/derive endpoint.
func (s *constructionAPIService) ConstructionDerive(
	ctx context.Context,
	request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionDerive: network validation failed", "err", terr.Message)
		return nil, terr
	}

	var pk signature.PublicKey
	if err := pk.UnmarshalBinary(request.PublicKey.Bytes); err != nil {
		loggerCons.Error("ConstructionDerive: malformed public key",
			"public_key_hex_bytes", hex.EncodeToString(request.PublicKey.Bytes),
			"err", err,
		)
		return nil, ErrMalformedValue
	}

	resp := &types.ConstructionDeriveResponse{
		Address: StringFromAddress(staking.NewAddress(pk)),
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionDerive OK", "response", jr)

	return resp, nil
}

// ConstructionCombine implements the /construction/combine endpoint.
func (s *constructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionCombine: network validation failed", "err", terr.Message)
		return nil, terr
	}

	// Combine creates a network-specific transaction from an unsigned
	// transaction and an array of provided signatures. The signed
	// transaction returned from this method will be sent to the
	// `/construction/submit` endpoint by the caller.

	ut, err := DecodeUnsignedTransaction(request.UnsignedTransaction)
	if err != nil {
		loggerCons.Error("ConstructionCombine: unmarshal unsigned transaction",
			"unsigned_transaction", request.UnsignedTransaction,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	if len(request.Signatures) != 1 {
		loggerCons.Error("ConstructionCombine: need exactly one signature",
			"len_signatures", len(request.Signatures),
		)
		return nil, ErrMalformedValue
	}
	sig := request.Signatures[0]
	var pk signature.PublicKey
	if err := pk.UnmarshalBinary(sig.PublicKey.Bytes); err != nil {
		loggerCons.Error("ConstructionCombine: malformed signature public key",
			"public_key_hex_bytes", hex.EncodeToString(sig.PublicKey.Bytes),
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	var rs signature.RawSignature
	if err := rs.UnmarshalBinary(sig.Bytes); err != nil {
		loggerCons.Error("ConstructionCombine: malformed signature",
			"signature_hex_bytes", hex.EncodeToString(sig.Bytes),
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	tx := transaction.SignedTransaction{
		Signed: signature.Signed{
			Blob: ut.Tx,
			Signature: signature.Signature{
				PublicKey: pk,
				Signature: rs,
			},
		},
	}

	resp := &types.ConstructionCombineResponse{
		SignedTransaction: base64.StdEncoding.EncodeToString(cbor.Marshal(tx)),
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionCombine OK", "response", jr)

	return resp, nil
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *constructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionParse: network validation failed", "err", terr.Message)
		return nil, terr
	}

	// Parse is called on both unsigned and signed transactions to understand
	// the intent of the formulated transaction. This is run as a sanity check
	// before signing (after `/construction/payloads`) and before broadcast
	// (after `/construction/combine`).

	rawTx, err := base64.StdEncoding.DecodeString(request.Transaction)
	if err != nil {
		loggerCons.Error("ConstructionParse: base64 decoding failed",
			"err", err,
		)
		return nil, ErrMalformedValue
	}

	var tx transaction.Transaction
	var from string
	var signers []string
	if request.Signed {
		var st transaction.SignedTransaction
		if err = cbor.Unmarshal(rawTx, &st); err != nil {
			loggerCons.Error("ConstructionParse: signed transaction unmarshal",
				"src", request.Transaction,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if err = st.Open(&tx); err != nil {
			loggerCons.Error("ConstructionParse: signed transaction open",
				"signed_transaction", st,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		from = StringFromAddress(staking.NewAddress(st.Signature.PublicKey))
		signers = []string{from}
	} else {
		var ut UnsignedTransaction
		if err = cbor.Unmarshal(rawTx, &ut); err != nil {
			loggerCons.Error("ConstructionParse: unsigned transaction unmarshal",
				"src", request.Transaction,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if err = cbor.Unmarshal(ut.Tx, &tx); err != nil {
			loggerCons.Error("ConstructionParse: inner unsigned transaction unmarshal",
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		from = ut.Signer
	}

	om := newTransactionToOperationMapper(&tx, from, "", []*types.Operation{})
	om.EmitFeeOps()
	if err := om.EmitTxOps(); err != nil {
		loggerCons.Error("ConstructionParse: malformed transaction",
			"err", err,
		)
		return nil, ErrMalformedValue
	}

	resp := &types.ConstructionParseResponse{
		Operations: om.Operations(),
		Signers:    signers,
		Metadata: map[string]interface{}{
			NonceKey: tx.Nonce,
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionParse OK", "response", jr)

	return resp, nil
}

// ConstructionPreprocess implements the /construction/preprocess endpoint.
func (s *constructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionPreprocess: network validation failed", "err", terr.Message)
		return nil, terr
	}

	// Preprocess is called prior to `/construction/payloads` to construct a
	// request for any metadata that is needed for transaction construction
	// given (i.e. account nonce). The request returned from this method will
	// be used by the caller (in a different execution environment) to call
	// the `/construction/metadata` endpoint.

	om := newOperationToTransactionMapper(request.Operations)
	signWithAddr, _, err := om.GetTransaction()
	if err != nil {
		loggerCons.Error("ConstructionPreprocess: bad operations",
			"err", err,
		)
		return nil, NewDetailedError(ErrMalformedValue, err)
	}

	resp := &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			OptionsIDKey: signWithAddr,
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionPreprocess OK", "response", jr)

	return resp, nil
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *constructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionPayloads: network validation failed", "err", terr.Message)
		return nil, terr
	}

	// Payloads is called with an array of operations and the response from
	// `/construction/metadata`. It returns an unsigned transaction blob and
	// a collection of payloads that must be signed by particular addresses
	// using a certain SignatureType. The array of operations provided in
	// transaction construction often times can not specify all "effects" of
	// a transaction (consider invoked transactions in Ethereum). However,
	// they can deterministically specify the "intent" of the transaction,
	// which is sufficient for construction. For this reason, parsing the
	// corresponding transaction in the Data API (when it lands on chain)
	// will contain a superset of whatever operations were provided during
	// construction.

	nonceRaw, ok := request.Metadata[NonceKey]
	if !ok {
		loggerCons.Error("ConstructionPayloads: nonce metadata not given")
		return nil, ErrMalformedValue
	}
	nonceF64, ok := nonceRaw.(float64)
	if !ok {
		loggerCons.Error("ConstructionPayloads: malformed nonce metadata")
		return nil, ErrMalformedValue
	}
	nonce := uint64(nonceF64)

	om := newOperationToTransactionMapper(request.Operations)
	signWithAddr, tx, err := om.GetTransaction()
	if err != nil {
		loggerCons.Error("ConstructionPayloads: bad operations",
			"err", err,
		)
		return nil, NewDetailedError(ErrMalformedValue, err)
	}

	tx.Nonce = nonce
	ut := UnsignedTransaction{
		Tx:     cbor.Marshal(tx),
		Signer: signWithAddr,
	}

	utCBOR := cbor.Marshal(ut)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: marshal unsigned transaction",
			"unsigned_transaction", ut,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	txMessage, err := signature.PrepareSignerMessage(transaction.SignatureContext, ut.Tx)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: PrepareSignerMessage",
			"signature_context", transaction.SignatureContext,
			"tx_hex", hex.EncodeToString(ut.Tx),
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	resp := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: base64.StdEncoding.EncodeToString(utCBOR),
		Payloads: []*types.SigningPayload{
			{
				Address:       signWithAddr,
				Bytes:         txMessage,
				SignatureType: types.Ed25519,
			},
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionPayloads OK", "response", jr)

	return resp, nil
}

// DecodeSignedTransaction decodes a signed transaction from a Base64-encoded CBOR blob.
func DecodeSignedTransaction(raw string) (*transaction.SignedTransaction, error) {
	rawTx, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var tx transaction.SignedTransaction
	if err := cbor.Unmarshal(rawTx, &tx); err != nil {
		return nil, fmt.Errorf("CBOR decode failed: %w", err)
	}
	return &tx, nil
}

// DecodeUnsignedTransaction decodes an unsigned transaction from a Base64-encoded CBOR blob.
func DecodeUnsignedTransaction(raw string) (*UnsignedTransaction, error) {
	rawTx, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var tx UnsignedTransaction
	if err := cbor.Unmarshal(rawTx, &tx); err != nil {
		return nil, fmt.Errorf("CBOR decode failed: %w", err)
	}
	return &tx, nil
}
