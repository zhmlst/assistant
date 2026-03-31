package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	ByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.User, error)

	CreateUser(
		ctx context.Context,
	) (*domain.User, error)

	DeleteUser(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type handler struct {
	conversationv1.UnimplementedUserServiceServer
	service service
}

func New(service service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetUser(ctx context.Context, req *conversationv1.GetUserRequest) (*conversationv1.User, error) {
	usrID, err := uuid.FromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	usr, err := h.service.ByID(ctx, usrID)
	if err != nil {
		return nil, err
	}

	return &conversationv1.User{
		Id:         usr.ID[:],
		CreateTime: timestamppb.New(usr.CreatedAt),
		UpdateTime: timestamppb.New(usr.UpdatedAt),
		DeleteTime: timestamppb.New(usr.DeletedAt),
	}, nil
}

func (h *handler) CreateUser(ctx context.Context, req *conversationv1.CreateUserRequest) (*conversationv1.User, error) {
	usr, err := h.service.CreateUser(ctx)
	if err != nil {
		return nil, err
	}

	return &conversationv1.User{
		Id:         usr.ID[:],
		CreateTime: timestamppb.New(usr.CreatedAt),
		UpdateTime: timestamppb.New(usr.UpdatedAt),
		DeleteTime: timestamppb.New(usr.DeletedAt),
	}, nil
}

func (h *handler) UpdateUser(ctx context.Context, req *conversationv1.UpdateUserRequest) (*conversationv1.User, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateUser not implemented")
}

func (h *handler) DeleteUser(ctx context.Context, req *conversationv1.DeleteUserRequest) (*emptypb.Empty, error) {
	usrID, err := uuid.ParseBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	if err := h.service.DeleteUser(ctx, usrID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
