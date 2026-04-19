package conversation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	v1 "github.com/zhmlst/assistant/gateway/internal/handler/v1"
	"github.com/zhmlst/assistant/go/lib"
	"google.golang.org/grpc"
)

type client struct {
	messages conversationv1.MessageServiceClient
}

func New(conn *grpc.ClientConn) *client {
	return &client{
		messages: conversationv1.NewMessageServiceClient(conn),
	}
}

func (c *client) CreateMessage(ctx context.Context, msg *v1.Message) error {
	role, err := conversationv1.RoleToProto(msg.Role)
	if err != nil {
		return fmt.Errorf("role to proto: %w", err)
	}

	resmsg, err := c.messages.CreateMessage(ctx, &conversationv1.CreateMessageRequest{
		ParentId: msg.ID[:],
		ConversationId: msg.ConversationID[:],
		Role: role,
		Text: msg.Text,
	})
	if err != nil {
		return err
	}

	id, err := lib.HashFromBytes(resmsg.Id)
	if err != nil {
		return fmt.Errorf("hash from bytes: %w", err)
	}

	msg.ID = v1.Hash(id)
	return nil
}

func (c *client) ChildMessage(ctx context.Context, conversationID uuid.UUID, parentID lib.Hash) (*v1.Message, error) {
	return nil, nil
}
