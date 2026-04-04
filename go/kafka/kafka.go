package kafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const EventTypeKey = "event-type"

type Event string

type Topic string

type Handler func(ctx context.Context, msg *kafka.Message) error

type Consumer struct {
	consumer *kafka.Consumer
	routes map[Event]Handler
}

func New(
	consumer *kafka.Consumer,
	routes map[Event]Handler,
) *Consumer {
	return &Consumer{
		consumer: consumer,
		routes: routes,
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
			} else if !ok {
				return err
			}
		}

		go func(msg *kafka.Message) {
			event := Event(lookupHeader(msg.Headers, EventTypeKey))
			handler, ok := h.routes[event]
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
