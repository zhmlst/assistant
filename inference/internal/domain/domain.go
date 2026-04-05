package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/go/lib"
)

type Message struct {
	ID lib.Hash
	ConversationID uuid.UUID
	Role lib.Role
	Text string
	CreatedAt time.Time
}
