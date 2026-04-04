package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

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
