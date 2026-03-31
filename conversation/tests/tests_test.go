//go:build e2e

package tests

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	require.NoErrorf(t, err, "")
	defer container.Terminate(ctx)

	defer func() {
		reader, err := container.Logs(ctx)
		if err == nil {
			io.Copy(os.Stdout, reader)
		}
	}()

	port, err := container.MappedPort(ctx, inport)
	require.NoErrorf(t, err, "")

	conn, err := grpc.NewClient(
		"127.0.0.1:"+port.Port(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "create grpc client connection")

	client := conversationv1.NewUserServiceClient(conn)
	user1, err := client.CreateUser(ctx, &conversationv1.CreateUserRequest{})
	require.NoErrorf(t, err, "")

	user2, err := client.GetUser(ctx, &conversationv1.GetUserRequest{
		Id: user1.Id,
	})
	require.NoErrorf(t, err, "")

	assert.Equal(t, user1, user2)
}
