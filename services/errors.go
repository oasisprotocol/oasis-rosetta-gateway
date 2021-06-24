package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/errors"
)

const (
	// CauseKey is a key in an error's details map, mapping to an object with
	// keys ModuleKey, CodeKey (see
	// https://pkg.go.dev/github.com/oasisprotocol/oasis-core/go@v0.20.9/common/errors?tab=doc#Code),
	// and MsgKey with some information about an error
	// from the Oasis network.
	CauseKey  = "cause"
	ModuleKey = "module"
	CodeKey   = "code"
	MsgKey    = "msg"
)

var (
	ErrUnableToGetChainID = &types.Error{
		Code:      1,
		Message:   "unable to get chain ID",
		Retriable: true,
	}

	ErrInvalidBlockchain = &types.Error{
		Code:      2,
		Message:   "invalid blockchain specified in network identifier",
		Retriable: false,
	}

	ErrInvalidSubnetwork = &types.Error{
		Code:      3,
		Message:   "invalid sub-network identifier",
		Retriable: false,
	}

	ErrInvalidNetwork = &types.Error{
		Code:      4,
		Message:   "invalid network specified in network identifier",
		Retriable: false,
	}

	ErrMissingNID = &types.Error{
		Code:      5,
		Message:   "network identifier is missing",
		Retriable: false,
	}

	ErrUnableToGetLatestBlk = &types.Error{
		Code:      6,
		Message:   "unable to get latest block",
		Retriable: true,
	}

	ErrUnableToGetGenesisBlk = &types.Error{
		Code:      7,
		Message:   "unable to get genesis block",
		Retriable: true,
	}

	ErrUnableToGetAccount = &types.Error{
		Code:      8,
		Message:   "unable to get account",
		Retriable: true,
	}

	ErrMustQueryByIndex = &types.Error{
		Code:      9,
		Message:   "blocks must be queried by index and not hash",
		Retriable: false,
	}

	ErrInvalidAccountAddress = &types.Error{
		Code:      10,
		Message:   "invalid account address",
		Retriable: false,
	}

	ErrMustSpecifySubAccount = &types.Error{
		Code:      11,
		Message:   "a valid subaccount must be specified (absent or {\"address\": \"escrow\"})",
		Retriable: false,
	}

	ErrUnableToGetBlk = &types.Error{
		Code:      12,
		Message:   "unable to get block",
		Retriable: true,
	}

	ErrNotImplemented = &types.Error{
		Code:      13,
		Message:   "operation not implemented",
		Retriable: false,
	}

	ErrUnableToGetTxns = &types.Error{
		Code:      14,
		Message:   "unable to get transactions",
		Retriable: true,
	}

	ErrUnableToSubmitTx = &types.Error{
		Code:      15,
		Message:   "unable to submit transaction",
		Retriable: false,
	}

	ErrUnableToGetNextNonce = &types.Error{
		Code:      16,
		Message:   "unable to get next nonce",
		Retriable: true,
	}

	ErrMalformedValue = &types.Error{
		Code:      17,
		Message:   "malformed value",
		Retriable: false,
	}

	ErrUnableToGetNodeStatus = &types.Error{
		Code:      18,
		Message:   "unable to get node status",
		Retriable: true,
	}

	ErrTransactionNotFound = &types.Error{
		Code:      19,
		Message:   "transaction not found",
		Retriable: true,
	}

	ErrNotAvailableInOfflineMode = &types.Error{
		Code:      20,
		Message:   "not available in offline mode",
		Retriable: false,
	}

	ErrorList = []*types.Error{
		ErrUnableToGetChainID,
		ErrInvalidBlockchain,
		ErrInvalidSubnetwork,
		ErrInvalidNetwork,
		ErrMissingNID,
		ErrUnableToGetLatestBlk,
		ErrUnableToGetGenesisBlk,
		ErrUnableToGetAccount,
		ErrMustQueryByIndex,
		ErrInvalidAccountAddress,
		ErrMustSpecifySubAccount,
		ErrUnableToGetBlk,
		ErrNotImplemented,
		ErrUnableToGetTxns,
		ErrUnableToSubmitTx,
		ErrUnableToGetNextNonce,
		ErrMalformedValue,
		ErrUnableToGetNodeStatus,
		ErrTransactionNotFound,
		ErrNotAvailableInOfflineMode,
	}
)

// NewDetailedError returns a new Rosetta error Code, Message, and Retriable
// set from proto and Details[CauseKey] set from cause.
func NewDetailedError(proto *types.Error, cause error) *types.Error {
	module, code, msg := errors.Code(cause)
	detailedError := *proto
	detailedError.Details = map[string]interface{}{
		CauseKey: map[string]interface{}{
			ModuleKey: module,
			CodeKey:   code,
			MsgKey:    msg,
		},
	}
	return &detailedError
}
