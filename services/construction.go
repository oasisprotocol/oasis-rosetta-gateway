// https://djr6hkgq2tjcs.cloudfront.net/docs/construction_api_introduction.html
package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// OptionsIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the account ID.
const OptionsIDKey = "id"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "nonce"

// LiteralKey is the name of the key in the Metadata map inside a
// OpDummyLiteral operation that specifies the hex-encoded CBOR-encoded
// Oasis consensus transaction.
const LiteralKey = "literal"

// DefaultGas is the gas limit used in creating a transaction.
const DefaultGas = 10000

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

	txBuf, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		panic(err)
	}
	if len(request.Signatures) != 1 {
		panic("len(request.Signatures)")
	}
	sig := request.Signatures[0]
	var pk signature.PublicKey
	if err := pk.UnmarshalBinary(sig.PublicKey.Bytes); err != nil {
		panic(err)
	}
	var rs signature.RawSignature
	if err := rs.UnmarshalBinary(sig.Bytes); err != nil {
		panic(err)
	}
	st := transaction.SignedTransaction{
		Signed: signature.Signed{
			Blob: txBuf,
			Signature: signature.Signature{
				PublicKey: pk,
				Signature: rs,
			},
		},
	}
	stJSON, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}

	resp := &types.ConstructionCombineResponse{
		SignedTransaction: string(stJSON),
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

	if len(request.Operations) < 1 {
		loggerCons.Error("ConstructionPreprocess: missing fee operation")
		return nil, ErrMalformedValue
	}
	feeOp := request.Operations[0]

	resp := &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			OptionsIDKey: feeOp.Account.Address,
		},
	}

	jr, _ := json.Marshal(resp)
	loggerCons.Debug("ConstructionPreprocess OK", "response", jr)

	return resp, nil
}

func readCurrency(amount *types.Amount, currency *types.Currency, negative bool) (*quantity.Quantity, error) {
	// TODO: Is it up to us to check other fields?
	if amount.Currency.Symbol != currency.Symbol {
		return nil, fmt.Errorf("wrong currency")
	}
	bi := new(big.Int)
	if err := bi.UnmarshalText([]byte(amount.Value)); err != nil {
		return nil, fmt.Errorf("bi UnmarshalText Value: %w", err)
	}
	if negative {
		bi.Neg(bi)
	}
	q := quantity.NewQuantity()
	if err := q.FromBigInt(bi); err != nil {
		return nil, fmt.Errorf("q FromBigInt bi: %w", err)
	}
	return q, nil
}

func readOasisCurrency(amount *types.Amount) (*quantity.Quantity, error) {
	return readCurrency(amount, OasisCurrency, false)
}

func readOasisCurrencyNeg(amount *types.Amount) (*quantity.Quantity, error) {
	return readCurrency(amount, OasisCurrency, true)
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

	if len(request.Operations) < 2 {
		loggerCons.Error("ConstructionPayloads: missing fee operation")
		return nil, ErrMalformedValue
	}
	feeOp := request.Operations[0]
	if feeOp.Type != OpTransfer {
		loggerCons.Error("ConstructionPayloads: fee operation wrong type",
			"type", feeOp.Type,
			"expected_type", OpTransfer,
		)
		return nil, ErrMalformedValue
	}
	if feeOp.Account.SubAccount == nil {
		loggerCons.Error("ConstructionPayloads: missing fee operation subaccount")
		return nil, ErrMalformedValue
	}
	if feeOp.Account.SubAccount.Address != SubAccountGeneral {
		loggerCons.Error("ConstructionPayloads: fee operation wrong subaccount address",
			"subaccount", feeOp.Account.SubAccount.Address,
			"expected_subaccount", SubAccountGeneral,
		)
		return nil, ErrMalformedValue
	}
	signWithAddr := feeOp.Account.Address
	feeAmount, err := readOasisCurrencyNeg(feeOp.Amount)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: readOasisCurrency feeOp.Amount",
			"amount", feeOp.Amount,
			"err", err,
		)
		return nil, ErrMalformedValue
	}

	var method transaction.MethodName
	var body cbor.RawMessage
	if len(request.Operations) == 3 &&
		request.Operations[1].Type == OpTransfer &&
		request.Operations[1].Account.SubAccount != nil &&
		request.Operations[1].Account.SubAccount.Address == SubAccountGeneral &&
		request.Operations[2].Type == OpTransfer &&
		request.Operations[2].Account.SubAccount != nil &&
		request.Operations[2].Account.SubAccount.Address == SubAccountGeneral {
		loggerCons.Debug("ConstructionPayloads: matched transfer")
		method = staking.MethodTransfer

		if request.Operations[1].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: transfer from doesn't match signer",
				"from", request.Operations[1].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(request.Operations[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: transfer from amount",
				"amount", request.Operations[1].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		var to staking.Address
		if err = to.UnmarshalText([]byte(request.Operations[2].Account.Address)); err != nil {
			loggerCons.Error("ConstructionPayloads: transfer to UnmarshalText",
				"addr", request.Operations[2].Account.Address,
				"err", err,
			)
		}
		amount2, err := readOasisCurrency(request.Operations[2].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: transfer to amount",
				"amount", request.Operations[2].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if amount.Cmp(amount2) != 0 {
			loggerCons.Error("ConstructionPayloads: transfer amounts differ",
				"amount_from", amount,
				"amount_to", amount2,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		body = cbor.Marshal(staking.Transfer{
			To:     to,
			Tokens: *amount,
		})
	} else if len(request.Operations) == 2 &&
		request.Operations[1].Type == OpBurn &&
		request.Operations[1].Account.SubAccount != nil &&
		request.Operations[1].Account.SubAccount.Address == SubAccountGeneral {
		loggerCons.Debug("ConstructionPayloads: matched burn")
		method = staking.MethodBurn

		if request.Operations[1].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: burn from doesn't match signer",
				"from", request.Operations[1].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(request.Operations[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: burn from amount",
				"amount", request.Operations[1].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		body = cbor.Marshal(staking.Burn{
			Tokens: *amount,
		})
	} else if len(request.Operations) == 3 &&
		request.Operations[1].Type == OpTransfer &&
		request.Operations[1].Account.SubAccount != nil &&
		request.Operations[1].Account.SubAccount.Address == SubAccountGeneral &&
		request.Operations[2].Type == OpTransfer &&
		request.Operations[2].Account.SubAccount != nil &&
		request.Operations[2].Account.SubAccount.Address == SubAccountEscrow {
		loggerCons.Debug("ConstructionPayloads: matched add escrow")
		method = staking.MethodAddEscrow

		if request.Operations[1].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: add escrow from doesn't match signer",
				"from", request.Operations[1].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(request.Operations[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow from amount",
				"amount", request.Operations[1].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		var escrowAccount staking.Address
		if err = escrowAccount.UnmarshalText([]byte(request.Operations[2].Account.Address)); err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow account UnmarshalText",
				"addr", request.Operations[2].Account.Address,
				"err", err,
			)
		}
		amount2, err := readOasisCurrency(request.Operations[2].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow account amount",
				"amount", request.Operations[2].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if amount.Cmp(amount2) != 0 {
			loggerCons.Error("ConstructionPayloads: add escrow amounts differ",
				"amount_from", amount,
				"amount_to", amount2,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		body = cbor.Marshal(staking.Escrow{
			Account: escrowAccount,
			Tokens:  *amount,
		})
	} else {
		loggerCons.Error("ConstructionPayloads: unmatched operations list",
			"operations", request.Operations,
		)
		return nil, ErrNotImplemented
	}

	tx := transaction.Transaction{
		Nonce: nonce,
		Fee: &transaction.Fee{
			Amount: *feeAmount,
			Gas:    DefaultGas,
		},
		Method: method,
		Body:   body,
	}

	txCBOR := cbor.Marshal(tx)
	txHex := hex.EncodeToString(txCBOR)
	txMessage, err := signature.PrepareSignerMessage(transaction.SignatureContext, txCBOR)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: PrepareSignerMessage",
			"signature_context", transaction.SignatureContext,
			"tx_hex", txHex,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	resp := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: txHex,
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
