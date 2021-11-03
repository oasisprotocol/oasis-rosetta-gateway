module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.9-oasis2

require (
	github.com/coinbase/rosetta-cli v0.7.1
	github.com/coinbase/rosetta-sdk-go v0.7.0
	github.com/dgraph-io/badger v1.6.2
	github.com/ethereum/go-ethereum v1.10.9 // indirect
	github.com/oasisprotocol/ed25519 v0.0.0-20210127160119-f7017427c1ea
	github.com/oasisprotocol/oasis-core/go v0.2103.5
	github.com/vmihailenco/msgpack/v5 v5.3.4
	google.golang.org/grpc v1.41.0
)
