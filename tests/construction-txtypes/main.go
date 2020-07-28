package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
)

const dstAddress = "oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"
const dummyNonce = 3

func main() {
	_, signer, err := entity.TestEntity()
	if err != nil {
		panic(err)
	}
	testEntityAddress := api.NewAddress(signer.Public()).String()

	var dstAddr api.Address
	if err := dstAddr.UnmarshalText([]byte(dstAddress)); err != nil {
		panic(err)
	}
	fee100Op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: 0,
		},
		Type: services.OpTransfer,
		Account: &types.AccountIdentifier{
			Address: testEntityAddress,
			SubAccount: &types.SubAccountIdentifier{
				Address: services.SubAccountGeneral,
			},
		},
		Amount: &types.Amount{
			Value:    "-100",
			Currency: services.OasisCurrency,
		},
	}
	fee100 := &transaction.Fee{
		Amount: *quantity.NewFromUint64(100),
		Gas:    services.DefaultGas,
	}
	opsTransfer := []*types.Operation{
		fee100Op,
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: testEntityAddress,
				SubAccount: &types.SubAccountIdentifier{
					Address: services.SubAccountGeneral,
				},
			},
			Amount: &types.Amount{
				Value:    "-1000",
				Currency: services.OasisCurrency,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 2,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: dstAddress,
				SubAccount: &types.SubAccountIdentifier{
					Address: services.SubAccountGeneral,
				},
			},
			Amount: &types.Amount{
				Value:    "1000",
				Currency: services.OasisCurrency,
			},
		},
	}
	txTransfer := &transaction.Transaction{
		Nonce:  dummyNonce,
		Fee:    fee100,
		Method: api.MethodTransfer,
		Body: cbor.Marshal(api.Transfer{
			To:     dstAddr,
			Tokens: *quantity.NewFromUint64(1000),
		}),
	}
	opsBurn := []*types.Operation{
		fee100Op,
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type: services.OpBurn,
			Account: &types.AccountIdentifier{
				Address: testEntityAddress,
				SubAccount: &types.SubAccountIdentifier{
					Address: services.SubAccountGeneral,
				},
			},
			Amount: &types.Amount{
				Value:    "-1000",
				Currency: services.OasisCurrency,
			},
		},
	}
	txBurn := &transaction.Transaction{
		Nonce:  dummyNonce,
		Fee:    fee100,
		Method: api.MethodBurn,
		Body: cbor.Marshal(api.Burn{
			Tokens: *quantity.NewFromUint64(1000),
		}),
	}

	rc := client.NewAPIClient(client.NewConfiguration("http://localhost:8080", "rosetta-sdk-go", nil))

	r1, re, err := rc.NetworkAPI.NetworkList(context.Background(), &types.MetadataRequest{})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	if len(r1.NetworkIdentifiers) != 1 {
		panic("len(r1.NetworkIdentifiers)")
	}
	fmt.Println("network identifiers", r1.NetworkIdentifiers)
	ni := r1.NetworkIdentifiers[0]

	for _, tt := range []struct {
		name      string
		ops       []*types.Operation
		reference *transaction.Transaction
	}{
		{"transfer", opsTransfer, txTransfer},
		{"burn", opsBurn, txBurn},
	} {
		r2, re, err := rc.ConstructionAPI.ConstructionPayloads(context.Background(), &types.ConstructionPayloadsRequest{
			NetworkIdentifier: ni,
			Operations:        tt.ops,
			Metadata: map[string]interface{}{
				services.NonceKey: dummyNonce,
			},
		})
		if err != nil {
			panic(fmt.Errorf("%s: %w", tt.name, err))
		}
		if re != nil {
			panic(fmt.Errorf("%s: %v", tt.name, re))
		}
		fmt.Println(tt.name, "unsigned transaction", r2.UnsignedTransaction)
		fmt.Println(tt.name, "signing payloads", r2.Payloads)

		txBuf, err := hex.DecodeString(r2.UnsignedTransaction)
		if err != nil {
			panic(fmt.Errorf("%s: %w", tt.name, err))
		}
		refBuf := cbor.Marshal(tt.reference)
		if !bytes.Equal(txBuf, refBuf) {
			refHex := hex.EncodeToString(refBuf)
			fmt.Println(tt.name, "reference transaction", refHex)
			panic(fmt.Errorf("%s: transaction mismatch", tt.name))
		}
	}
}
