package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
)

type Transactor interface {
	Wrap(context.Context, func(context.Context) error) error
}

type Conversations interface {
	Store(ctx context.Context, cnv *domain.Conversation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Messages interface {
	ByID(ctx context.Context, id domain.Hash) (*domain.Message, error)
	Store(ctx context.Context, msg *domain.Message) error
	Delete(ctx context.Context, id domain.Hash) error
}

type UserIDProvider interface {
	UserID(context.Context) (uuid.UUID, bool)
}

type Service struct {
	transactor     Transactor
	messages       Messages
	conversations  Conversations
	userIDProvider UserIDProvider
}

func New(
	transactor Transactor,
	messages Messages,
	conversations Conversations,
	userIDProvider UserIDProvider,
) *Service {
	return &Service{
		transactor:     transactor,
		messages:       messages,
		conversations:  conversations,
		userIDProvider: userIDProvider,
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
			return nil, domain.ErrInvalidInput.New("cannot create conversation for unknown user")
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

func (s *Service) DeleteMessage(
	ctx context.Context,
	id domain.Hash,
	conversationID uuid.UUID,
) error {
	msg, err := s.messages.ByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	return s.transactor.Wrap(ctx, func(ctx context.Context) error {
		if msg.ParentID == domain.NilHash {
			if err := s.conversations.Delete(ctx, msg.ConversationID); err != nil {
				return fmt.Errorf("delete conversation: %w", err)
			}
		}

		if err := s.messages.Delete(ctx, msg.ID); err != nil {
			return fmt.Errorf("delete message: %w", err)
		}

		return nil
	})
}
