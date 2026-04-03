//go:build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestUsers(t *testing.T) {
	const domain = "user:"
	ctx := t.Context()

	conn, err := grpc.NewClient(
		"127.0.0.1:"+port.Port(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "create grpc client connection")

	client := conversationv1.NewUserServiceClient(conn)

	t.Run("create user", func(t *testing.T) {
		user1, err := client.CreateUser(ctx, &conversationv1.CreateUserRequest{
			Username: domain + "jonh doe",
		})
		require.NoErrorf(t, err, "")

		user2, err := client.GetUser(ctx, &conversationv1.GetUserRequest{
			Id: user1.Id,
		})
		require.NoErrorf(t, err, "")

		assert.Equal(t, user1, user2)
	})

	t.Run("update user", func(t *testing.T) {
		user1, err := client.CreateUser(ctx, &conversationv1.CreateUserRequest{
			Username: domain + "john doe",
		})
		require.NoErrorf(t, err, "")

		_, err = client.UpdateUser(ctx, &conversationv1.UpdateUserRequest{
			User: &conversationv1.User{
				Id:       user1.Id,
				Username: domain + "lolkek",
			},
			FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"username"}},
		})
		require.NoErrorf(t, err, "")

		user2, err := client.GetUser(ctx, &conversationv1.GetUserRequest{
			Id: user1.Id,
		})
		require.NoErrorf(t, err, "")

		assert.Equal(t, domain+"lolkek", user2.Username)
	})
}
