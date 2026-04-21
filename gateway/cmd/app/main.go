package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zhmlst/assistant/gateway/internal/adapter/conversation"
	v1 "github.com/zhmlst/assistant/gateway/internal/handler/v1"
	"github.com/zhmlst/assistant/go/logger"
	"google.golang.org/grpc"
)

type Config struct {
	Logger           logger.Config `envPrefix:"LOGGER_"`
	ConversationAddr string
}

func run() error {
	config, err := env.ParseAsWithOptions[Config](env.Options{
		UseFieldNameByDefault: true,
	})
	if err != nil {
		return err
	}

	conn, err := grpc.NewClient(config.ConversationAddr)
	if err != nil {
		return fmt.Errorf("grpc new client: %w", err)
	}

	conversationClient := conversation.New(conn)

	handlerv1 := v1.New(conversationClient, nil)

	_ = handlerv1

	lgr := logger.New(&config.Logger)

	lgr.Info("terminated")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
