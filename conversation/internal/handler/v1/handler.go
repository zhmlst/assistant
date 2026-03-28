package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	GetMessage(
		ctx context.Context,
		id domain.Hash,
	) (*domain.Message, error)

	GetConversation(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Conversation, error)

	CreateUser(
		ctx context.Context,
	) (*domain.User, error)

	UserByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.User, error)

	DeleteUser(
		ctx context.Context,
		id uuid.UUID,
	) error

	CreateMessage(
		ctx context.Context,
		parentID domain.Hash,
		conversationID uuid.UUID,
		role domain.Role,
		text string,
	) (*domain.Message, error)

	DeleteMessage(
		ctx context.Context,
		id domain.Hash,
		conversationID uuid.UUID,
	) error
}

func roleFromProto(r conversationv1.Role) (domain.Role, error) {
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

func roleToProto(r domain.Role) (conversationv1.Role, error) {
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

type Handler struct {
	conversationv1.UnimplementedUserServiceServer
	conversationv1.UnimplementedConversationServiceServer
	conversationv1.UnimplementedMessageServiceServer
	service Service
}

func New(
	service Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

var (
	ErrMetadataMissing = errors.New("metadata not found in context")
	ErrUserIDMissing   = errors.New("x-user-id header is missing")
	ErrInvalidUUID     = errors.New("x-user-id is not a valid uuid")
)

func (h *Handler) UserID(ctx context.Context) (uuid.UUID, error) {
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

func (h *Handler) GetConversation(ctx context.Context, req *conversationv1.GetConversationRequest) (*conversationv1.Conversation, error) {
	cnvID, err := uuid.FromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("convert conversation id to uuid: %w", err)
	}

	cnv, err := h.service.GetConversation(ctx, cnvID)
	if err != nil {
		return nil, err
	}

	return &conversationv1.Conversation{
		Id:          cnv.ID[:],
		UserId:      cnv.UserID[:],
		Title:       cnv.Title,
		Preferences: &conversationv1.Conversation_Preferences{},
		CreateTime:  timestamppb.New(cnv.CreatedAt),
		UpdateTime:  timestamppb.New(cnv.UpdatedAt),
	}, nil
}

func (h *Handler) ListConversations(ctx context.Context, req *conversationv1.ListConversationsRequest) (*conversationv1.ListConversationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListConversations not implemented")
}

func (h *Handler) UpdateConversation(ctx context.Context, req *conversationv1.UpdateConversationRequest) (*conversationv1.Conversation, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateConversation not implemented")
}

func (h *Handler) GetMessage(ctx context.Context, req *conversationv1.GetMessageRequest) (*conversationv1.Message, error) {
	msgID, err := domain.HashFromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("convert in bytes into hash: %w", err)
	}

	msg, err := h.service.GetMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}

	role, err := roleToProto(msg.Role)
	if err != nil {
		return nil, fmt.Errorf("convert domain role to proto: %w", err)
	}

	return &conversationv1.Message{
		Id:             msg.ID[:],
		ParentId:       msg.ParentID[:],
		ConversationId: msg.ConversationID[:],
		Role:           role,
		Text:           msg.Text,
		CreateTime:     timestamppb.New(msg.CreatedAt),
	}, nil
}

func (h *Handler) ListMessages(ctx context.Context, req *conversationv1.ListMessagesRequest) (*conversationv1.ListMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListMessages not implemented")
}

func (h *Handler) CreateMessage(
	ctx context.Context,
	req *conversationv1.CreateMessageRequest,
) (*conversationv1.Message, error) {
	var err error

	conversationID := uuid.Nil
	if len(req.ConversationId) > 0 {
		conversationID, err = uuid.FromBytes(req.ConversationId)
		if err != nil {
			return nil, fmt.Errorf("convert conversation id to uuid: %w", err)
		}
	}

	role, err := roleFromProto(req.Role)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	parentID := domain.NilHash
	if len(req.ParentId) > 0 {
		parentID, err = domain.HashFromBytes(req.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid parent id: %w", err)
		}
	}

	msg, err := h.service.CreateMessage(
		ctx,
		parentID,
		conversationID,
		role,
		req.Text,
	)
	if err != nil {
		return nil, err
	}

	protoRole, err := roleToProto(msg.Role)
	if err != nil {
		protoRole = conversationv1.Role_ROLE_UNSPECIFIED
	}

	return &conversationv1.Message{
		Id:             msg.ID[:],
		ParentId:       msg.ParentID[:],
		ConversationId: msg.ConversationID[:],
		Role:           protoRole,
		Text:           msg.Text,
		CreateTime:     timestamppb.New(msg.CreatedAt),
	}, nil
}

func (h *Handler) DeleteMessage(ctx context.Context, req *conversationv1.DeleteMessageRequest) (*emptypb.Empty, error) {
	conversationID, err := uuid.ParseBytes(req.ConversationId)
	if err != nil {
		return nil, fmt.Errorf("convert conversation id to uuid: %w", err)
	}

	id, err := domain.HashFromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse message id: %w", err)
	}

	if err := h.service.DeleteMessage(ctx, id, conversationID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) GetHistory(*conversationv1.GetHistoryRequest, grpc.ServerStreamingServer[conversationv1.Message]) error {
	return status.Error(codes.Unimplemented, "method GetHistory not implemented")
}

func (h *Handler) GetHistoryChunk(ctx context.Context, req *conversationv1.GetHistoryChunkRequest) (*conversationv1.GetHistoryChunkResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetHistoryChunk not implemented")
}

func (h *Handler) ListVariants(ctx context.Context, req *conversationv1.ListVariantsRequest) (*conversationv1.ListVariantsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListVariants not implemented")
}

func (h *Handler) SelectVariant(ctx context.Context, req *conversationv1.SelectVariantRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "method SelectVariant not implemented")
}

func (h *Handler) GetUser(ctx context.Context, req *conversationv1.GetUserRequest) (*conversationv1.User, error) {
	usrID, err := uuid.ParseBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	usr, err := h.service.UserByID(ctx, usrID)
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

func (h *Handler) CreateUser(ctx context.Context, req *conversationv1.CreateUserRequest) (*conversationv1.User, error) {
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

func (h *Handler) UpdateUser(ctx context.Context, req *conversationv1.UpdateUserRequest) (*conversationv1.User, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateUser not implemented")
}

func (h *Handler) DeleteUser(ctx context.Context, req *conversationv1.DeleteUserRequest) (*emptypb.Empty, error) {
	usrID, err := uuid.ParseBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	if err := h.service.DeleteUser(ctx, usrID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
