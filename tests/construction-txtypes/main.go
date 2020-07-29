package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
	"reflect"
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
	opsAddEscrow := []*types.Operation{
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
	opsReclaimEscrow := []*types.Operation{
		fee100Op,
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: dstAddress,
				SubAccount: &types.SubAccountIdentifier{
					Address: services.SubAccountEscrow,
				},
			},
			Amount: &types.Amount{
				Value:    "-1000",
				Currency: services.PoolShare,
			},
		},
	}
	txReclaimEscrow := &transaction.Transaction{
		Nonce:  dummyNonce,
		Fee:    fee100,
		Method: api.MethodReclaimEscrow,
		Body: cbor.Marshal(api.ReclaimEscrow{
			Account: dstAddr,
			Shares:  *quantity.NewFromUint64(1000),
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
		{"add escrow", opsAddEscrow, txAddEscrow},
		{"reclaim escrow", opsReclaimEscrow, txReclaimEscrow},
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
		fmt.Println(tt.name, "operations", r3.Operations)
		fmt.Println(tt.name, "operations", r3.Operations)
		fmt.Println(tt.name, "signers", r3.Signers)
		fmt.Println(tt.name, "metadata", r3.Metadata)

		var parsedOpsResolved []*types.Operation
		for _, op := range r3.Operations {
			if op.Account.Address == services.FromPlaceholder {
				op = &*op
				op.Account = &*op.Account
				op.Account.Address = testEntityAddress
			}
			parsedOpsResolved = append(parsedOpsResolved, op)
		}
		if !reflect.DeepEqual(parsedOpsResolved, tt.ops) {
			parsedOpsJSON, err := json.Marshal(r3.Operations)
			if err != nil {
				panic(fmt.Errorf("%s marshal parsed operations: %w", tt.name, err))
			}
			fmt.Println(tt.name, "parsed operations", string(parsedOpsJSON))
			parsedOpsResolvedJSON, err := json.Marshal(parsedOpsResolved)
			if err != nil {
				panic(fmt.Errorf("%s marshal parsed operations resolved: %w", tt.name, err))
			}
			fmt.Println(tt.name, "parsed operations resolved", string(parsedOpsResolvedJSON))
			referenceOpsJSON, err := json.Marshal(tt.ops)
			if err != nil {
				panic(fmt.Errorf("%s marshal reference operations: %w", tt.name, err))
			}
			fmt.Println(tt.name, "reference operations", referenceOpsJSON)
			panic(fmt.Errorf("%s: operations mismatch", tt.name))
		}
	}
}
