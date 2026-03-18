package logger

import (
	"context"
	"log/slog"
	"os"
)

type Config struct {
	JSON  bool
	Level slog.Level
}

type key struct{}

type handler struct{ slog.Handler }

func (h handler) Handle(ctx context.Context, rec slog.Record) error {
	attrs, ok := ctx.Value(key{}).([]slog.Attr)
	if ok {
		rec.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, rec)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler{h.Handler.WithAttrs(attrs)}
}

func (h handler) WithGroup(name string) slog.Handler {
	return handler{h.Handler.WithGroup(name)}
}

func DefaultConfig() *Config {
	return &Config{
		JSON:  false,
		Level: slog.LevelInfo,
	}
}

func New(cfg *Config) *slog.Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	var h slog.Handler
	if cfg.JSON {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     cfg.Level,
			AddSource: cfg.Level == slog.LevelDebug,
		})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     cfg.Level,
			AddSource: cfg.Level == slog.LevelDebug,
		})
	}
	return slog.New(handler{h})
}

func ContextWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	return context.WithValue(ctx, key{}, attrs)
}
