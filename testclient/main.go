package main

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/ed25519"
	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
)

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
			Type: services.OpDummyLiteral,
			Account: &types.AccountIdentifier{
				Address: testEntityAddress,
			},
			Metadata: map[string]interface{}{
				services.LiteralKey: "a463666565a26367617319271066616d6f756e744064626f6479a267786665725f746f55006dd9ae2525cd42c3a8988383b1f041fb91bbb1916b786665725f746f6b656e734203e8656e6f6e636500666d6574686f64707374616b696e672e5472616e73666572",
			},
		},
	}
	fmt.Println("operations", ops)

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
	fmt.Println("metadata options", r2.Options)

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
	fmt.Println("metadata", r3.Metadata)

	r4, re, err := rc.ConstructionAPI.ConstructionPayloads(context.Background(), &types.ConstructionPayloadsRequest{
		NetworkIdentifier: ni,
		Operations:        ops,
	})
	if err != nil {
		panic(err)
	}
	if re != nil {
		panic(re)
	}
	fmt.Println("unsigned transaction", r4.UnsignedTransaction)
	fmt.Println("signing payloads", r4.Payloads)

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
	fmt.Println("transaction metadata", r6.Metadata)
}
