package message

import (
	"context"
	"fmt"
	"iter"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/zhmlst/assistant/conversation/internal/service/conversation"
	"github.com/zhmlst/assistant/go/lib"
)

type EventProvider interface {
	CreateMessage(ctx context.Context, msg *domain.Message)
}

type Storage interface {
	History(ctx context.Context, conversationID uuid.UUID, anchorID lib.Hash) iter.Seq2[*domain.Message, error]
	Select(ctx context.Context, parentID, variantID lib.Hash, conversationID uuid.UUID) error
	ByID(ctx context.Context, conversationID uuid.UUID, id lib.Hash) (*domain.Message, error)
	Store(ctx context.Context, msg *domain.Message) error
	Delete(ctx context.Context, id lib.Hash) error
}

type service struct {
	transactor          domain.Transactor
	storage             Storage
	conversationStorage conversation.Storage
	userIDProvider      domain.UserIDProvider
}

func New(
	transactor domain.Transactor,
	storage Storage,
	userIDProvider domain.UserIDProvider,
	conversationStorage conversation.Storage,
) *service {
	return &service{
		transactor:          transactor,
		storage:             storage,
		userIDProvider:      userIDProvider,
		conversationStorage: conversationStorage,
	}
}

func (s *service) CreateMessage(
	ctx context.Context,
	parentID lib.Hash,
	conversationID uuid.UUID,
	role lib.Role,
	text string,
) (*domain.Message, error) {
	if parentID == lib.NilHash {
		userID, err := s.userIDProvider.UserID(ctx)
		if err != nil {
			return nil, fmt.Errorf("get user id: %w", err)
		}

		cnv, err := domain.NewConversation(userID, text[:min(16, len(text))])
		if err != nil {
			return nil, err
		}

		if err := s.conversationStorage.Store(ctx, cnv); err != nil {
			return nil, fmt.Errorf("store new conversation: %w", err)
		}

		conversationID = cnv.ID
	}

	msg, err := domain.NewMessage(parentID, conversationID, text, role)
	if err != nil {
		return nil, err
	}

	err = s.storage.Store(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("store new message: %w", err)
	}

	return msg, nil
}

func (s *service) DeleteMessage(
	ctx context.Context,
	id lib.Hash,
	conversationID uuid.UUID,
) error {
	msg, err := s.storage.ByID(ctx, conversationID, id)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	return s.transactor.Wrap(ctx, func(ctx context.Context) error {
		if msg.ParentID == lib.NilHash {
			if err := s.conversationStorage.Delete(ctx, msg.ConversationID); err != nil {
				return fmt.Errorf("delete conversation: %w", err)
			}
		}

		if err := s.storage.Delete(ctx, msg.ID); err != nil {
			return fmt.Errorf("delete message: %w", err)
		}

		return nil
	})
}

func (s *service) GetMessage(ctx context.Context, conversationID uuid.UUID, id lib.Hash) (*domain.Message, error) {
	msg, err := s.storage.ByID(ctx, conversationID, id)
	if err != nil {
		return nil, fmt.Errorf("restore message by id: %w", err)
	}
	return msg, nil
}

func (s *service) SelectVariant(ctx context.Context, parentID, variantID lib.Hash, conversationID uuid.UUID) error {
	if err := s.storage.Select(ctx, parentID, variantID, conversationID); err != nil {
		return fmt.Errorf("select variant: %w", err)
	}
	return nil
}

func (s *service) History(ctx context.Context, conversationID uuid.UUID, anchorID lib.Hash) iter.Seq2[*domain.Message, error] {
	return s.storage.History(ctx, conversationID, anchorID)
}
