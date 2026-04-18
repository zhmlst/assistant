package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/go/lib"
)

type Message struct {
	ConversationID  uuid.UUID `json:"conversation_id"`
	ParentMessageID lib.Hash  `json:"parent_message_id"`
	Role            lib.Role  `json:"role"`
	Text            string    `json:"text"`
}

type Conversation interface {
	CreateMessage(ctx context.Context, msg *Message) error
}

type handler struct {
	conversation Conversation
}

func New(conversation Conversation) *handler {
	return &handler{
		conversation: conversation,
	}
}

func (h *handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var msg Message
	_ = json.NewDecoder(r.Body).Decode(&msg)
	_ = h.conversation.CreateMessage(r.Context(), &msg)
	_ = json.NewEncoder(w).Encode(&msg)
}
