module github.com/oasislabs/oasis-core-rosetta-gateway

go 1.14

replace github.com/tendermint/tendermint => github.com/oasislabs/tendermint v0.33.4-oasis1

require (
	github.com/coinbase/rosetta-sdk-go v0.1.5
	github.com/oasislabs/oasis-core/go v0.0.0-20200527082817-11046d140273
	google.golang.org/grpc v1.29.1
)
