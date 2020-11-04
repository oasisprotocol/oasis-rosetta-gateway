package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"

	"github.com/coinbase/rosetta-cli/configuration"
	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/dgraph-io/badger"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/tests/common"
)

func storageEncode(v interface{}) []byte {
	var buf bytes.Buffer
	enc := msgpack.GetEncoder()
	enc.Reset(&buf)
	enc.UseJSONTag(true)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	msgpack.PutEncoder(enc)
	return buf.Bytes()
}

func getRosettaConfig(ni *types.NetworkIdentifier) *configuration.Configuration {
	// Create a configuration file for the local testnet.
	config := configuration.DefaultConfiguration()

	config.Network = ni

	config.DataDirectory = "/tmp/rosetta-cli-oasistests"

	config.Construction.Currency = services.OasisCurrency
	config.Construction.MaximumFee = "0"
	config.Construction.CurveType = types.Edwards25519
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

	// Create an account for construction tests.
	testEntityAddress, testEntityKeyPair := common.TestEntity()

	constructionStorePath := path.Join(config.DataDirectory, "check-construction", types.Hash(config.Network))
	if err = os.MkdirAll(constructionStorePath, 0o777); err != nil {
		panic(err)
	}
	db, err := badger.Open(badger.DefaultOptions(constructionStorePath))
	if err != nil {
		panic(err)
	}
	if err := db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte("key/"+testEntityAddress), storageEncode(&struct {
			Address string        `json:"address"`
			KeyPair *keys.KeyPair `json:"keypair"`
		}{
			Address: testEntityAddress,
			KeyPair: testEntityKeyPair,
		})); err != nil {
			panic(err)
		}
		testEntityAccountIdentifier := types.AccountIdentifier{Address: testEntityAddress}
		if err := txn.Set(
			[]byte("balance/"+types.Hash(&testEntityAccountIdentifier)+"/"+types.Hash(services.OasisCurrency)),
			storageEncode(
				&struct {
					Account *types.AccountIdentifier `json:"account"`
					Amount  *types.Amount            `json:"amount"`
					Block   *types.BlockIdentifier   `json:"block"`
				}{
					Account: &testEntityAccountIdentifier,
					Amount:  common.TestEntityAmount,
				},
			),
		); err != nil {
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}
	if err := db.Close(); err != nil {
		panic(err)
	}
}
