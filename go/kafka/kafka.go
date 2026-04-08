package kafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const EventTypeKey = "event-type"

type Handler func(ctx context.Context, msg *kafka.Message) error

type Config struct {
	Brokers string
	GroupID string
}

func (c *Config) Map() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
		"group.id":          c.GroupID,
	}
}

type Consumer struct {
	consumer *kafka.Consumer
	routes   map[string]Handler
}

func New(
	consumer *kafka.Consumer,
	routes map[string]Handler,
) *Consumer {
	return &Consumer{
		consumer: consumer,
		routes:   routes,
	}
}

func (h *Consumer) Consume(ctx context.Context) error {
	for {
		msg, err := h.consumer.ReadMessage(500 * time.Millisecond)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok &&
				kerr.Code() == kafka.ErrTimedOut ||
				kerr.IsRetriable() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			} else {
				return err
			}
		}

		go func(msg *kafka.Message) {
			event := lookupHeader(msg.Headers, EventTypeKey)
			handler, ok := h.routes[string(event)]
			if !ok {
				return
			}
			_ = handler(ctx, msg)
		}(msg)
	}
}

func lookupHeader(headers []kafka.Header, key string) []byte {
	for _, header := range headers {
		if header.Key == key {
			return header.Value
		}
	}
	return nil
}
