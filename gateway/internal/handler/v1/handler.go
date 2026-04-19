package v1

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/go/lib"
)

type Hash lib.Hash

func (h *Hash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	*h = Hash(decoded)
	return nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h[:]))
}

type Message struct {
	ConversationID  uuid.UUID `json:"conversation_id"`
	ParentMessageID Hash      `json:"parent_message_id"`
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

func (h *handler) handleError(w http.ResponseWriter, err error) {
	type Error struct {
		Error string `json:"error"`
	}

	jerr := Error{
		Error: err.Error(),
	}

	if err := json.NewEncoder(w).Encode(jerr); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func (h *handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.handleError(w, fmt.Errorf("decode request: %w", err))
	}

	if err := h.conversation.CreateMessage(r.Context(), &msg); err != nil {
		h.handleError(w, err)
	}

	if err := json.NewEncoder(w).Encode(&msg); err != nil {
		h.handleError(w, fmt.Errorf("encode response: %w", err))
	}
}
