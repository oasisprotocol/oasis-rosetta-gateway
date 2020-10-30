module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.0-rc4-oasis2

require (
	github.com/coinbase/rosetta-cli v0.5.19
	github.com/coinbase/rosetta-sdk-go v0.5.9-0.20201029210921-d7499a34c1f6
	github.com/dgraph-io/badger v1.6.2
	github.com/oasisprotocol/ed25519 v0.0.0-20200819094954-65138ca6ec7c
	github.com/oasisprotocol/oasis-core/go v0.2011.3
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.9
	google.golang.org/grpc v1.32.0
)
