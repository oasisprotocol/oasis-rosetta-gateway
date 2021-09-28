module github.com/oasisprotocol/oasis-core-rosetta-gateway

go 1.15

replace github.com/tendermint/tendermint => github.com/oasisprotocol/tendermint v0.34.9-oasis2

require (
	github.com/blevesearch/bleve v1.0.14 // indirect
	github.com/coinbase/rosetta-cli v0.4.0
	github.com/coinbase/rosetta-sdk-go v0.3.3
	github.com/dgraph-io/badger v1.6.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/marten-seemann/qtls v0.10.0 // indirect
	github.com/oasisprotocol/ed25519 v0.0.0-20210127160119-f7017427c1ea
	github.com/oasisprotocol/oasis-core/go v0.2103.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.4
	google.golang.org/grpc v1.41.0
)
