include .env
export

.env: env.example
	cp env.example .env

DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

all: image

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

postgres:
	$(DOCKER) inspect postgres > /dev/null 2>&1 || \
		$(DOCKER) run -d --name postgres -e POSTGRES_PASSWORD=$(POSTGRES_PASS) -p 5432:5432 docker.io/library/postgres:17-alpine
	$(DOCKER) start postgres

run: gen postgres
	go run ./cmd/app

.PHONY: image
image: gen
	cd .. && podman build -t conversation -f ./conversation/Dockerfile .
