package messages

import (
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/zhmlst/assistant/go/postgres"
	"context"
	"fmt"
)

type Messages struct {
	pool postgres.Pool
}

func New(pool postgres.Pool) *Messages {
	return &Messages{pool}
}

func (m *Messages) Store(ctx context.Context, msg *domain.Message) (error) {
	if err := m.pool.QueryRow(ctx, `
		INSERT INTO messages (id, conversation_id, parent_message_id, role, text)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at;
	`).Scan(
		&msg.CreatedAt,
	); err != nil {
		return fmt.Errorf("query row: %w", err)
	}

	return nil
}

func (m *Messages) ByID(ctx context.Context, id domain.Hash) (*domain.Message, error) {
	return nil, nil
}

func (m *Messages) Delete(ctx context.Context, id domain.Hash) error {
	return nil
}
