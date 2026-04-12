include .env
export

all: image

.env: env.example
	cp env.example .env

run:
	go run ./cmd/app

image:
	cd .. && $(DOCKER) build -t inference -f inference/Dockerfile .
