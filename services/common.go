package services

import (
	"context"
	"os"

	"github.com/coinbase/rosetta-sdk-go/types"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
)

// OasisBlockchainName is the name of the Oasis blockchain.
const OasisBlockchainName = "Oasis"

// OasisCurrency is the currency used on the Oasis blockchain.
var OasisCurrency = &types.Currency{
	Symbol:   "ROSE",
	Decimals: 9,
}

// OfflineModeChainIDEnvVar is the name of the environment variable that
// specifies the chain ID when running in offline mode.  This is required to
// be able to properly sign transactions, since we can't get the chain ID from
// the node.
const OfflineModeChainIDEnvVar = "OASIS_ROSETTA_GATEWAY_OFFLINE_MODE_CHAIN_ID"

// GetChainID returns the chain ID.
func GetChainID(ctx context.Context, oc oc.OasisClient) (string, *types.Error) {
	chainID, err := oc.GetChainID(ctx)
	if err != nil {
		return "", ErrUnableToGetChainID
	}
	return chainID, nil
}

// ValidateNetworkIdentifier validates the network identifier and fetches the
// chain ID either from the given OasisClient (if not nil), or from the env
// variable (if oc is nil).
func ValidateNetworkIdentifier(ctx context.Context, oc oc.OasisClient, ni *types.NetworkIdentifier) *types.Error {
	var chainID string

	if oc != nil {
		// Obtain chain ID from node.
		var err *types.Error
		chainID, err = GetChainID(ctx, oc)
		if err != nil {
			return err
		}
	} else {
		// Obtain chain ID from env var.
		// Note that we already check that this env var is not empty in main.go.
		chainID = os.Getenv(OfflineModeChainIDEnvVar)
	}
	return ValidateNetworkIdentifierWithChainID(chainID, ni)
}

// ValidateNetworkIdentifierWithChainID validates the network identifier and
// uses the given chain ID.
func ValidateNetworkIdentifierWithChainID(chainID string, ni *types.NetworkIdentifier) *types.Error {
	if ni != nil {
		if ni.Blockchain != OasisBlockchainName {
			return ErrInvalidBlockchain
		}
		if ni.SubNetworkIdentifier != nil {
			return ErrInvalidSubnetwork
		}
		if ni.Network != chainID {
			return ErrInvalidNetwork
		}
	} else {
		return ErrMissingNID
	}
	return nil
}

// StringFromAddress converts a staking API address to string using MarshalText.
// If marshalling fails, this panics.
func StringFromAddress(address staking.Address) string {
	buf, err := address.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(buf)
}
