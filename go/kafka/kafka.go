package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const EventTypeKey = "event-type"

type Handler func(ctx context.Context, msg *kafka.Message) error

func NopHandler(ctx context.Context, msg *kafka.Message) error {
	fmt.Println(msg)
	return nil
}

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
	Base         *kafka.Consumer
	Routes       map[string]Handler
	ErrorHandler func(msg *kafka.Message, err error)
}

func (h *Consumer) Consume(ctx context.Context) (err error) {
	if h.Base == nil {
		h.Base, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": "127.0.0.1:9092",
		})
		if err != nil {
			return err
		}
	}

	if len(h.Routes) == 0 {
		h.Routes = map[string]Handler{"": NopHandler}
	}

	if noph := h.Routes[""]; noph == nil {
		h.Routes[""] = NopHandler
	}

	for {
		msg, err := h.Base.ReadMessage(500 * time.Millisecond)
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
			handler, ok := h.Routes[string(event)]
			if !ok {
				handler = h.Routes[""]
			}

			if err = handler(ctx, msg); h.ErrorHandler != nil && err != nil {
				h.ErrorHandler(msg, err)
			}
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
