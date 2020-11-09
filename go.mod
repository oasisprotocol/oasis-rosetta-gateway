module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.0-rc4-oasis2.3

require (
	github.com/coinbase/rosetta-cli v0.5.23
	github.com/coinbase/rosetta-sdk-go v0.5.11-0.20201106052603-0322f7b18be8
	github.com/dgraph-io/badger v1.6.2
	github.com/oasisprotocol/ed25519 v0.0.0-20201103162138-a1dadbe24dd5
	github.com/oasisprotocol/oasis-core/go v0.2012.1
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.9
	google.golang.org/grpc v1.32.0
)
