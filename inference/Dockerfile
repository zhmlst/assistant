FROM docker.io/library/golang:1.26-bookworm AS builder

WORKDIR /src

COPY inference/go.mod inference/go.sum ./inference/
COPY go/go.mod go/go.sum ./go/

WORKDIR /src/inference
RUN go mod download

WORKDIR /src
COPY . .

WORKDIR /src/inference

RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg/mod \
	CGO_ENABLED=1 GOOS=linux go build \
	-ldflags "-linkmode external -extldflags '-static'" \
	-o /app ./cmd/app

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app ./app
CMD ["./app"]
