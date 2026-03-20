package domain

import (
	"crypto/sha256"
	"github.com/google/uuid"
	"hash"
	"sync"
	"time"
)

type Role uint8

const (
	_ Role = iota
	RoleAssistant
	RoleSystem
	RoleUser
)

type Hash [32]byte

var NilHash Hash

type Message struct {
	ID             Hash
	ParentID       Hash
	ConversationID uuid.UUID
	Text           string
	Role           Role
	CreatedAt      time.Time
}

var hashes = sync.Pool{
	New: func() any { return sha256.New() },
}

func NewMessage(
	parentID Hash,
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
