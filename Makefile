include .env
export

DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

.PHONY: up
up: conversation
	$(DOCKER) compose up -d

down:
	$(DOCKER) compose down

.PHONY: conversation
conversation:
	make -C conversation
