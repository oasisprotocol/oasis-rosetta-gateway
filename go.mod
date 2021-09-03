module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.9-oasis2

require (
	github.com/coinbase/rosetta-cli v0.7.0
	github.com/coinbase/rosetta-sdk-go v0.7.0
	github.com/dgraph-io/badger v1.6.2
	github.com/google/addlicense v0.0.0-20200622132530-df58acafd6d5 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
	github.com/oasisprotocol/ed25519 v0.0.0-20210127160119-f7017427c1ea
	github.com/oasisprotocol/oasis-core/go v0.2102.8
	github.com/segmentio/golines v0.0.0-20200306054842-869934f8da7b // indirect
	github.com/ugorji/go v1.1.4 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.11 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.4
	google.golang.org/grpc v1.39.1
)
