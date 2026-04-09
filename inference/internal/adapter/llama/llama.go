package llama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/zhmlst/assistant/go/lib"
	"github.com/zhmlst/assistant/inference/internal/domain"
)

type Config struct {
	Addr string
}

func defaultConfig() *Config {
	return &Config{
		Addr: "127.0.0.1:8000",
	}
}

type llama struct {
	config Config
}

func New(cfg *Config) *llama {
	if cfg == nil {
		cfg = defaultConfig()
	}
	return &llama{
		config: *cfg,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func roleToRest(r lib.Role) string {
	switch r {
	case lib.RoleAssistant:
		return "assistant"
	case lib.RoleSystem:
		return "system"
	case lib.RoleUser:
		return "user"
	default:
		panic(fmt.Sprintf("unknown role: %v", r))
	}
}

func messagesToRest(msgs []domain.Message) []Message {
	res := make([]Message, len(msgs))
	for i := range len(msgs) {
		res[i] = Message{
			Role:    roleToRest(msgs[i].Role),
			Content: fmt.Sprintf("[%s] %s", msgs[i].CreatedAt, msgs[i].Text),
		}
	}
	return res
}

type CompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func readCompletion(ctx context.Context, dst io.Writer, src io.Reader) error {
	reader := bufio.NewReader(src)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read line: %w", err)
		}

		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		data, ok := bytes.CutPrefix(line, []byte("data: "))
		if !ok {
			continue
		}

		if bytes.Equal(data, []byte("[DONE]")) {
			return nil
		}

		var chunk CompletionChunk
		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) < 1 {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		if _, err := io.WriteString(dst, content); err != nil {
			return err
		}
	}
}

func (l *llama) Complete(ctx context.Context, history []domain.Message, dst io.Writer) error {
	reqBody := map[string]any{
		"messages": messagesToRest(history),
		"stream":   true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return err
	}

	u := url.URL{
		Scheme: "http",
		Host:   l.config.Addr,
		Path:   "/v1/chat/completions",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buf)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(res.Status)
	}

	return readCompletion(ctx, dst, res.Body)
}
