package conversations

import (
	"context"
	"fmt"
	"github.com/zhmlst/assistant/go/postgres"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/google/uuid"
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
	return nil
}
