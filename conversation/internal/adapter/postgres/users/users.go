package users

import (
	"fmt"
	"time"
	"github.com/zhmlst/assistant/go/postgres"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/google/uuid"
	"context"
)

type Users struct {
	pool postgres.Pool
}

func New(pool postgres.Pool) *Users {
	return &Users{pool}
}

func (u *Users) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	usr := domain.User{ID:id}
	if err := u.pool.QueryRow(ctx, `
		SELECT created_at, updated_at, deleted_at FROM users
		WHERE id = $1
	`,
		id,
	).Scan(
		&usr.CreatedAt,
		&usr.UpdatedAt,
		&usr.DeletedAt,
	); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	return &usr, nil
}

type nullableTime time.Time

func (t *nullableTime) Scan(src any) error {
	if src == nil {
		return nil
	}

	val, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("cannot convert %T into time.Time", src)
	}

	*t = nullableTime(val)
	return nil
}

func (u *Users) Store(ctx context.Context, usr *domain.User) error {
	if err := u.pool.QueryRow(ctx, `
		INSERT INTO users (id)
		VALUES ($1)
		RETURNING created_at, updated_at, deleted_at
	`,
		usr.ID,
	).Scan(
		&usr.CreatedAt,
		&usr.UpdatedAt,
		(*nullableTime)(&usr.DeletedAt),
	); err != nil {
		return  fmt.Errorf("scan row: %w", err)
	}

	return nil
}

func (u *Users) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := u.pool.Exec(ctx, `
		DELETE FROM users
		WHERE id = $1
	`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
