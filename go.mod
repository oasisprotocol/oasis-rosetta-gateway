module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.9-oasis2

require (
	github.com/coinbase/rosetta-cli v0.7.3
	github.com/coinbase/rosetta-sdk-go v0.7.2
	github.com/oasisprotocol/oasis-core/go v0.2103.9
	google.golang.org/grpc v1.41.0
)
