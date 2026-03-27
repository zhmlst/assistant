include .env
export

.env: env.example
	cp env.example .env

DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

.PHONY: all
all: image

$(GOPATH)/bin/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

$(GOPATH)/bin/buf:
	go install github.com/bufbuild/buf/cmd/buf@latest

$(GOPATH)/bin/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$(GOPATH)/bin/protoc-gen-go-grpc:
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: gen
gen: $(GOPATH)/bin/buf $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-go-grpc $(GOPATH)/bin/stringer
	buf generate .. --template ../buf.gen.yaml --path ../proto/conversation
	go generate ./...
	gofmt -w .
	go mod tidy

.PHONY: postgres
postgres:
	@if [ "$$($(DOCKER) ps -q -f 'name=^assistant-postgres$$')" ]; then \
		:; \
	elif [ "$$($(DOCKER) ps -qa -f 'name=^assistant-postgres$$')" ]; then \
		$(DOCKER) start assistant-postgres; \
	else \
		$(DOCKER) run -d --name assistant-postgres -e POSTGRES_PASSWORD=2345 -p 5432:5432 docker.io/library/postgres:17-alpine ; \
	fi

.PHONY: run
run: gen postgres
	go run ./cmd/app

$(GOPATH)/bin/grpcui:
	go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

.PHONY: ui
ui: $(GOPATH)/bin/grpcui
	grpcui -plaintext localhost:50050

.PHONY: image
image: gen
	cd .. && podman build -t conversation -f ./conversation/Dockerfile .
