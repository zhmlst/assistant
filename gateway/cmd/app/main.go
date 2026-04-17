package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/zhmlst/assistant/go/logger"
)

type Config struct {
	Logger logger.Config `envPrefix:"LOGGER_"`
}

func run() error {
	config, err := env.ParseAsWithOptions[Config](env.Options{
		UseFieldNameByDefault: true,
	})
	if err != nil {
		return err
	}

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
