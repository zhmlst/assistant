package lib

import "fmt"

type Role uint8

const (
	_ Role = iota
	RoleAssistant
	RoleSystem
	RoleUser
)

type Hash [32]byte

var NilHash Hash

func HashFromBytes(b []byte) (Hash, error) {
	if len(b) < len(NilHash) {
		return NilHash, fmt.Errorf("invalid hash length %d", len(b))
	}
	return Hash(b), nil
}

