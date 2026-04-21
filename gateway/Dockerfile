FROM docker.io/library/golang:1.26-bookworm AS builder

WORKDIR /app

COPY gateway/go.mod gateway/go.sum ./gateway/
COPY go/go.mod go/go.sum ./go/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -o /app ./gateway/cmd/app

FROM docker.io/library/alpine:latest

WORKDIR /root/

COPY --from=builder /app .

CMD ["./app"]
