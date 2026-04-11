include .env
export

DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

.PHONY: up
up: conversation inference
	$(DOCKER) compose up -d

down:
	$(DOCKER) compose down

.PHONY: conversation
conversation:
	make -C conversation

.PHONY: inference
inference:
	make -C inference
