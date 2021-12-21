package main

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/coinbase/rosetta-cli/configuration"
	"github.com/coinbase/rosetta-sdk-go/storage/modules"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/tests/common"
)

func getRosettaConfig(ni *types.NetworkIdentifier) *configuration.Configuration {
	// Create a configuration file for the local testnet.
	config := configuration.DefaultConfiguration()

	config.Network = ni

	config.DataDirectory = "/tmp/rosetta-cli-oasistests"

	testEntityAddress, testEntityKeyPair := common.TestEntity()
	config.Construction = &configuration.ConstructionConfiguration{
		PrefundedAccounts: []*modules.PrefundedAccount{
			{
				PrivateKeyHex: hex.EncodeToString(testEntityKeyPair.PrivateKey),
				AccountIdentifier: &types.AccountIdentifier{
					Address:    testEntityAddress,
					SubAccount: nil,
					Metadata:   nil,
				},
				CurveType: types.Edwards25519,
				Currency:  services.OasisCurrency,
			},
		},
	}

	config.Construction.Scenario = []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: "{{ SENDER }}",
			},
			Amount: &types.Amount{
				Value:    "-100",
				Currency: services.OasisCurrency,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: services.StringFromAddress(api.FeeAccumulatorAddress),
			},
			Amount: &types.Amount{
				Value:    "100",
				Currency: services.OasisCurrency,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 2,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: "{{ SENDER }}",
			},
			Amount: &types.Amount{
				Value:    "{{ SENDER_VALUE }}",
				Currency: services.OasisCurrency,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 3,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: "{{ RECIPIENT }}",
			},
			Amount: &types.Amount{
				Value:    "{{ RECIPIENT_VALUE }}",
				Currency: services.OasisCurrency,
			},
		},
	}

	return config
}

func main() {
	var err error

	_, ni := common.NewRosettaClient()

	config := getRosettaConfig(ni)
	if err = ioutil.WriteFile("rosetta-cli-config.json", []byte(common.DumpJSON(config)), 0o600); err != nil {
		panic(err)
	}
}
