package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/zhmlst/assistant/gateway/internal/adapter/conversation"
	v1 "github.com/zhmlst/assistant/gateway/internal/handler/v1"
	"github.com/zhmlst/assistant/go/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	ShutdownTimeout  time.Duration
	Logger           logger.Config `envPrefix:"LOGGER_"`
	ConversationAddr string
}

func run() error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	config, err := env.ParseAsWithOptions[Config](env.Options{
		UseFieldNameByDefault: true,
	})
	if err != nil {
		return err
	}
	
	lgr := logger.New(&config.Logger)

	conn, err := grpc.NewClient(
		config.ConversationAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("grpc new client: %w", err)
	}

	conversationClient := conversation.New(conn)

	handlerv1 := v1.New(conversationClient, nil)

	_ = handlerv1

	server := http.Server{}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			lgr.Error("serve http", slog.Any("err", err))
			cancel()
		}
	}()

	<-ctx.Done()
	ctx, cancel = context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		lgr.Error("shutdown server", slog.Any("err", err))
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
