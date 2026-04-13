package handler

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"github.com/zhmlst/assistant/go/lib"
	"github.com/zhmlst/assistant/inference/internal/domain"
	"google.golang.org/protobuf/proto"
)

type Service interface {
	Reply(ctx context.Context, msg *domain.Message) error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) MessageCreated(ctx context.Context, msg *kafka.Message) error {
	var protomsg conversationv1.Message
	if err := proto.Unmarshal(msg.Value, &protomsg); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	cnvID, err := uuid.FromBytes(protomsg.ConversationId)
	if err != nil {
		return fmt.Errorf("conversation id from bytes: %w", err)
	}

	id, err := lib.HashFromBytes(protomsg.Id)
	if err != nil {
		return fmt.Errorf("message id from bytes: %w", err)
	}

	role, err := conversationv1.RoleFromProto(protomsg.Role)
	if err != nil {
		return fmt.Errorf("role from proto: %w", err)
	}

	return h.service.Reply(ctx, &domain.Message{
		ID:             id,
		ConversationID: cnvID,
		Role:           role,
		Text:           protomsg.Text,
		CreatedAt:      protomsg.CreateTime.AsTime(),
	})
}
