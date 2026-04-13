FROM docker.io/library/golang:1.26-bookworm AS builder

WORKDIR /src

COPY conversation/go.mod conversation/go.sum ./conversation/
COPY go/go.mod go/go.sum ./go/

WORKDIR /src/conversation
RUN go mod download

WORKDIR /src
COPY . .

WORKDIR /src/conversation
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-linkmode external -extldflags '-static'" \
    -o /app ./cmd/app

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app ./app
CMD ["./app"]
