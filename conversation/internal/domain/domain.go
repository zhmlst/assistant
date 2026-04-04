package domain

import (
	"context"
	"crypto/sha256"
	"hash"
	"sync"
	"time"

	"github.com/zhmlst/assistant/go/lib"
	"github.com/google/uuid"
)

type UserFieldMask uint64

const (
	UserFieldUsername UserFieldMask = 1 << iota
)

type User struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func NewUser(username string) (*User, error) {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type Preferences struct {
}

func DefaultPreferences() *Preferences {
	return &Preferences{}
}

type ConversationFieldMask uint64

const (
	ConversationFieldTitle ConversationFieldMask = 1 << iota
)

type Conversation struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Preferences Preferences
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewConversation(
	userID uuid.UUID,
	title string,
) (*Conversation, error) {
	now := time.Now()
	return &Conversation{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Preferences: *DefaultPreferences(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

type Role uint8

const (
	_ Role = iota
	RoleAssistant
	RoleSystem
	RoleUser
)
type Message struct {
	ID             lib.Hash
	ParentID       lib.Hash
	ConversationID uuid.UUID
	Text           string
	Role           Role
	CreatedAt      time.Time
}

var hashes = sync.Pool{
	New: func() any { return sha256.New() },
}

func NewMessage(
	parentID lib.Hash,
	conversationID uuid.UUID,
	text string,
	role Role,
) (*Message, error) {
	m := Message{
		ParentID:       parentID,
		ConversationID: conversationID,
		Text:           text,
		Role:           role,
		CreatedAt:      time.Now(),
	}

	h := hashes.Get().(hash.Hash)
	defer func() { h.Reset(); hashes.Put(h) }()

	h.Write(parentID[:])
	h.Write([]byte(text))
	h.Write([]byte{byte(role)})

	_ = h.Sum(m.ID[:0])

	return &m, nil
}

type Transactor interface {
	Wrap(context.Context, func(context.Context) error) error
}

type UserIDProvider interface {
	UserID(context.Context) (uuid.UUID, error)
}

type ListParameters struct {
	PageSize  int
	PageToken []byte
	Filter    string
	OrderBy   string
}
