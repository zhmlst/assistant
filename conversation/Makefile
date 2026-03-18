ifneq (,$(wildcard .env))
    include .env
    export
endif

GOPATH := $(shell go env GOPATH)

all: gen

$(GOPATH)/bin/buf:
	go install github.com/bufbuild/buf/cmd/buf@latest

$(GOPATH)/bin/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

gen: $(GOPATH/bin/buf) $(GOPATH)/bin/protoc-gen-go
	buf generate .. --template ../buf.gen.yaml --path ../proto/conversation
	go mod tidy

run:
	go run ./cmd
