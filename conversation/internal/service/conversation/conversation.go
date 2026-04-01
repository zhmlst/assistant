package conversation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
)

type Storage interface {
	ByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	Store(ctx context.Context, cnv *domain.Conversation) error
	Update(ctx context.Context, cnv *domain.Conversation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	transactor domain.Transactor
	storage    Storage
}

func New(transactor domain.Transactor, storage Storage) *service {
	return &service{
		transactor: transactor,
		storage:    storage,
	}
}

func (s *service) ByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	cnv, err := s.storage.ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("restore conversation by id: %w", err)
	}
	return cnv, nil
}

func (s *service) Update(ctx context.Context, upd *domain.Conversation, mask domain.ConversationFieldMask) error {
	return s.transactor.Wrap(ctx, func(ctx context.Context) error {
		cnv, err := s.storage.ByID(ctx, upd.ID)
		if err != nil {
			return fmt.Errorf("storage get by id: %w", err)
		}

		if mask&domain.ConversationFieldTitle == domain.ConversationFieldTitle {
			cnv.Title = upd.Title
		}

		if err := s.storage.Update(ctx, cnv); err != nil {
			return fmt.Errorf("storage update: %w", err)
		}

		return nil
	})
}
