package service

import (
	"context"
	"fmt"
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

type Config struct {
	ContextSize     int
	TokensPerByte   float64
	SummarizePrompt string
}

func defaultConfig() *Config {
	return &Config{
		ContextSize:     4000,
		TokensPerByte:   0.25,
		SummarizePrompt: `summarize conversation`,
	}
}

type Service struct {
	config       Config
	conversation Conversation
	llama        LLaMA
	redis        Redis
}

func New(
	config *Config,
	conversation Conversation,
	llama LLaMA,
	redis Redis,
) *Service {
	if config == nil {
		config = defaultConfig()
	}
	return &Service{
		config:       *config,
		conversation: conversation,
		llama:        llama,
		redis:        redis,
	}
}

func (s *Service) Reply(ctx context.Context, msg *domain.Message) error {
	history, err := s.collectHistory(ctx, msg)
	if err != nil {
		return fmt.Errorf("collect history: %w", err)
	}

	w, err := s.redis.Writer(msg.ConversationID.String())
	if err != nil {
		return fmt.Errorf("get redis writer: %w", err)
	}
	defer w.Close()

	var sb strings.Builder
	if err := s.llama.Complete(ctx, history, io.MultiWriter(w, &sb)); err != nil {
		return fmt.Errorf("complete llama generation: %w", err)
	}

	if err := s.conversation.CreateMessage(ctx, msg.ConversationID, sb.String()); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

func (s *Service) collectHistory(ctx context.Context, msg *domain.Message) ([]domain.Message, error) {
	var history []domain.Message
	var contextSize int

	for m, err := range s.conversation.History(ctx, msg.ConversationID, msg.ID) {
		if err != nil {
			return nil, fmt.Errorf("iterate conversation history: %w", err)
		}

		sum, err := s.redis.Summary(m.ID)
		if err != nil {
			return nil, fmt.Errorf("get summary from redis: %w", err)
		}

		if sum != "" {
			history = append(history, domain.Message{Role: lib.RoleSystem, Text: sum})
			break
		}

		history = append(history, *m)
		contextSize += int(float64(len(m.Text)) * s.config.TokensPerByte)

		if contextSize >= s.config.ContextSize {
			return s.summarizeAndShorten(ctx, history, m.ID, s.config.ContextSize/2)
		}
	}

	prompt, err := s.conversation.Prompt(ctx, msg.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("get conversation prompt: %w", err)
	}

	if prompt != "" {
		history = append(history, domain.Message{Role: lib.RoleSystem, Text: prompt})
	}

	slices.Reverse(history)
	return history, nil
}

func (s *Service) summarizeAndShorten(
	ctx context.Context,
	history []domain.Message,
	anchor lib.Hash,
	limit int,
) ([]domain.Message, error) {
	var splitIdx int
	var current int

	for i, h := range history {
		current += int(float64(len(h.Text)) * s.config.TokensPerByte)
		if current >= limit {
			splitIdx = i
			break
		}
	}

	oldPart := history[splitIdx:]
	slices.Reverse(oldPart)

	var sb strings.Builder
	prompt := domain.Message{Role: lib.RoleSystem, Text: s.config.SummarizePrompt}
	if err := s.llama.Complete(ctx, append(oldPart, prompt), &sb); err != nil {
		return nil, fmt.Errorf("complete summarization: %w", err)
	}

	summary := sb.String()
	if err := s.redis.SetSummary(anchor, summary); err != nil {
		return nil, fmt.Errorf("save summary to redis: %w", err)
	}

	result := append(history[:splitIdx], domain.Message{Role: lib.RoleSystem, Text: summary})
	slices.Reverse(result)
	return result, nil
}
