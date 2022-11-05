package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature" //nolint:typecheck
	"github.com/oasisprotocol/oasis-core/go/common/entity"
	genesis "github.com/oasisprotocol/oasis-core/go/genesis/api"
	cmdGenesis "github.com/oasisprotocol/oasis-core/go/oasis-node/cmd/common/genesis"
	"github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-rosetta-gateway/services"
)

const DstAddressText = "oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"

var (
	TestEntityAddressText, _ = TestEntity()

	DstAddress        = unmarshalAddressOrPanic(DstAddressText)
	TestEntityAddress = unmarshalAddressOrPanic(TestEntityAddressText)

	TestEntityAmount = getTestEntityAmount()
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
	pub := signer.Public()
	address := services.StringFromAddress(api.NewAddress(pub))
	priv := signer.(signature.UnsafeSigner).UnsafeBytes()

	kp := &keys.KeyPair{
		PublicKey: &types.PublicKey{
			Bytes:     pub[:],
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

// NOTE: Consider exposing the test entity's general balance directly in Oasis
// Core to avoid this manual extraction procedure.
func getTestEntityAmount() *types.Amount {
	genesisDoc := &genesis.Document{}
	stakingState, err := cmdGenesis.NewAppendableStakingState()
	if err != nil {
		panic(err)
	}
	stakingState.DebugTestEntity = true
	stakingState.State.TokenSymbol = "TEST"
	if err := stakingState.AppendTo(genesisDoc); err != nil {
		panic(err)
	}

	testEntityAccount := genesisDoc.Staking.Ledger[TestEntityAddress]
	return &types.Amount{
		Value:    testEntityAccount.General.Balance.String(),
		Currency: services.OasisCurrency,
	}
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
