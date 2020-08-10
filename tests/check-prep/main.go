package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/coinbase/rosetta-cli/configuration"
	"github.com/coinbase/rosetta-sdk-go/client"
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

func main() {
	// Create a configuration file for the local testnet.
	config := configuration.DefaultConfiguration()

	rc := client.NewAPIClient(client.NewConfiguration("http://localhost:8080", "rosetta-sdk-go", nil))
	nlr, re, err := rc.NetworkAPI.NetworkList(context.Background(), &types.MetadataRequest{})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	if len(nlr.NetworkIdentifiers) != 1 {
		panic("len(nlr.NetworkIdentifiers)")
	}
	fmt.Println("network identifiers", common.DumpJSON(nlr.NetworkIdentifiers))
	config.Network = nlr.NetworkIdentifiers[0]

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

	if err := ioutil.WriteFile("rosetta-cli-config.json", []byte(common.DumpJSON(config)), 0o666); err != nil {
		panic(err)
	}

	// Create an account for construction tests.
	testEntityAddress, testEntityKeyPair := common.TestEntity()

	constructionStorePath := path.Join(config.DataDirectory, "check-construction", types.Hash(config.Network))
	if err := os.MkdirAll(constructionStorePath, 0o777); err != nil {
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
		if err := txn.Set([]byte("balance/"+types.Hash(&testEntityAccountIdentifier)+"/"+types.Hash(services.OasisCurrency)), storageEncode(&struct {
			Account *types.AccountIdentifier `json:"account"`
			Amount  *types.Amount            `json:"amount"`
			Block   *types.BlockIdentifier   `json:"block"`
		}{
			Account: &testEntityAccountIdentifier,
			Amount: &types.Amount{
				// https://github.com/oasisprotocol/oasis-core/blob/v20.8.2/go/oasis-node/cmd/genesis/genesis.go#L534
				Value:    "100000000000",
				Currency: services.OasisCurrency,
			},
		})); err != nil {
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
