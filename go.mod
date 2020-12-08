module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.0-rc4-oasis2.3

require (
	github.com/coinbase/rosetta-cli v0.6.3
	github.com/coinbase/rosetta-sdk-go v0.6.4
	github.com/dgraph-io/badger v1.6.2
	github.com/oasisprotocol/ed25519 v0.0.0-20201103162138-a1dadbe24dd5
	github.com/oasisprotocol/oasis-core/go v0.2012.3
	github.com/vmihailenco/msgpack/v5 v5.1.0
	google.golang.org/grpc v1.34.0
)
