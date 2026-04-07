package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	kfkconversation "github.com/zhmlst/assistant/conversation/pkg/kafka"
	kfk "github.com/zhmlst/assistant/go/kafka"
	"github.com/zhmlst/assistant/inference/internal/adapter/conversation"
	"github.com/zhmlst/assistant/inference/internal/handler"
	"github.com/zhmlst/assistant/inference/internal/service"
	"google.golang.org/grpc"
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

	conversationClientConn, err := grpc.NewClient("127.0.0.1:50050")

	messageServiceClient := conversationv1.NewMessageServiceClient(conversationClientConn)

	cnv := conversation.New(messageServiceClient)

	service := service.New(nil, cnv, nil, nil)

	handler := handler.New(service)

	consumer := kfk.New(c, map[string]kfk.Handler{
		kfkconversation.EventMessageCreated: handler.MessageCreated,
	})

	if err := consumer.Consume(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}
