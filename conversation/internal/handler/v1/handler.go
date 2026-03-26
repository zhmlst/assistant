package v1

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc/metadata"
)

type Service interface {
	CreateMessage(
		ctx context.Context,
		parentID domain.Hash,
		conversationID uuid.UUID,
		role domain.Role,
		text string,
	) (*domain.Message, error)

	DeleteMessage (
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
	conversationv1.UnimplementedMessageServiceServer
	conversationv1.ConversationServiceServer
	service Service
}

func New(
	service Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) UserID(ctx context.Context) (uuid.UUID, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, false
	}

	xuserid := md.Get("x-user-id")
	if len(xuserid) < 1 {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(xuserid[0])
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

func (h *Handler) GetConversation(ctx context.Context, req *conversationv1.GetConversationRequest) (*conversationv1.Conversation, error) {
	return nil, status.Error(codes.Unimplemented, "method GetConversation not implemented")
}

func (h *Handler) ListConversations(ctx context.Context, req *conversationv1.ListConversationsRequest) (*conversationv1.ListConversationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListConversations not implemented")
}

func (h *Handler) UpdateConversation(ctx context.Context, req *conversationv1.UpdateConversationRequest) (*conversationv1.Conversation, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateConversation not implemented")
}

func (h *Handler) GetMessage(ctx context.Context, req *conversationv1.GetMessageRequest) (*conversationv1.Message, error) {
	return nil, status.Error(codes.Unimplemented, "method GetMessage not implemented")
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
		conversationID, err = uuid.ParseBytes(req.ConversationId)
		if err != nil {
			return nil, fmt.Errorf("parse conversation id: %w", err)
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
	conversationID , err := uuid.ParseBytes(req.ConversationId)
	if err != nil {
		return nil, fmt.Errorf("parse conversation id: %w", err)
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
