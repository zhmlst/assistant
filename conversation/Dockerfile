FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /src

COPY . .

WORKDIR /src/conversation

RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/app

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app ./app

CMD ["/app"]
