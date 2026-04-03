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

func (c *Conversations) List(ctx context.Context, params domain.ListParameters) ([]domain.Conversation, []byte, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		LIMIT $1
	`,
		max(params.PageSize, -params.PageSize),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("pool query: %w", err)
	}
	defer rows.Close()

	list := make([]domain.Conversation, 0, max(params.PageSize, -params.PageSize))
	for rows.Next() {
		var cnv domain.Conversation
		if err := rows.Scan(
			&cnv.ID,
			&cnv.UserID,
			&cnv.Title,
			&cnv.CreatedAt,
			&cnv.UpdatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan row: %w", err)
		}
		list = append(list, cnv)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate rows: %w", err)
	}

	return list, nil, nil
}

func (c *Conversations) Update(ctx context.Context, cnv *domain.Conversation) error {
	if err := c.pool.QueryRow(ctx, `
		UPDATE conversations
		SET title = $1
		WHERE id = $2
		RETURNING updated_at
	`,
		cnv.Title,
		cnv.ID,
	).Scan(
		&cnv.UpdatedAt,
	); err != nil {
		return fmt.Errorf("query row scan: %w", err)
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
