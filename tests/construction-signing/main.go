package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/oasisprotocol/oasis-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-rosetta-gateway/tests/common"
)

func main() { //nolint:funlen
	testEntityAddress, testEntityKeyPair := common.TestEntity()
	rs := keys.SignerEdwards25519{KeyPair: testEntityKeyPair}

	ops := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
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
				Index: 1,
			},
			Type: services.OpTransfer,
			Account: &types.AccountIdentifier{
				Address: common.DstAddressText,
			},
			Amount: &types.Amount{
				Value:    "1000",
				Currency: services.OasisCurrency,
			},
			RelatedOperations: []*types.OperationIdentifier{
				{
					Index: 0,
				},
			},
		},
	}
	fmt.Println("operations", common.DumpJSON(ops))

	rc, ni := common.NewRosettaClient()

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
	fmt.Println("metadata options", common.DumpJSON(r2.Options))

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
	fmt.Println("metadata", common.DumpJSON(r3.Metadata))

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
	fmt.Println("signing payloads", common.DumpJSON(r4.Payloads))

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
	fmt.Println("unsigned operations", common.DumpJSON(r4p.Operations))
	fmt.Println("unsigned signers", common.DumpJSON(r4p.AccountIdentifierSigners))
	fmt.Println("unsigned metadata", common.DumpJSON(r4p.Metadata))
	r4pRef := &types.ConstructionParseResponse{
		Operations: ops,
		Metadata:   r3.Metadata,
	}
	if !reflect.DeepEqual(r4p, r4pRef) {
		fmt.Println("unsigned transaction parsed", common.DumpJSON(r4p))
		fmt.Println("reference", common.DumpJSON(r4pRef))
		panic(fmt.Errorf("unsigned transaction parsed wrong"))
	}

	sigs := make([]*types.Signature, 0, len(r4.Payloads))
	for i, sp := range r4.Payloads {
		if sp.AccountIdentifier.Address != testEntityAddress {
			panic(i)
		}
		sig, err2 := rs.Sign(sp, sp.SignatureType)
		if err2 != nil {
			panic(err2)
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
	fmt.Println("signed operations", common.DumpJSON(r5p.Operations))
	fmt.Println("signed signers", common.DumpJSON(r5p.AccountIdentifierSigners))
	fmt.Println("signed metadata", common.DumpJSON(r5p.Metadata))
	r5pRef := &types.ConstructionParseResponse{
		Operations: ops,
		AccountIdentifierSigners: []*types.AccountIdentifier{{
			Address:    testEntityAddress,
			SubAccount: nil,
			Metadata:   nil,
		}},
		Metadata: r3.Metadata,
	}
	if !reflect.DeepEqual(r5p, r5pRef) {
		fmt.Println("signed transaction parsed", common.DumpJSON(r5p))
		fmt.Println("reference", common.DumpJSON(r5pRef))
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
	fmt.Println("transaction metadata", common.DumpJSON(r6.Metadata))

	r7, re, err := rc.MempoolAPI.Mempool(context.Background(), &types.NetworkRequest{
		NetworkIdentifier: ni,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("transactions in mempool", len(r7.TransactionIdentifiers))

	for _, ti := range r7.TransactionIdentifiers {
		tr, re, err := rc.MempoolAPI.MempoolTransaction(context.Background(), &types.MempoolTransactionRequest{
			NetworkIdentifier:     ni,
			TransactionIdentifier: ti,
		})
		if err != nil {
			panic(err)
		}
		if re != nil {
			panic(re)
		}
		fmt.Println("transaction from mempool", common.DumpJSON(tr))
	}
}
