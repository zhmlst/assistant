package messages

import (
	"context"
	"fmt"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"github.com/zhmlst/assistant/go/postgres"
	adapter "github.com/zhmlst/assistant/conversation/internal/adapter/postgres"
)

type Messages struct {
	pool postgres.Pool
}

func New(pool postgres.Pool) *Messages {
	return &Messages{pool}
}

func (m *Messages) Store(ctx context.Context, msg *domain.Message) error {
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
	var msg domain.Message

	if err := m.pool.QueryRow(ctx, `
		SELECT id, conversation_id, parent_message_id, role, text, created_at
		FROM messages
		WHERE id = $1
	`,
		adapter.Hash(id),
	).Scan(
		(*adapter.Hash)(&msg.ID),
		&msg.ConversationID,
		(*adapter.Hash)(&msg.ParentID),
		(*adapter.Role)(&msg.Role),
		&msg.Text,
		&msg.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}

	return &msg, nil
}

func (m *Messages) Delete(ctx context.Context, id domain.Hash) error {
	_, err := m.pool.Exec(ctx, `
		DELETE FROM messages
		WHERE id = $1
	`,
		adapter.Hash(id),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}
