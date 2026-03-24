include .env
export

all: conversation

.PHONY: up
up:
	podman compose up

.PHONY: conversation
conversation:
	make -C conversation
