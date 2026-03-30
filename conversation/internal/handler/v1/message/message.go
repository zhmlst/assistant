package message

import (
	"context"
	"fmt"
	"iter"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	v1 "github.com/zhmlst/assistant/conversation/internal/handler/v1"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	History(
		ctx context.Context,
		conversationID uuid.UUID,
		anchorID domain.Hash,
	) iter.Seq2[*domain.Message, error]

	SelectVariant(
		ctx context.Context,
		parentID domain.Hash,
		variantID domain.Hash,
		conversationID uuid.UUID,
	) error

	GetMessage(
		ctx context.Context,
		id domain.Hash,
	) (*domain.Message, error)

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

type handler struct {
	conversationv1.UnimplementedMessageServiceServer
	service service
}

func New(
	service service,
) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetMessage(ctx context.Context, req *conversationv1.GetMessageRequest) (*conversationv1.Message, error) {
	msgID, err := domain.HashFromBytes(req.Id)
	if err != nil {
		return nil, fmt.Errorf("convert in bytes into hash: %w", err)
	}

	msg, err := h.service.GetMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}

	role, err := v1.RoleToProto(msg.Role)
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

func (h *handler) ListMessages(ctx context.Context, req *conversationv1.ListMessagesRequest) (*conversationv1.ListMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListMessages not implemented")
}

func (h *handler) CreateMessage(
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

	role, err := v1.RoleFromProto(req.Role)
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

	protoRole, err := v1.RoleToProto(msg.Role)
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

func (h *handler) DeleteMessage(ctx context.Context, req *conversationv1.DeleteMessageRequest) (*emptypb.Empty, error) {
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

func (h *handler) GetHistory(req *conversationv1.GetHistoryRequest, stream grpc.ServerStreamingServer[conversationv1.Message]) error {
	conversationID, err := uuid.FromBytes(req.ConversationId)
	if err != nil {
		return fmt.Errorf("conversation id from bytes: %w", err)
	}

	anchorID, err := domain.HashFromBytes(req.AnchorMessageId)
	if err != nil {
		return fmt.Errorf("anchor id from bytes: %w", err)
	}

	history := h.service.History(stream.Context(), conversationID, anchorID)
	for msg, err := range history {
		if err != nil {
			return fmt.Errorf("next message: %w", err)
		}

		role, err := v1.RoleToProto(msg.Role)
		if err != nil {
			return fmt.Errorf("convert role to proto: %w", err)
		}

		stream.Send(&conversationv1.Message{
			Id:             msg.ID[:],
			ParentId:       msg.ParentID[:],
			ConversationId: msg.ConversationID[:],
			Role:           role,
			Text:           msg.Text,
			CreateTime:     timestamppb.New(msg.CreatedAt),
		})
	}

	return nil
}

func (h *handler) GetHistoryChunk(ctx context.Context, req *conversationv1.GetHistoryChunkRequest) (*conversationv1.GetHistoryChunkResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetHistoryChunk not implemented")
}

func (h *handler) ListVariants(ctx context.Context, req *conversationv1.ListVariantsRequest) (*conversationv1.ListVariantsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListVariants not implemented")
}

func (h *handler) SelectVariant(ctx context.Context, req *conversationv1.SelectVariantRequest) (*emptypb.Empty, error) {
	conversationID, err := uuid.FromBytes(req.ConversationId)
	if err != nil {
		return nil, fmt.Errorf("conversation id from bytes: %w", err)
	}

	parentID, err := domain.HashFromBytes(req.ParentMessageId)
	if err != nil {
		return nil, fmt.Errorf("parent id from bytes: %w", err)
	}

	variantID, err := domain.HashFromBytes(req.VariantMessageId)
	if err != nil {
		return nil, fmt.Errorf("variant id from bytes: %w", err)
	}

	if err := h.service.SelectVariant(ctx, parentID, variantID, conversationID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
