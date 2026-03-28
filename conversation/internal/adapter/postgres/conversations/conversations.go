package conversations

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/zhmlst/assistant/go/postgres"
)

type Conversations struct {
	pool postgres.Pool
}

func New(pool postgres.Pool) *Conversations {
	return &Conversations{pool}
}

func (c *Conversations) Store(ctx context.Context, cnv *domain.Conversation) error {
	if err := c.pool.QueryRow(ctx, `
		INSERT INTO conversations (id, user_id, title)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`,
		cnv.ID,
		cnv.UserID,
		cnv.Title,
	).Scan(
		&cnv.CreatedAt,
		&cnv.UpdatedAt,
	); err != nil {
		return fmt.Errorf("query row: %w", err)
	}

	return nil
}

func (c *Conversations) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := c.pool.Exec(ctx, `
		DELETE FROM conversations
		WHERE id = $1
	`,
		id,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (c *Conversations) ByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	cnv := domain.Conversation{ID: id}

	if err := c.pool.QueryRow(ctx, `
		SELECT user_id, title, created_at, updated_at FROM conversations
		WHERE id = $1
	`,
		id,
	).Scan(
		&cnv.UserID,
		&cnv.Title,
		&cnv.CreatedAt,
		&cnv.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}

	return &cnv, nil
}
