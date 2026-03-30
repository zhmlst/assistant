package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc/metadata"
)

func RoleFromProto(r conversationv1.Role) (domain.Role, error) {
	switch r {
	case conversationv1.Role_ROLE_ASSISTANT:
		return domain.RoleAssistant, nil
	case conversationv1.Role_ROLE_SYSTEM:
		return domain.RoleSystem, nil
	case conversationv1.Role_ROLE_USER:
		return domain.RoleUser, nil
	default:
		return 0, fmt.Errorf("invalid proto role %v", r)
	}
}

func RoleToProto(r domain.Role) (conversationv1.Role, error) {
	switch r {
	case domain.RoleAssistant:
		return conversationv1.Role_ROLE_ASSISTANT, nil
	case domain.RoleSystem:
		return conversationv1.Role_ROLE_SYSTEM, nil
	case domain.RoleUser:
		return conversationv1.Role_ROLE_USER, nil
	default:
		return conversationv1.Role_ROLE_UNSPECIFIED, fmt.Errorf("invalid domain role %v", r)
	}
}

var (
	ErrMetadataMissing = errors.New("metadata not found in context")
	ErrUserIDMissing   = errors.New("x-user-id header is missing")
	ErrInvalidUUID     = errors.New("x-user-id is not a valid uuid")
)

type UserIDProvider struct{}

func (UserIDProvider) UserID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, ErrMetadataMissing
	}

	xuserid := md.Get("x-user-id")
	if len(xuserid) == 0 || xuserid[0] == "" {
		return uuid.Nil, ErrUserIDMissing
	}

	userID, err := uuid.Parse(xuserid[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrInvalidUUID, err)
	}

	return userID, nil
}
