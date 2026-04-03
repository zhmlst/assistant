package conversation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	ByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Conversation, error)

	List(
		ctx context.Context,
		params domain.ListParameters,
	) (
		cnvs []domain.Conversation,
		nextPageToken []byte,
		err error,
	)

	Update(
		ctx context.Context,
		cnv *domain.Conversation,
		mask domain.ConversationFieldMask,
	) error
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

func conversationsToProto(cnvs []domain.Conversation) []*conversationv1.Conversation {
	res := make([]*conversationv1.Conversation, len(cnvs))
	for i := range cnvs {
		res[i] = &conversationv1.Conversation{
			Id:          cnvs[i].ID[:],
			UserId:      cnvs[i].UserID[:],
			Title:       cnvs[i].Title,
			Preferences: &conversationv1.Conversation_Preferences{},
			CreateTime:  timestamppb.New(cnvs[i].CreatedAt),
			UpdateTime:  timestamppb.New(cnvs[i].UpdatedAt),
		}
	}
	return res
}

func (h *handler) ListConversations(
	ctx context.Context,
	req *conversationv1.ListConversationsRequest,
) (*conversationv1.ListConversationsResponse, error) {
	cnvs, token, err := h.service.List(ctx, domain.ListParameters{
		PageSize:  int(req.PageSize),
		PageToken: req.PageToken,
		Filter:    req.Filter,
		OrderBy:   req.OrderBy,
	})
	if err != nil {
		return nil, err
	}

	return &conversationv1.ListConversationsResponse{
		Conversations: conversationsToProto(cnvs),
		NextPageToken: token,
	}, nil
}

func fieldMaskFromProto(proto *fieldmaskpb.FieldMask) (domain.ConversationFieldMask, error) {
	var mask domain.ConversationFieldMask
	for _, path := range proto.Paths {
		if path == "title" {
			mask |= domain.ConversationFieldTitle
		}
	}
	return mask, nil
}

func (h *handler) UpdateConversation(ctx context.Context, req *conversationv1.UpdateConversationRequest) (*conversationv1.Conversation, error) {
	cnvID, err := uuid.FromBytes(req.Conversation.Id)
	if err != nil {
		return nil, fmt.Errorf("conversation id from bytes: %w", err)
	}

	cnv := domain.Conversation{
		ID:    cnvID,
		Title: req.Conversation.Title,
	}

	mask, err := fieldMaskFromProto(req.FieldMask)
	if err != nil {
		return nil, fmt.Errorf("conversation field mask from proto: %w", err)
	}

	if err := h.service.Update(ctx, &cnv, mask); err != nil {
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
