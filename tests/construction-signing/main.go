package main

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/ed25519"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
	"reflect"
)

const dstAddress = "oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"

func dumpJSON(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func main() {
	_, signer, err := entity.TestEntity()
	if err != nil {
		panic(err)
	}
	testEntityAddress := api.NewAddress(signer.Public()).String()

	seed := sha512.Sum512_256([]byte("ekiden test entity key seed"))
	priv := ed25519.NewKeyFromSeed(seed[:])
	pub := priv.Public().(ed25519.PublicKey)
	var pub2 signature.PublicKey
	if err = pub2.UnmarshalBinary(pub); err != nil {
		panic(err)
	}
	if pub2 != signer.Public() {
		panic(fmt.Sprintf("public key mismatch %s %s", pub2, signer.Public()))
	}
	rs := keys.SignerEdwards25519{
		KeyPair: &keys.KeyPair{
			PublicKey: &types.PublicKey{
				Bytes:     pub,
				CurveType: types.Edwards25519,
			},
			PrivateKey: priv[:32],
		},
	}

	ops := []*types.Operation{
		{
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
				Value:    "-0",
				Currency: services.OasisCurrency,
			},
		},
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
	fmt.Println("operations", dumpJSON(ops))

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
	fmt.Println("network identifiers", dumpJSON(r1.NetworkIdentifiers))
	ni := r1.NetworkIdentifiers[0]

	r2, re, err := rc.ConstructionAPI.ConstructionPreprocess(context.Background(), &types.ConstructionPreprocessRequest{
		NetworkIdentifier: ni,
		Operations:        ops,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("metadata options", dumpJSON(r2.Options))

	r3, re, err := rc.ConstructionAPI.ConstructionMetadata(context.Background(), &types.ConstructionMetadataRequest{
		NetworkIdentifier: ni,
		Options:           r2.Options,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("metadata", dumpJSON(r3.Metadata))

	r4, re, err := rc.ConstructionAPI.ConstructionPayloads(context.Background(), &types.ConstructionPayloadsRequest{
		NetworkIdentifier: ni,
		Operations:        ops,
		Metadata:          r3.Metadata,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("unsigned transaction", r4.UnsignedTransaction)
	fmt.Println("signing payloads", dumpJSON(r4.Payloads))

	r4p, re, err := rc.ConstructionAPI.ConstructionParse(context.Background(), &types.ConstructionParseRequest{
		NetworkIdentifier: ni,
		Signed:            false,
		Transaction:       r4.UnsignedTransaction,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("unsigned operations", dumpJSON(r4p.Operations))
	fmt.Println("unsigned signers", dumpJSON(r4p.Signers))
	fmt.Println("unsigned metadata", dumpJSON(r4p.Metadata))
	r4pRef := &types.ConstructionParseResponse{
		Operations: []*types.Operation{
			{
				OperationIdentifier: ops[0].OperationIdentifier,
				Type:                ops[0].Type,
				Account: &types.AccountIdentifier{
					Address:    services.FromPlaceholder,
					SubAccount: ops[0].Account.SubAccount,
				},
				Amount: ops[0].Amount,
				Metadata: map[string]interface{}{
					services.FeeGasKey: float64(services.DefaultGas),
				},
			},
			{
				OperationIdentifier: ops[1].OperationIdentifier,
				Type:                ops[1].Type,
				Account: &types.AccountIdentifier{
					Address:    services.FromPlaceholder,
					SubAccount: ops[1].Account.SubAccount,
				},
				Amount: ops[1].Amount,
			},
			ops[2],
		},
		Metadata: r3.Metadata,
	}
	if !reflect.DeepEqual(r4p, r4pRef) {
		fmt.Println("unsigned transaction parsed", dumpJSON(r4p))
		fmt.Println("reference", dumpJSON(r4pRef))
		panic(fmt.Errorf("unsigned transaction parsed wrong"))
	}

	var sigs []*types.Signature
	for i, sp := range r4.Payloads {
		if sp.Address != testEntityAddress {
			panic(i)
		}
		sig, err := rs.Sign(sp, sp.SignatureType)
		if err != nil {
			panic(err)
		}
		sigs = append(sigs, sig)
	}

	r5, re, err := rc.ConstructionAPI.ConstructionCombine(context.Background(), &types.ConstructionCombineRequest{
		NetworkIdentifier:   ni,
		UnsignedTransaction: r4.UnsignedTransaction,
		Signatures:          sigs,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("signed transaction", r5.SignedTransaction)

	r5p, re, err := rc.ConstructionAPI.ConstructionParse(context.Background(), &types.ConstructionParseRequest{
		NetworkIdentifier: ni,
		Signed:            true,
		Transaction:       r5.SignedTransaction,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("signed operations", dumpJSON(r5p.Operations))
	fmt.Println("signed signers", dumpJSON(r5p.Signers))
	fmt.Println("signed metadata", dumpJSON(r5p.Metadata))
	r5pRef := &types.ConstructionParseResponse{
		Operations: []*types.Operation{
			{
				OperationIdentifier: ops[0].OperationIdentifier,
				Type:                ops[0].Type,
				Account:             ops[0].Account,
				Amount:              ops[0].Amount,
				Metadata: map[string]interface{}{
					services.FeeGasKey: float64(services.DefaultGas),
				},
			},
			ops[1],
			ops[2],
		},
		Signers:  []string{testEntityAddress},
		Metadata: r3.Metadata,
	}
	if !reflect.DeepEqual(r5p, r5pRef) {
		fmt.Println("signed transaction parsed", dumpJSON(r5p))
		fmt.Println("reference", dumpJSON(r5pRef))
		panic(fmt.Errorf("signed transaction parsed wrong"))
	}

	r6, re, err := rc.ConstructionAPI.ConstructionSubmit(context.Background(), &types.ConstructionSubmitRequest{
		NetworkIdentifier: ni,
		SignedTransaction: r5.SignedTransaction,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("transaction hash", r6.TransactionIdentifier.Hash)
	fmt.Println("transaction metadata", dumpJSON(r6.Metadata))
}
