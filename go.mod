module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.14

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.0-rc3-oasis1

require (
	github.com/coinbase/rosetta-cli v0.4.0
	github.com/coinbase/rosetta-sdk-go v0.3.3
	github.com/dgraph-io/badger v1.6.1
	github.com/oasisprotocol/ed25519 v0.0.0-20200528083105-55566edd6df0
	github.com/oasisprotocol/oasis-core/go v0.20.10-0.20200904184626-90a7b2e85ebb
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.1
	google.golang.org/grpc v1.31.0
)
