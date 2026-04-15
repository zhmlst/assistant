GOPATH = $(shell go env GOPATH)
DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

export PATH := $(GOPATH)/bin:$(PATH)

-include .env
export

.env: env.example
	cp -n env.example .env || true

