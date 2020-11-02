package common

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/ed25519"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
)

const DstAddressText = "oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"

var (
	TestEntityAddressText, _ = TestEntity()

	DstAddress        = unmarshalAddressOrPanic(DstAddressText)
	TestEntityAddress = unmarshalAddressOrPanic(TestEntityAddressText)
)

func DumpJSON(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func TestEntity() (string, *keys.KeyPair) {
	_, signer, err := entity.TestEntity()
	if err != nil {
		panic(err)
	}
	address := services.StringFromAddress(api.NewAddress(signer.Public()))

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
	kp := &keys.KeyPair{
		PublicKey: &types.PublicKey{
			Bytes:     pub,
			CurveType: types.Edwards25519,
		},
		PrivateKey: priv[:32],
	}

	return address, kp
}

func unmarshalAddressOrPanic(addrText string) (addr api.Address) {
	if err := addr.UnmarshalText([]byte(addrText)); err != nil {
		panic(err)
	}
	return
}

// NewRosettaClient returns a new Rosetta API Client for tests or panics.
func NewRosettaClient() (*client.APIClient, *types.NetworkIdentifier) {
	rClient := client.NewAPIClient(client.NewConfiguration("http://localhost:8080", "rosetta-sdk-go", nil))
	nl, rErr, err := rClient.NetworkAPI.NetworkList(context.Background(), &types.MetadataRequest{})
	if err != nil {
		panic(err)
	}
	if rErr != nil {
		panic(rErr)
	}
	if len(nl.NetworkIdentifiers) != 1 {
		panic("there should only be one network identifier")
	}
	fmt.Println("network identifiers", DumpJSON(nl.NetworkIdentifiers))
	return rClient, nl.NetworkIdentifiers[0]
}
