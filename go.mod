module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.14

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.33.4-oasis2

require (
	github.com/coinbase/rosetta-sdk-go v0.2.0
	github.com/oasisprotocol/oasis-core/go v0.0.0-20200702171459-20d1a2dc6b66
	google.golang.org/grpc v1.29.1
)
