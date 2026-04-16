package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	kfkconversation "github.com/zhmlst/assistant/conversation/pkg/kafka"
	gokafka "github.com/zhmlst/assistant/go/kafka"
	"github.com/zhmlst/assistant/go/logger"
	"github.com/zhmlst/assistant/inference/internal/adapter/conversation"
	"github.com/zhmlst/assistant/inference/internal/adapter/llama"
	"github.com/zhmlst/assistant/inference/internal/adapter/redis"
	"github.com/zhmlst/assistant/inference/internal/handler"
	"github.com/zhmlst/assistant/inference/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Logger logger.Config `envPrefix:"LOGGER_"`
	Kafka gokafka.Config `envPrefix:"KAFKA_"`
	GRPC  struct {
		Addr string
	} `envPrefix:"GRPC_"`
	ConversationAddr string
	Service          service.Config `envPrefix:"SERVICE_"`
	LLaMA            llama.Config   `envPrefix:"LLAMA_"`
}

func run() error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	config, err := env.ParseAsWithOptions[Config](env.Options{
		UseFieldNameByDefault: true,
	})
	if err != nil {
		return fmt.Errorf("parse config from env: %w", err)
	}

	lgr := logger.New(&config.Logger)

	c, err := kafka.NewConsumer(config.Kafka.Map())
	if err != nil {
		return fmt.Errorf("kafka new consumer: %w", err)
	}

	if err := c.SubscribeTopics([]string{kfkconversation.TopicMessage}, nil); err != nil {
		return fmt.Errorf("subscribe topics: %w", err)
	}

	conversationClientConn, err := grpc.NewClient(
		config.ConversationAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("conversation new client conn: %w", err)
	}

	messageServiceClient := conversationv1.NewMessageServiceClient(conversationClientConn)

	cnv := conversation.New(messageServiceClient)

	llama := llama.New(&config.LLaMA)

	rds := redis.New()

	service := service.New(&config.Service, cnv, llama, rds)

	handler := handler.New(service)

	consumer := gokafka.Consumer{
		Base: c,
		Routes: map[string]gokafka.Handler{
			kfkconversation.EventMessageCreated: handler.MessageCreated,
		},
	}

	consumer.ErrorHandler = func(msg *kafka.Message, err error) {
		fmt.Println(err.Error())
	}

	errCh := make(chan error, 1)
	go func() {
		if err := consumer.Consume(ctx); err != nil {
			errCh <- err
		}
	}()

	lgr.Info("launched")

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	lgr.Info("terminated")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
