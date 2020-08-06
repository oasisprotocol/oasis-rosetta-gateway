package common

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/ed25519"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	"github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-core-rosetta-gateway/services"
)

const DstAddress = "oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"

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
