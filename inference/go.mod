module github.com/zhmlst/assistant/inference

replace github.com/zhmlst/assistant/go => ../go

go 1.26.1

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.14.0
	github.com/google/uuid v1.6.0
	github.com/zhmlst/assistant/conversation v0.0.0-20260404025434-b08a12fbfc7a
	github.com/zhmlst/assistant/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
	google.golang.org/grpc v1.79.3 // indirect
)
