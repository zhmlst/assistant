package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5/stdlib"
	kafkapub "github.com/zhmlst/assistant/conversation/internal/adapter/kafka"
	conversationstorage "github.com/zhmlst/assistant/conversation/internal/adapter/postgres/conversations"
	messagestorage "github.com/zhmlst/assistant/conversation/internal/adapter/postgres/messages"
	"github.com/zhmlst/assistant/conversation/internal/adapter/postgres/migrations"
	userstorage "github.com/zhmlst/assistant/conversation/internal/adapter/postgres/users"
	v1 "github.com/zhmlst/assistant/conversation/internal/handler/v1"
	conversationhandler "github.com/zhmlst/assistant/conversation/internal/handler/v1/conversation"
	messagehandler "github.com/zhmlst/assistant/conversation/internal/handler/v1/message"
	userhandler "github.com/zhmlst/assistant/conversation/internal/handler/v1/user"
	conversationservice "github.com/zhmlst/assistant/conversation/internal/service/conversation"
	messageservice "github.com/zhmlst/assistant/conversation/internal/service/message"
	userservice "github.com/zhmlst/assistant/conversation/internal/service/user"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	gokafka "github.com/zhmlst/assistant/go/kafka"
	"github.com/zhmlst/assistant/go/logger"
	"github.com/zhmlst/assistant/go/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Postgres postgres.Config `envPrefix:"POSTGRES_"`
	Kafka    gokafka.Config  `envPrefix:"KAFKA_"`
	Logger   logger.Config   `envPrefix:"LOGGER_"`
	GRPC     struct {
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

	if err := migrations.Up(stdlib.OpenDBFromPool(pgpool.Pool), "postgres"); err != nil {
		return fmt.Errorf("migate db: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Kafka.Map())
	if err != nil {
		return fmt.Errorf("kafka new producer: %w", err)
	}

	lgr := logger.New(&cfg.Logger)

	userStorage := userstorage.New(pgpool)
	conversationStorage := conversationstorage.New(pgpool)
	messageStorage := messagestorage.New(pgpool)

	userService := userservice.New(pgpool, userStorage)
	conversationService := conversationservice.New(pgpool, conversationStorage)
	eventPublisher := kafkapub.New(producer)
	messageService := messageservice.New(pgpool, messageStorage, v1.UserIDProvider{}, conversationStorage, eventPublisher)

	userHandler := userhandler.New(userService)
	conversationHandler := conversationhandler.New(conversationService)
	messageHandler := messagehandler.New(messageService)

	healthHandler := health.NewServer()

	server := grpc.NewServer()
	conversationv1.RegisterUserServiceServer(server, userHandler)
	conversationv1.RegisterConversationServiceServer(server, conversationHandler)
	conversationv1.RegisterMessageServiceServer(server, messageHandler)
	grpc_health_v1.RegisterHealthServer(server, healthHandler)
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

	go func() {
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
