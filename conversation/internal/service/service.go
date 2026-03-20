package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"fmt"
)

type Messages interface{
	Store(ctx context.Context, msg *domain.Message) (error)
}

type Service struct {
	messages Messages
}

func New(messages Messages) *Service {
	return &Service{
		messages: messages,
	}
}

func (s *Service) CreateMessage(
	ctx context.Context,
	parentID domain.Hash,
	conversationID uuid.UUID,
	role domain.Role,
	text string,
) (*domain.Message, error) {
	msg, err := domain.NewMessage(parentID, conversationID, text, role)
	if err != nil {
		return nil, err
	}

	err = s.messages.Store(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("store new message: %w", err)
	}

	return msg, nil
}
