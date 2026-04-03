//go:build e2e

package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestConversations(t *testing.T) {
	ctx := t.Context()

	conn, err := grpc.NewClient(
		"127.0.0.1:"+port.Port(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "")

	usersClient := conversationv1.NewUserServiceClient(conn)
	usr, err := usersClient.CreateUser(ctx, &conversationv1.CreateUserRequest{Username: "conversations:john doe"})
	require.NoErrorf(t, err, "")

	client := conversationv1.NewConversationServiceClient(conn)
	messagesClient := conversationv1.NewMessageServiceClient(conn)

	t.Run("create", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", uuid.UUID(usr.Id).String()))

		msg, err := messagesClient.CreateMessage(ctx, &conversationv1.CreateMessageRequest{
			Role: conversationv1.Role_ROLE_SYSTEM,
			Text: "conversations:prompt",
		})
		require.NoErrorf(t, err, "")

		cnv, err := client.GetConversation(ctx, &conversationv1.GetConversationRequest{Id: msg.ConversationId})
		require.NoErrorf(t, err, "")

		assert.Equal(t, usr.Id, cnv.UserId)
	})
}
