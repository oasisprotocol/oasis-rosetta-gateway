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
	"github.com/vmihailenco/msgpack/v5"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/tests/common"
)

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
	config.Construction.CurveType = types.Edwards25519

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
		key := struct {
			Address string        `json:"address"`
			KeyPair *keys.KeyPair `json:"keypair"`
		}{testEntityAddress, testEntityKeyPair}
		var keyBuf bytes.Buffer
		enc := msgpack.GetEncoder()
		enc.Reset(&keyBuf)
		enc.UseJSONTag(true)
		if err := enc.Encode(&key); err != nil {
			panic(err)
		}
		msgpack.PutEncoder(enc)
		if err := txn.Set([]byte("key/"+testEntityAddress), keyBuf.Bytes()); err != nil {
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
