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
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	storage Storage
}

func New(storage Storage) *service {
	return &service{
		storage: storage,
	}
}

func (s *service) ByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	cnv, err := s.storage.ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("restore conversation by id: %w", err)
	}
	return cnv, nil
}
