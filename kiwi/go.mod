module github.com/vegris/alas-go/kiwi

go 1.23.6

replace github.com/vegris/alas-go/shared => ../shared

require (
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.7.1
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.1
	github.com/segmentio/kafka-go v0.4.47
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/vegris/alas-go/shared v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.14.0 // indirect
)
