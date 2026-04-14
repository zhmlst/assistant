include .env
export

DOCKER ?= $(shell command -v podman 2> /dev/null || echo docker)

.PHONY: all
all: conversation inference

.PHONY: up
up: conversation inference
	$(DOCKER) compose up -d --build

.PHONY: run
run: conversation inference
	$(DOCKER) compose up --build

.PHONY: down
down:
	$(DOCKER) compose down

.PHONY: conversation
conversation:
	make -C conversation

.PHONY: inference
inference:
	make -C inference

.PHONY: install
install: all
	mkdir -p /opt/assistant
	cp -r . /opt/assistant
	cp assistant.service /etc/systemd/system/

.PHONY: uninstall
uninstall:
	rm -f /etc/systemd/system/assistant.service
