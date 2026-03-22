ifneq (,$(wildcard .env))
    include .env
    export
endif

GOPATH := $(shell go env GOPATH)

all: gen

$(GOPATH)/bin/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

$(GOPATH)/bin/buf:
	go install github.com/bufbuild/buf/cmd/buf@latest

$(GOPATH)/bin/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$(GOPATH)/bin/protoc-gen-go-grpc:
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

gen: $(GOPATH)/bin/buf $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-go-grpc $(GOPATH)/bin/stringer
	buf generate .. --template ../buf.gen.yaml --path ../proto/conversation
	go generate ./...
	go mod tidy

run:
	go run ./cmd
