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

type Conversation interface {
	History(ctx context.Context, conversationID uuid.UUID, id lib.Hash) iter.Seq2[*domain.Message, error]
	Prompt(ctx context.Context, conversationID uuid.UUID) (string, error)
	CreateMessage(ctx context.Context, conversationID uuid.UUID, text string) error
}

type LLaMA interface {
	Complete(ctx context.Context, history []domain.Message, w io.Writer) error
}

type Redis interface {
	Summary(msgID lib.Hash) (string, error)
	SetSummary(anchor lib.Hash, summary string) error
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
		llama:        llama,
		redis:        redis,
	}
}

func (s *Service) Reply(ctx context.Context, msg *domain.Message) error {
	history := s.collectHistory(ctx, msg)

	w, _ := s.redis.Writer(msg.ConversationID.String())
	defer w.Close()

	var sb strings.Builder
	_ = s.llama.Complete(ctx, history, io.MultiWriter(w, &sb))

	return s.conversation.CreateMessage(ctx, msg.ConversationID, sb.String())
}

const (
	tokPerByte      = 0.25
	maxContextSize  = 4000
	summarizePrompt = `summarize conversation`
)

func (s *Service) collectHistory(ctx context.Context, msg *domain.Message) []domain.Message {
	var history []domain.Message
	var contextSize int

	for m, err := range s.conversation.History(ctx, msg.ConversationID, msg.ID) {
		if err != nil {
			break
		}

		if sum, _ := s.redis.Summary(m.ID); sum != "" {
			history = append(history, domain.Message{Role: lib.RoleSystem, Text: sum})
			break
		}

		history = append(history, *m)
		contextSize += int(float64(len(m.Text)) * tokPerByte)

		if contextSize >= maxContextSize {
			return s.summarizeAndShorten(ctx, history, m.ID, maxContextSize/2)
		}
	}

	if prompt, _ := s.conversation.Prompt(ctx, msg.ConversationID); prompt != "" {
		history = append(history, domain.Message{Role: lib.RoleSystem, Text: prompt})
	}

	slices.Reverse(history)
	return history
}

func (s *Service) summarizeAndShorten(
	ctx context.Context,
	history []domain.Message,
	anchor lib.Hash,
	limit int,
) []domain.Message {
	var splitIdx int
	var current int

	for i, h := range history {
		current += int(float64(len(h.Text)) * tokPerByte)
		if current >= limit {
			splitIdx = i
			break
		}
	}

	oldPart := history[splitIdx:]
	slices.Reverse(oldPart)

	var sb strings.Builder
	prompt := domain.Message{Role: lib.RoleSystem, Text: summarizePrompt}
	_ = s.llama.Complete(ctx, append(oldPart, prompt), &sb)

	summary := sb.String()
	_ = s.redis.SetSummary(anchor, summary)

	result := append(history[:splitIdx], domain.Message{Role: lib.RoleSystem, Text: summary})
	slices.Reverse(result)
	return result
}
