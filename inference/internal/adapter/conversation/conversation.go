package conversation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/google/uuid"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"github.com/zhmlst/assistant/go/lib"
	"github.com/zhmlst/assistant/inference/internal/domain"
)

type conversation struct {
	client conversationv1.MessageServiceClient
}

func New(client conversationv1.MessageServiceClient) *conversation {
	return &conversation{client}
}

func (c *conversation) History(
	ctx context.Context,
	conversationID uuid.UUID,
	anchorMessageID lib.Hash,
) iter.Seq2[*domain.Message, error] {
	return func(yield func(*domain.Message, error) bool) {
		stream, err := c.client.GetHistory(ctx, &conversationv1.GetHistoryRequest{
			ConversationId:  conversationID[:],
			AnchorMessageId: anchorMessageID[:],
		})
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			yield(nil, fmt.Errorf("get history stream: %w", err))
			return
		}

		for {
			protomsg, err := stream.Recv()
			if err != nil {
				yield(nil, fmt.Errorf("receive message: %w", err))
				return
			}

			msgID, err := lib.HashFromBytes(protomsg.Id)
			if err != nil {
				yield(nil, fmt.Errorf("message id from bytes: %w", err))
				return
			}

			cnvID, err := uuid.FromBytes(protomsg.ConversationId)
			if err != nil {
				yield(nil, fmt.Errorf("conversation id from bytes: %w", err))
				return
			}

			role, err := conversationv1.RoleFromProto(protomsg.Role)
			if err != nil {
				yield(nil, fmt.Errorf("role from proto: %w", err))
				return
			}

			if !yield(&domain.Message{
				ID:             msgID,
				ConversationID: cnvID,
				Role:           role,
				Text:           protomsg.Text,
				CreatedAt:      protomsg.CreateTime.AsTime(),
			}, nil) {
				return
			}
		}
	}
}

func (c *conversation) Prompt(ctx context.Context, conversationID uuid.UUID) (string, error) {
	return "", nil
}

func (c *conversation) CreateMessage(ctx context.Context, conversationID uuid.UUID, text string) error {
	return nil
}
