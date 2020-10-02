package services

import (
	"context"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/oasisprotocol/oasis-core/go/common/logging"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
)

// SubAccountEscrow specifies the name of the escrow subaccount.
const SubAccountEscrow = "escrow"

// ActiveBalanceKey is the name of the key in the Metadata map inside
// the response of an account balance request for an escrow account.
// The value in the Metadata map specifies how many token base units are in
// the active escrow pool.
const ActiveBalanceKey = "active_balance"

// ActiveSharesKey is the name of the key in the Metadata map inside
// the response of an account balance request for an escrow account.
// The value in the Metadata map specifies how many shares are in
// the active escrow pool.
const ActiveSharesKey = "active_shares"

// DebondingBalanceKey is the name of the key in the Metadata map inside
// the response of an account balance request for an escrow account.
// The value in the Metadata map specifies how many token base units are in
// the debonding escrow pool.
const DebondingBalanceKey = "debonding_balance"

// DebondingSharesKey is the name of the key in the Metadata map inside
// the response of an account balance request for an escrow account.
// The value in the Metadata map specifies how many shares are in
// the debonding escrow pool.
const DebondingSharesKey = "debonding_shares"

var loggerAcct = logging.GetLogger("services/account")

type accountAPIService struct {
	oasisClient oc.OasisClient
}

// NewAccountAPIService creates a new instance of an AccountAPIService.
func NewAccountAPIService(oasisClient oc.OasisClient) server.AccountAPIServicer {
	return &accountAPIService{
		oasisClient: oasisClient,
	}
}

// AccountBalance implements the /account/balance endpoint.
func (s *accountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerAcct.Error("AccountBalance: network validation failed", "err", terr.Message)
		return nil, terr
	}

	height := oc.LatestHeight

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index != nil {
			height = *request.BlockIdentifier.Index
		} else if request.BlockIdentifier.Hash != nil {
			loggerAcct.Error("AccountBalance: must query block by index")
			return nil, ErrMustQueryByIndex
		}
	}

	if request.AccountIdentifier.Address == "" {
		loggerAcct.Error("AccountBalance: invalid account address (empty)")
		return nil, ErrInvalidAccountAddress
	}

	var owner staking.Address
	err := owner.UnmarshalText([]byte(request.AccountIdentifier.Address))
	if err != nil {
		loggerAcct.Error("AccountBalance: invalid account address", "err", err)
		return nil, ErrInvalidAccountAddress
	}

	switch {
	case request.AccountIdentifier.SubAccount == nil:
	case request.AccountIdentifier.SubAccount.Address == SubAccountEscrow:
	default:
		loggerAcct.Error("AccountBalance: invalid subaccount", "sub_account", request.AccountIdentifier.SubAccount)
		return nil, ErrMustSpecifySubAccount
	}

	act, err := s.oasisClient.GetAccount(ctx, height, owner)
	if err != nil {
		loggerAcct.Error("AccountBalance: unable to get account",
			"account_id", owner.String(),
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetAccount
	}

	blk, err := s.oasisClient.GetBlock(ctx, height)
	if err != nil {
		loggerAcct.Error("AccountBalance: unable to get block",
			"height", height,
			"err", err,
		)
		return nil, ErrUnableToGetBlk
	}

	md := make(map[string]interface{})
	md[NonceKey] = act.General.Nonce

	var value string
	switch {
	case request.AccountIdentifier.SubAccount == nil:
		value = act.General.Balance.String()
	case request.AccountIdentifier.SubAccount.Address == SubAccountEscrow:
		// Total is Active + Debonding.
		total := act.Escrow.Active.Balance.Clone()
		if err := total.Add(&act.Escrow.Debonding.Balance); err != nil {
			loggerAcct.Error("AccountBalance: escrow: unable to add debonding to active",
				"account_id", owner.String(),
				"height", height,
				"escrow_active_balance", act.Escrow.Active.Balance.String(),
				"escrow_debonding_balance", act.Escrow.Debonding.Balance.String(),
				"err", err,
			)
			return nil, ErrMalformedValue
		}
		value = total.String()

		md[ActiveBalanceKey] = act.Escrow.Active.Balance.String()
		md[ActiveSharesKey] = act.Escrow.Active.TotalShares.String()
		md[DebondingBalanceKey] = act.Escrow.Debonding.Balance.String()
		md[DebondingSharesKey] = act.Escrow.Debonding.TotalShares.String()
	default:
		// This shouldn't happen, since we already check for this above.
		return nil, ErrMustSpecifySubAccount
	}

	resp := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: blk.Height,
			Hash:  blk.Hash,
		},
		Balances: []*types.Amount{
			&types.Amount{
				Value:    value,
				Currency: OasisCurrency,
			},
		},
		Metadata: md,
	}

	jr, _ := json.Marshal(resp)
	loggerAcct.Debug("AccountBalance OK",
		"response", jr,
		"account_id", owner.String(),
		"sub_account", request.AccountIdentifier.SubAccount,
	)

	return resp, nil
}
