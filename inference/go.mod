module github.com/zhmlst/assistant/inference

replace github.com/zhmlst/assistant/go => ../go

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.14.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/zhmlst/assistant/conversation v0.0.0-20260414014601-ff05663ac364
	github.com/zhmlst/assistant/go v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
)
