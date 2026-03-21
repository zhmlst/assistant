package postgres

import (
	"github.com/zhmlst/assistant/go/postgres"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	"context"
	"fmt"
	"database/sql/driver"
)

type Role domain.Role

func (r *Role) Scan(src any) error {
	val, ok := src.(int64)
	if !ok {
		return fmt.Errorf("cannot scan %T into domain.Role, expected int64", val)
	}
	*r = Role(val)
	return nil
}

func (r Role) Value() (driver.Value, error) {
	return int64(r), nil
}

type Hash domain.Hash

func (h *Hash) Scan(src any) error {
	val, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into domain.Hash, expected []byte", val)
	}
	if len(val) < len(h) {
		return fmt.Errorf("cannot scan slice with len %d into domain.Hash array with len 32", len(val))
	}
	*h = Hash(val)
	return nil
}

func (h Hash) Value() (driver.Value, error) {
	return h[:], nil
}

type Messages struct {
	pool postgres.Pool
}

func New(pool postgres.Pool) *Messages {
	return &Messages{}
}

func (m *Messages) Store(ctx context.Context, msg *domain.Message) (error) {
	return nil
}
