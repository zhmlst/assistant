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

.PHONY: install
install: all
	mkdir -p /opt/assistant
	cp -r . /opt/assistant
	cp assistant.service /etc/systemd/system/

.PHONY: uninstall
uninstall:
	rm -f /etc/systemd/system/assistant.service
