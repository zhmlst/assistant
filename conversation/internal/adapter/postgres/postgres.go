package postgres

import (
	"database/sql/driver"
	"fmt"
	"github.com/zhmlst/assistant/conversation/internal/domain"
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
	if src == nil {
		*h = Hash(domain.NilHash)
		return nil
	}
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
	if h == Hash(domain.NilHash) {
		return nil, nil
	}
	return h[:], nil
}
