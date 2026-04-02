package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	ByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.User, error)

	CreateUser(
		ctx context.Context,
		username string,
	) (*domain.User, error)

	UpdateUser(
		ctx context.Context,
		usr *domain.User,
		mask domain.UserFieldMask,
	) error

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
		Username:   usr.Username,
		CreateTime: timestamppb.New(usr.CreatedAt),
		UpdateTime: timestamppb.New(usr.UpdatedAt),
		DeleteTime: timestamppb.New(usr.DeletedAt),
	}, nil
}

func (h *handler) CreateUser(ctx context.Context, req *conversationv1.CreateUserRequest) (*conversationv1.User, error) {
	usr, err := h.service.CreateUser(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return &conversationv1.User{
		Id:         usr.ID[:],
		Username:   usr.Username,
		CreateTime: timestamppb.New(usr.CreatedAt),
		UpdateTime: timestamppb.New(usr.UpdatedAt),
		DeleteTime: timestamppb.New(usr.DeletedAt),
	}, nil
}

func fieldMaskFromProto(proto *fieldmaskpb.FieldMask) (domain.UserFieldMask, error) {
	var mask domain.UserFieldMask
	for _, path := range proto.Paths {
		if path == "username" {
			mask |= domain.UserFieldUsername
		}
	}
	return mask, nil
}

func (h *handler) UpdateUser(ctx context.Context, req *conversationv1.UpdateUserRequest) (*conversationv1.User, error) {
	usrID, err := uuid.FromBytes(req.User.Id)
	if err != nil {
		return nil, fmt.Errorf("user id from bytes: %w", err)
	}

	usr := domain.User{
		ID:       usrID,
		Username: req.User.Username,
	}

	mask, err := fieldMaskFromProto(req.FieldMask)
	if err != nil {
		return nil, fmt.Errorf("user field mask from proto: %w", err)
	}

	if err := h.service.UpdateUser(ctx, &usr, mask); err != nil {
		return nil, err
	}

	return &conversationv1.User{
		Id:         usr.ID[:],
		Username:   usr.Username,
		CreateTime: timestamppb.New(usr.CreatedAt),
		UpdateTime: timestamppb.New(usr.UpdatedAt),
	}, nil
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
