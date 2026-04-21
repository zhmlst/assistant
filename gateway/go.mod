module github.com/zhmlst/assistant/gateway

go 1.26.1

replace github.com/zhmlst/assistant/go => ../go

replace github.com/zhmlst/assistant/conversation => ../conversation

require (
	github.com/caarlos0/env/v11 v11.4.0
	github.com/google/uuid v1.6.0
	github.com/zhmlst/assistant/conversation v0.0.0-00010101000000-000000000000
	github.com/zhmlst/assistant/go v0.0.0-20260419174741-f4fbef8143dc
	google.golang.org/grpc v1.79.3
)

require (
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
