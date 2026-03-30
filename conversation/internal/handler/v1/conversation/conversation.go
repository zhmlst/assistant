package conversation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	ByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Conversation, error)
}

type handler struct {
	conversationv1.ConversationServiceServer
	service service
}

func New(service service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ListConversations(ctx context.Context, req *conversationv1.ListConversationsRequest) (*conversationv1.ListConversationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListConversations not implemented")
}

func (h *handler) UpdateConversation(ctx context.Context, req *conversationv1.UpdateConversationRequest) (*conversationv1.Conversation, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateConversation not implemented")
}

func (h *handler) GetConversation(ctx context.Context, req *conversationv1.GetConversationRequest) (*conversationv1.Conversation, error) {
	cnvID, err := uuid.FromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("convert conversation id to uuid: %w", err)
	}

	cnv, err := h.service.ByID(ctx, cnvID)
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
