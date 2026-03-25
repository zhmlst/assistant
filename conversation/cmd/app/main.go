package main

import (
	"context"
	"errors"
	"fmt"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	handlerv1 "github.com/zhmlst/assistant/conversation/internal/handler/v1"
	"github.com/zhmlst/assistant/conversation/internal/service"
	"github.com/caarlos0/env/v11"
	"github.com/zhmlst/assistant/go/logger"
	"github.com/zhmlst/assistant/go/postgres"
	"net/url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
	"net"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"log/slog"
)

type Config struct {
	Postgres postgres.Config `envPrefix:"POSTGRES_"`
	Logger   logger.Config   `envPrefix:"LOGGER_"`
	GRPC struct {
		Addr string
	} `envPrefix:"GRPC_"`
}

func run() (cause error) {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	cfg, err := env.ParseAsWithOptions[Config](env.Options{
		UseFieldNameByDefault: true,
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeFor[url.Values](): func(s string) (any, error) {
				vals, err := url.ParseQuery(s)
				if err != nil {
					return nil, err
				}
				return vals, nil
			},
		},
	})
	if err != nil {
		return fmt.Errorf("new config: %w", err)
	}

	pgpool, err := func() (postgres.Pool, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return postgres.New(ctx, &cfg.Postgres)
	}()
	if err != nil {
		return fmt.Errorf("postgres new: %w", err)
	}

	lgr := logger.New(&cfg.Logger)

	service := service.New(pgpool, nil, nil, nil)

	server := grpc.NewServer()
	handlerV1 := handlerv1.New(service)
	conversationv1.RegisterMessageServiceServer(server, handlerV1)
	conversationv1.RegisterConversationServiceServer(server, handlerV1)
	reflection.Register(server)

	addr, err := net.ResolveTCPAddr("tcp", cfg.GRPC.Addr)
	if err != nil {
		return fmt.Errorf("resolve tcp addr: %w", err)
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}

	lgr.Info("listen grpc", slog.String("addr", lis.Addr().String()))

	go func () {
		if err := server.Serve(lis); errors.Is(err, grpc.ErrServerStopped) {
			cause = err
		}
	}()
	defer server.GracefulStop()

	lgr.Info("started")
	<-ctx.Done()
	lgr.Info("terminated")

	return
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
