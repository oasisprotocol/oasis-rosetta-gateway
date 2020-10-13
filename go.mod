module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.0-rc4-oasis2

require (
	github.com/coinbase/rosetta-cli v0.5.10
	github.com/coinbase/rosetta-sdk-go v0.5.2-0.20201006190307-bf4606611446
	github.com/dgraph-io/badger v1.6.2
	github.com/oasisprotocol/ed25519 v0.0.0-20200819094954-65138ca6ec7c
	github.com/oasisprotocol/oasis-core/go v0.2010.1
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.1
	google.golang.org/grpc v1.32.0
)
