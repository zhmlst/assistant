package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
)

type Storage interface {
	ByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Store(ctx context.Context, usr *domain.User) error
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

func (s *service) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	usr, err := s.storage.ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}

	return usr, nil
}

func (s *service) CreateUser(ctx context.Context) (*domain.User, error) {
	usr := domain.User{
		ID: uuid.New(),
	}
	err := s.storage.Store(ctx, &usr)
	if err != nil {
		return nil, fmt.Errorf("store user: %w", err)
	}
	return &usr, nil
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
