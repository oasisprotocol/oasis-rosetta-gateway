module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.9-oasis2

require (
	github.com/coinbase/rosetta-cli v0.4.0
	github.com/coinbase/rosetta-sdk-go v0.6.10
	github.com/dgraph-io/badger v1.6.2
	github.com/oasisprotocol/ed25519 v0.0.0-20210127160119-f7017427c1ea
	github.com/oasisprotocol/oasis-core/go v0.2102.8
	github.com/vmihailenco/msgpack/v5 v5.1.4
	google.golang.org/grpc v1.39.1
)
