package conversationv1

import (
	"fmt"

	"github.com/zhmlst/assistant/go/lib"
)

func RoleFromProto(r Role) (lib.Role, error) {
	switch r {
	case Role_ROLE_ASSISTANT:
		return lib.RoleAssistant, nil
	case Role_ROLE_SYSTEM:
		return lib.RoleSystem, nil
	case Role_ROLE_USER:
		return lib.RoleUser, nil
	default:
		return 0, fmt.Errorf("invalid proto role %v", r)
	}
}

func RoleToProto(r lib.Role) (Role, error) {
	switch r {
	case lib.RoleAssistant:
		return Role_ROLE_ASSISTANT, nil
	case lib.RoleSystem:
		return Role_ROLE_SYSTEM, nil
	case lib.RoleUser:
		return Role_ROLE_USER, nil
	default:
		return Role_ROLE_UNSPECIFIED, fmt.Errorf("invalid domain role %v", r)
	}
}

