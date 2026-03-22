package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
)

type Conversations interface {
	Store(ctx context.Context, cnv *domain.Conversation) error
}

type Messages interface {
	Store(ctx context.Context, msg *domain.Message) error
}

type UserIDProvider interface {
	UserID(context.Context) (uuid.UUID, bool)
}

type Service struct {
	messages       Messages
	conversations  Conversations
	userIDProvider UserIDProvider
}

func New(
	messages Messages,
	conversations Conversations,
) *Service {
	return &Service{
		messages:      messages,
		conversations: conversations,
	}
}

func (s *Service) CreateMessage(
	ctx context.Context,
	parentID domain.Hash,
	conversationID uuid.UUID,
	role domain.Role,
	text string,
) (*domain.Message, error) {
	if parentID == domain.NilHash {
		userID, ok := s.userIDProvider.UserID(ctx)
		if !ok {
			return nil, fmt.Errorf("cannot create conversation for unknown user")
		}

		cnv, err := domain.NewConversation(userID, text[:min(16, len(text))])
		if err != nil {
			return nil, err
		}

		if err := s.conversations.Store(ctx, cnv); err != nil {
			return nil, fmt.Errorf("store new conversation: %w", err)
		}

		conversationID = cnv.ID
	}

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
