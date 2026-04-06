package service

import (
	"context"
	"io"
	"iter"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zhmlst/assistant/go/lib"
	"github.com/zhmlst/assistant/inference/internal/domain"
)

const maxContextSize = 4000
type Conversation interface {
	History(ctx context.Context, msg *domain.Message) iter.Seq2[*domain.Message, error]
	CreateMessage(ctx context.Context, conversationID uuid.UUID, text string) error
}

type LLaMA interface {
	Complete(ctx context.Context, history []domain.Message, w io.Writer) error
}

type Redis interface {
	Summary(msgID lib.Hash) (*domain.Message, error)
	Writer(channel string) (io.WriteCloser, error)
}

type Service struct {
	conversation Conversation
	llama        LLaMA
	redis        Redis
}

func New(
	conversation Conversation,
	llama LLaMA,
	redis Redis,
) *Service {
	return &Service{
		conversation: conversation,
		llama: llama,
		redis: redis,
	}
}

func (s *Service) Reply(ctx context.Context, msg *domain.Message) error {
	if msg.Role == lib.RoleAssistant {
		return nil
	}

	var history []domain.Message
	var contextSize int
	for msg := range s.conversation.History(ctx, msg) {
		if sum, _ := s.redis.Summary(msg.ID); sum != nil {
			history = append(history, *sum)
			break
		}

		history = append(history, *msg)
		contextSize += len(msg.Text) / 3
		if contextSize >= maxContextSize {
			break
		}
	}

	slices.Reverse(history)
	history = append(history, *msg)

	w, _ := s.redis.Writer(msg.ConversationID.String())

	var sb strings.Builder
	_ = s.llama.Complete(ctx, history, io.MultiWriter(&sb, w))
	w.Close()

	_ = s.conversation.CreateMessage(ctx, msg.ConversationID, sb.String())

	return nil
}
