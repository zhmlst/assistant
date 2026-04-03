//go:build e2e

package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var port nat.Port

func TestMain(t *testing.M) {
	ctx := context.Background()

	const inport = "50051/tcp"
	req := testcontainers.ContainerRequest{
		Image: "assistant-conversation",
		Env: map[string]string{
			"GRPC_ADDR":     "0.0.0.0:50051",
			"POSTGRES_HOST": "172.17.0.1",
			"POSTGRES_PORT": "5432",
			"POSTGRES_USER": "postgres",
			"POSTGRES_PASS": "2345",
			"POSTGRES_DB":   "postgres",
		},
		ExposedPorts: []string{inport},
		WaitingFor:   wait.ForListeningPort(inport),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	port, err = container.MappedPort(ctx, inport)
	if err != nil {
		log.Fatalln(err.Error())
	}

	code := t.Run()
	container.Terminate(ctx)
	os.Exit(code)
}
