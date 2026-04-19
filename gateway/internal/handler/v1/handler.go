package v1

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

type Role lib.Role

func (r Role) String() string {
	switch lib.Role(r) {
	case lib.RoleAssistant:
		return "assistant"
	case lib.RoleSystem:
		return "assistant"
	case lib.RoleUser:
		return "user"
	default:
		panic("unknown role")
	}
}

func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Role) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "assistant":
		*r = Role(lib.RoleAssistant)
	case "system":
		*r = Role(lib.RoleSystem)
	case "user":
		*r = Role(lib.RoleUser)
	default:
		return fmt.Errorf("unknown role: %s", s)
	}

	return nil
}

type Message struct {
	ConversationID  uuid.UUID `json:"conversation_id"`
	ID              Hash      `json:"hash"`
	ParentMessageID Hash      `json:"parent_message_id"`
	Role            lib.Role  `json:"role"`
	Text            string    `json:"text"`
}

type Conversation interface {
	CreateMessage(ctx context.Context, msg *Message) error
	ChildMessage(ctx context.Context, conversationID uuid.UUID, parentID lib.Hash) (*Message, error)
}

type Redis interface {
	Subscribe(ctx context.Context, id lib.Hash) (<-chan string, io.Closer, error)
}

type handler struct {
	conversation Conversation
	redis        Redis
}

func New(conversation Conversation, redis Redis) *handler {
	return &handler{
		conversation: conversation,
		redis:        redis,
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
		return
	}

	if err := h.conversation.CreateMessage(r.Context(), &msg); err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher := w.(http.Flusher)

	data, _ := json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	if flusher != nil {
		flusher.Flush()
	}

	if msg.Role != lib.RoleUser {
		return
	}

	ch, closer, err := h.redis.Subscribe(r.Context(), lib.Hash(msg.ID))
	if err != nil {
		h.handleError(w, fmt.Errorf("redis subscribe: %w", err))
		return
	}
	defer closer.Close()

	for chunk := range ch {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		if flusher != nil {
			flusher.Flush()
		}
	}

	reply, err := h.conversation.ChildMessage(r.Context(), msg.ConversationID, lib.Hash(msg.ID))
	if err != nil {
		return
	}

	replyData, _ := json.Marshal(reply)
	fmt.Fprintf(w, "data: %s\n\n", string(replyData))
	if flusher != nil {
		flusher.Flush()
	}
}

