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
	Update(ctx context.Context, usr *domain.User) error
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

func (s *service) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	usr, err := s.storage.ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}

	return usr, nil
}

func (s *service) UpdateUser(ctx context.Context, upd *domain.User, mask domain.UserFieldMask) error {
	return s.transactor.Wrap(ctx, func(ctx context.Context) error {
		usr, err := s.storage.ByID(ctx, upd.ID)
		if err != nil {
			return fmt.Errorf("get user by id: %w", err)
		}

		if mask&domain.UserFieldUsername == domain.UserFieldUsername {
			usr.Username = upd.Username
		}

		if err := s.storage.Update(ctx, usr); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	})
}

func (s *service) CreateUser(ctx context.Context, username string) (*domain.User, error) {
	usr, err := domain.NewUser(username)
	if err != nil {
		return nil, err
	}

	if err = s.storage.Store(ctx, usr); err != nil {
		return nil, fmt.Errorf("store user: %w", err)
	}
	return usr, nil
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
