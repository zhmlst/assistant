package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kfkconversation "github.com/zhmlst/assistant/conversation/pkg/kafka"
	kfk "github.com/zhmlst/assistant/go/kafka"
	"github.com/zhmlst/assistant/inference/internal/handler"
	"github.com/zhmlst/assistant/inference/internal/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":"127.0.0.1:9092",
		"group.id":"inference",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	if err := c.SubscribeTopics([]string{kfkconversation.TopicMessage}, nil); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	service := service.New()

	handler := handler.New(service)

	consumer := kfk.New(c, map[string]kfk.Handler{
		kfkconversation.EventMessageCreated: handler.MessageCreated,
	})

	if err := consumer.Consume(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}
