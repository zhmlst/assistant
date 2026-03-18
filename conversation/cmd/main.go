package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/zhmlst/assistant/go/logger"
	"github.com/zhmlst/assistant/go/postgres"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

type Config struct {
	Postgres postgres.Config `envPrefix:"POSTGRES_"`
	Logger   logger.Config   `envPrefix:"LOGGER_"`
}

func run() error {
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
	_ = pgpool

	lgr := logger.New(&cfg.Logger)

	lgr.Info("started")
	<-ctx.Done()
	lgr.Info("terminated")

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
