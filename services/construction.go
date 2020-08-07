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
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
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

// FeeGasKey is the name of the key in the Metadata map inside a fee
// operation that specifies the gas value in the transaction fee.
// This is optional, and we use DefaultGas if it's absent.
const FeeGasKey = "fee_gas"

// DefaultGas is the gas limit used in creating a transaction.
const DefaultGas transaction.Gas = 10000

// UnsignedTransaction is a transaction with the account that would sign it.
type UnsignedTransaction struct {
	Tx     transaction.Transaction `json:"tx"`
	Signer string                  `json:"signer"`
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
		loggerCons.Error("ConstructionSubmit: unmarshal unsigned transaction",
			"unsigned_transaction", request.SignedTransaction,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	h.From(st)
	txID := h.String()

	resp := &types.TransactionIdentifierResponse{
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
) (*types.TransactionIdentifierResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerCons.Error("ConstructionHash: network validation failed", "err", terr.Message)
		return nil, terr
	}

	var h hash.Hash
	var st transaction.SignedTransaction
	if err := json.Unmarshal([]byte(request.SignedTransaction), &st); err != nil {
		loggerCons.Error("ConstructionHash: unmarshal unsigned transaction",
			"unsigned_transaction", request.SignedTransaction,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	h.From(st)
	txID := h.String()

	resp := &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID,
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

	var ut UnsignedTransaction
	if err := json.Unmarshal([]byte(request.UnsignedTransaction), &ut); err != nil {
		loggerCons.Error("ConstructionCombine: unmarshal unsigned transaction",
			"unsigned_transaction", request.UnsignedTransaction,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	txBuf := cbor.Marshal(ut.Tx)
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
		loggerCons.Error("ConstructionCombine: marshal signed transaction",
			"signed_transaction", st,
			"err", err,
		)
		return nil, ErrMalformedValue
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

	// TODO: Unify some of this verboseness with block.go. If you prefer.

	var tx transaction.Transaction
	var from string
	var signers []string
	if request.Signed {
		var st transaction.SignedTransaction
		if err := json.Unmarshal([]byte(request.Transaction), &st); err != nil {
			loggerCons.Error("ConstructionParse: signed transaction unmarshal",
				"src", request.Transaction,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if err := st.Open(&tx); err != nil {
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
		if err := json.Unmarshal([]byte(request.Transaction), &ut); err != nil {
			loggerCons.Error("ConstructionParse: unsigned transaction unmarshal",
				"src", request.Transaction,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		tx = ut.Tx
		from = ut.Signer
	}

	feeAmountStr := "-0"
	feeGas := transaction.Gas(0)
	if tx.Fee != nil {
		feeAmountStr = "-" + tx.Fee.Amount.String()
		feeGas = tx.Fee.Gas
	}
	ops := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type: OpTransfer,
			Account: &types.AccountIdentifier{
				Address: from,
			},
			Amount: &types.Amount{
				Value:    feeAmountStr,
				Currency: OasisCurrency,
			},
			Metadata: map[string]interface{}{
				FeeGasKey: feeGas,
			},
		},
	}
	switch tx.Method {
	case staking.MethodTransfer:
		var body staking.Transfer
		if err := cbor.Unmarshal(tx.Body, &body); err != nil {
			loggerCons.Error("ConstructionParse: transfer unmarshal",
				"body", tx.Body,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		ops = append(ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 1,
				},
				Type: OpTransfer,
				Account: &types.AccountIdentifier{
					Address: from,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Tokens.String(),
					Currency: OasisCurrency,
				},
			},
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 2,
				},
				Type: OpTransfer,
				Account: &types.AccountIdentifier{
					Address: StringFromAddress(body.To),
				},
				Amount: &types.Amount{
					Value:    body.Tokens.String(),
					Currency: OasisCurrency,
				},
			},
		)
	case staking.MethodBurn:
		var body staking.Burn
		if err := cbor.Unmarshal(tx.Body, &body); err != nil {
			loggerCons.Error("ConstructionParse: burn unmarshal",
				"body", tx.Body,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		ops = append(ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 1,
				},
				Type: OpBurn,
				Account: &types.AccountIdentifier{
					Address: from,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Tokens.String(),
					Currency: OasisCurrency,
				},
			},
		)
	case staking.MethodAddEscrow:
		var body staking.Escrow
		if err := cbor.Unmarshal(tx.Body, &body); err != nil {
			loggerCons.Error("ConstructionParse: add escrow unmarshal",
				"body", tx.Body,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		ops = append(ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 1,
				},
				Type: OpTransfer,
				Account: &types.AccountIdentifier{
					Address: from,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Tokens.String(),
					Currency: OasisCurrency,
				},
			},
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 2,
				},
				Type: OpTransfer,
				Account: &types.AccountIdentifier{
					Address: StringFromAddress(body.Account),
					SubAccount: &types.SubAccountIdentifier{
						Address: SubAccountEscrow,
					},
				},
				Amount: &types.Amount{
					Value:    body.Tokens.String(),
					Currency: OasisCurrency,
				},
			},
		)
	default:
		loggerCons.Error("ConstructionParse: unmatched method",
			"method", tx.Method,
		)
		return nil, ErrNotImplemented
	}

	resp := &types.ConstructionParseResponse{
		Operations: ops,
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

	// Coincidentally the first operation comes from the sender, whether it's
	// a fee transfer or any of the transaction types currently supported.
	if len(request.Operations) < 1 {
		loggerCons.Error("ConstructionPreprocess: missing operations")
		return nil, ErrMalformedValue
	}
	senderOp := request.Operations[0]

	resp := &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			OptionsIDKey: senderOp.Account.Address,
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

	remainingOps := request.Operations
	var signWithAddr string
	feeGas := DefaultGas
	feeAmount := quantity.NewQuantity()
	var err error

	if len(remainingOps) >= 2 &&
		remainingOps[0].Type == OpTransfer &&
		remainingOps[1].Type == OpTransfer &&
		remainingOps[1].Account.Address == StringFromAddress(staking.FeeAccumulatorAddress) {
		loggerCons.Debug("ConstructionPayloads: matched fee transfer")

		signWithAddr = remainingOps[0].Account.Address
		if remainingOps[0].Account.SubAccount != nil {
			loggerCons.Error("ConstructionPayloads: fee transfer from wrong subaccount",
				"sub_account", remainingOps[0].Account.SubAccount,
				"expected_sub_account", nil,
			)
			return nil, ErrMalformedValue
		}
		feeAmount, err = readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: fee transfer from amount",
				"amount", remainingOps[0].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if feeGasRaw, ok := remainingOps[0].Metadata[FeeGasKey]; ok {
			feeGasF64, ok := feeGasRaw.(float64)
			if !ok {
				loggerCons.Error("ConstructionPayloads: malformed fee transfer gas metadata")
				return nil, ErrMalformedValue
			}
			feeGas = transaction.Gas(feeGasF64)
		}

		if remainingOps[1].Account.SubAccount != nil {
			loggerCons.Error("ConstructionPayloads: fee transfer to wrong subaccount",
				"sub_account", remainingOps[1].Account.SubAccount,
				"expected_sub_account", nil,
			)
			return nil, ErrMalformedValue
		}
		feeAmount2, err := readOasisCurrency(remainingOps[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: transfer to amount",
				"amount", remainingOps[1].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		if feeAmount.Cmp(feeAmount2) != 0 {
			loggerCons.Error("ConstructionPayloads: fee transfer amounts differ",
				"amount_from", feeAmount,
				"amount_to", feeAmount2,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		remainingOps = remainingOps[2:]
	}

	var method transaction.MethodName
	var body cbor.RawMessage
	switch {
	case len(remainingOps) == 2 &&
		remainingOps[0].Type == OpTransfer &&
		remainingOps[0].Account.SubAccount == nil &&
		remainingOps[1].Type == OpTransfer &&
		remainingOps[1].Account.SubAccount == nil:
		loggerCons.Debug("ConstructionPayloads: matched transfer")
		method = staking.MethodTransfer

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: transfer from doesn't match signer",
				"from", remainingOps[0].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: transfer from amount",
				"amount", remainingOps[0].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		var to staking.Address
		if err = to.UnmarshalText([]byte(remainingOps[1].Account.Address)); err != nil {
			loggerCons.Error("ConstructionPayloads: transfer to UnmarshalText",
				"addr", remainingOps[1].Account.Address,
				"err", err,
			)
		}
		amount2, err := readOasisCurrency(remainingOps[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: transfer to amount",
				"amount", remainingOps[1].Amount,
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
	case len(remainingOps) == 1 &&
		remainingOps[0].Type == OpBurn &&
		remainingOps[0].Account.SubAccount == nil:
		loggerCons.Debug("ConstructionPayloads: matched burn")
		method = staking.MethodBurn

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: burn from doesn't match signer",
				"from", remainingOps[0].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: burn from amount",
				"amount", remainingOps[0].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		body = cbor.Marshal(staking.Burn{
			Tokens: *amount,
		})
	case len(remainingOps) == 2 &&
		remainingOps[0].Type == OpTransfer &&
		remainingOps[0].Account.SubAccount == nil &&
		remainingOps[1].Type == OpTransfer &&
		remainingOps[1].Account.SubAccount != nil &&
		remainingOps[1].Account.SubAccount.Address == SubAccountEscrow:
		loggerCons.Debug("ConstructionPayloads: matched add escrow")
		method = staking.MethodAddEscrow

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			loggerCons.Error("ConstructionPayloads: add escrow from doesn't match signer",
				"from", remainingOps[0].Account.Address,
				"signer", signWithAddr,
			)
			return nil, ErrMalformedValue
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow from amount",
				"amount", remainingOps[0].Amount,
				"err", err,
			)
			return nil, ErrMalformedValue
		}

		var escrowAccount staking.Address
		if err = escrowAccount.UnmarshalText([]byte(remainingOps[1].Account.Address)); err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow account UnmarshalText",
				"addr", remainingOps[1].Account.Address,
				"err", err,
			)
		}
		amount2, err := readOasisCurrency(remainingOps[1].Amount)
		if err != nil {
			loggerCons.Error("ConstructionPayloads: add escrow account amount",
				"amount", remainingOps[1].Amount,
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
	// TODO: Devise a way to support reclaim escrow.
	default:
		loggerCons.Error("ConstructionPayloads: unmatched operations list",
			"operations", remainingOps,
		)
		return nil, ErrNotImplemented
	}

	ut := UnsignedTransaction{
		Tx: transaction.Transaction{
			Nonce: nonce,
			Fee: &transaction.Fee{
				Amount: *feeAmount,
				Gas:    feeGas,
			},
			Method: method,
			Body:   body,
		},
		Signer: signWithAddr,
	}

	utJSON, err := json.Marshal(ut)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: marshal unsigned transaction",
			"unsigned_transaction", ut,
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	txCBOR := cbor.Marshal(ut.Tx)
	txMessage, err := signature.PrepareSignerMessage(transaction.SignatureContext, txCBOR)
	if err != nil {
		loggerCons.Error("ConstructionPayloads: PrepareSignerMessage",
			"signature_context", transaction.SignatureContext,
			"tx_hex", hex.EncodeToString(txCBOR),
			"err", err,
		)
		return nil, ErrMalformedValue
	}
	resp := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(utJSON),
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
