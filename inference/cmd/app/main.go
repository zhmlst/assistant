package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	gokafka "github.com/zhmlst/assistant/go/kafka"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	kfkconversation "github.com/zhmlst/assistant/conversation/pkg/kafka"
	"github.com/zhmlst/assistant/inference/internal/adapter/conversation"
	"github.com/zhmlst/assistant/inference/internal/handler"
	"github.com/zhmlst/assistant/inference/internal/service"
	"google.golang.org/grpc"
)

type Config struct {
	Kafka gokafka.Config
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	kafkaConfig, err := gokafka.NewConfig()
	if err != nil {
		return
	}

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	if err := c.SubscribeTopics([]string{kfkconversation.TopicMessage}, nil); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	conversationClientConn, err := grpc.NewClient("127.0.0.1:50050")
	if err != nil {
		return
	}

	messageServiceClient := conversationv1.NewMessageServiceClient(conversationClientConn)

	cnv := conversation.New(messageServiceClient)

	service := service.New(nil, cnv, nil, nil)

	handler := handler.New(service)

	consumer := gokafka.New(c, map[string]gokafka.Handler{
		kfkconversation.EventMessageCreated: handler.MessageCreated,
	})

	if err := consumer.Consume(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}
