// https://djr6hkgq2tjcs.cloudfront.net/docs/construction_api_introduction.html
package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// OptionsIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the account ID.
const OptionsIDKey = "id"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "nonce"

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
) (*types.ConstructionSubmitResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionSubmit: network validation failed", "err", terr.Message)
		return nil, terr
	}

	if err := s.oasisClient.SubmitTx(ctx, request.SignedTransaction); err != nil {
		loggerCons.Error("ConstructionSubmit: SubmitTx failed", "err", err)
		return nil, ErrUnableToSubmitTx
	}

	var h hash.Hash
	var st transaction.SignedTransaction
	if err := json.Unmarshal([]byte(request.SignedTransaction), &st); err != nil {
		panic(err)
	}
	h.From(st)
	txID := h.String()

	resp := &types.ConstructionSubmitResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID,
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
) (*types.ConstructionHashResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionHash: network validation failed", "err", terr.Message)
		return nil, terr
	}

	var h hash.Hash
	var st transaction.SignedTransaction
	if err := json.Unmarshal([]byte(request.SignedTransaction), &st); err != nil {
		panic(err)
	}
	h.From(st)
	txID := h.String()

	resp := &types.ConstructionHashResponse{
		TransactionHash: txID,
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

	// According to the Rosetta spec:
	//     Blockchains that require an on-chain action to create an account
	//     should not implement this method.
	return nil, ErrNotImplemented
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

	/*resp := &types.ConstructionCombineResponse{
		SignedTransaction: // TODO
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionCombine OK", "response", jr)

	return resp, nil*/

	return nil, ErrNotImplemented
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

	// TODO: Write helpers that generate types.Operations from a staking
	// transaction (also, `appendOp` from `services/block.go` should be made
	// public and the code probably needs a refactor).

	/*resp := &types.ConstructionParseResponse{
		Operations: // TODO
		Signers:    // TODO
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionParse OK", "response", jr)

	return resp, nil*/

	return nil, ErrNotImplemented
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

	// TODO: Parse request.Operations and create the Options map for the
	// ConstructionMetadataRequest with OptionsIDKey in the map set to the
	// address of the account that's making the transaction.

	/*resp := &types.ConstructionPreprocessResponse{
		Options: // TODO
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionPreprocess OK", "response", jr)

	return resp, nil*/

	return nil, ErrNotImplemented
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

	/*resp := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: // TODO
		Payloads:            // TODO
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionPayloads OK", "response", jr)

	return resp, nil*/

	return nil, ErrNotImplemented
}
