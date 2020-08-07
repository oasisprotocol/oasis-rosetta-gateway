module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.14

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.33.4-oasis2

require (
	github.com/coinbase/rosetta-cli v0.4.0
	github.com/coinbase/rosetta-sdk-go v0.3.3
	github.com/dgraph-io/badger v1.6.1
	github.com/oasisprotocol/ed25519 v0.0.0-20200528083105-55566edd6df0
	github.com/oasisprotocol/oasis-core/go v0.0.0-20200702171459-20d1a2dc6b66
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.1
	google.golang.org/grpc v1.29.1
)
