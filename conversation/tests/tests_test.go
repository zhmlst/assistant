//go:build e2e

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test(t *testing.T) {
	ctx := t.Context()

	const inport = "50051/tcp"
	req := testcontainers.ContainerRequest{
		Image: "assistant-conversation",
		Env: map[string]string{
			"POSTGRES_HOST": "172.17.0.1",
			"POSTGRES_PORT": "5432",
			"POSTGRES_USER": "postgres",
			"POSTGRES_PASS": "2345",
			"POSTGRES_DB":   "postgres",
		},
		ExposedPorts: []string{inport},
		WaitingFor: wait.ForLog("msg=started").
			WithOccurrence(1).
			WithPollInterval(100 * time.Millisecond),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoErrorf(t, err, "")
	defer container.Terminate(ctx)

	port, err := container.MappedPort(ctx, inport)
	require.NoErrorf(t, err, "")

	conn, err := grpc.NewClient(
		"localhost:"+port.Port(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "create grpc client connection")

	client := conversationv1.NewUserServiceClient(conn)
	_ = client
}
