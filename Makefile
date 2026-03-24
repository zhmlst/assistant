include .env
export

.PHONY: up
up:
	podman compose up

.PHONY: conversation
conversation:
	make -C conversation
