include .env
export

.PHONY: all
all: image

.env: env.example
	cp env.example .env

.PHONY: run
run:
	go run ./cmd/app

.PHONY: image
image:
	cd .. && $(DOCKER) build -t inference -f inference/Dockerfile .
