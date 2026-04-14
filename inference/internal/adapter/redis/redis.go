package redis

import (
	"io"

	"github.com/zhmlst/assistant/go/lib"
	"github.com/redis/go-redis/v9"
)

type adapter struct {
	redis.Client
}

func New() *adapter {
	return &adapter{}
}

func (r *adapter) Summary(msgID lib.Hash) (string, error) {
	return "", nil
}

func (r *adapter) SetSummary(anchor lib.Hash, summary string) error {
	return nil
}

func (r *adapter) Writer(channel string) (io.WriteCloser, error) {
	return nil, nil
}
