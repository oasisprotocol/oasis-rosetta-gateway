package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/tests/common"
)

const dummyNonce = 3

func main() {
	testEntityAddress, _ := common.TestEntity()

	var dstAddr api.Address
	if err := dstAddr.UnmarshalText([]byte(common.DstAddress)); err != nil {
		panic(err)
	}
	fee100Op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: 0,
		},
		Type: services.OpTransfer,
		Account: &types.AccountIdentifier{
			Address: testEntityAddress,
		},
		Amount: &types.Amount{
			Value:    "-100",
			Currency: services.OasisCurrency,
		},
		Metadata: map[string]interface{}{
			services.FeeGasKey: 10001.,
		},
	}
	fee100 := &transaction.Fee{
		Amount: *quantity.NewFromUint64(100),
		Gas:    10001,
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
				Address: common.DstAddress,
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
	opsAddEscrow := []*types.Operation{
		fee100Op,
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: testEntityAddress,
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
				Address: common.DstAddress,
				SubAccount: &types.SubAccountIdentifier{
					Address: services.SubAccountEscrow,
				},
			},
			Amount: &types.Amount{
				Value:    "1000",
				Currency: services.OasisCurrency,
			},
		},
	}
	txAddEscrow := &transaction.Transaction{
		Nonce:  dummyNonce,
		Fee:    fee100,
		Method: api.MethodAddEscrow,
		Body: cbor.Marshal(api.Escrow{
			Account: dstAddr,
			Tokens:  *quantity.NewFromUint64(1000),
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
	fmt.Println("network identifiers", common.DumpJSON(r1.NetworkIdentifiers))
	ni := r1.NetworkIdentifiers[0]

	for _, tt := range []struct {
		name      string
		ops       []*types.Operation
		reference *transaction.Transaction
	}{
		{"transfer", opsTransfer, txTransfer},
		{"burn", opsBurn, txBurn},
		{"add escrow", opsAddEscrow, txAddEscrow},
	} {
		r2, re, err := rc.ConstructionAPI.ConstructionPayloads(context.Background(), &types.ConstructionPayloadsRequest{
			NetworkIdentifier: ni,
			Operations:        tt.ops,
			Metadata: map[string]interface{}{
				services.NonceKey: dummyNonce,
			},
		})
		if err != nil {
			panic(fmt.Errorf("%s payloads: %w", tt.name, err))
		}
		if re != nil {
			panic(fmt.Errorf("%s payloads: %v", tt.name, re))
		}
		fmt.Println(tt.name, "unsigned transaction", r2.UnsignedTransaction)
		fmt.Println(tt.name, "signing payloads", common.DumpJSON(r2.Payloads))

		var ut services.UnsignedTransaction
		if err := json.Unmarshal([]byte(r2.UnsignedTransaction), &ut); err != nil {
			panic(err)
		}
		if !reflect.DeepEqual(&ut.Tx, tt.reference) {
			fmt.Println(tt.name, "reference transaction", common.DumpJSON(tt.reference))
			panic(fmt.Errorf("%s: transaction mismatch", tt.name))
		}

		r3, re, err := rc.ConstructionAPI.ConstructionParse(context.Background(), &types.ConstructionParseRequest{
			NetworkIdentifier: ni,
			Signed:            false,
			Transaction:       r2.UnsignedTransaction,
		})
		if err != nil {
			panic(fmt.Errorf("%s parse: %w", tt.name, err))
		}
		if re != nil {
			panic(fmt.Errorf("%s parse: %v", tt.name, re))
		}
		fmt.Println(tt.name, "parsed operations", common.DumpJSON(r3.Operations))
		fmt.Println(tt.name, "parsed signers", common.DumpJSON(r3.Signers))
		fmt.Println(tt.name, "parsed metadata", common.DumpJSON(r3.Metadata))

		if !reflect.DeepEqual(r3.Operations, tt.ops) {
			fmt.Println(tt.name, "parsed operations", common.DumpJSON(r3.Operations))
			fmt.Println(tt.name, "reference operations", common.DumpJSON(tt.ops))
			panic(fmt.Errorf("%s: operations mismatch", tt.name))
		}
	}
}
