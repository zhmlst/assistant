package postgres

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	DSN  string
	Host string
	Port uint
	User string
	Pass string
	Args url.Values
	RetryInterval time.Duration
}

func (c *Config) String() string {
	if c.DSN != "" {
		return c.DSN
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Pass),
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port))),
		RawQuery: c.Args.Encode(),
	}

	return u.String()
}

type Pool struct{ *pgxpool.Pool }

func New(ctx context.Context, cfg *Config) (Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.String())
	if err != nil {
		return Pool{}, fmt.Errorf("pgxpool new: %w", err)
	}

	cfg.RetryInterval = max(cfg.RetryInterval, 500*time.Millisecond)

	timer := time.NewTimer(cfg.RetryInterval)
	defer timer.Stop()
	for {
		if err := pool.Ping(ctx); err == nil {
			return Pool{Pool: pool}, nil
		}

		timer.Reset(cfg.RetryInterval)
		select {
		case <-ctx.Done():
			return Pool{}, ctx.Err()
		case <-timer.C:
		}
	}
}


type key struct{}

func (p Pool) Wrap(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, ok := ctx.Value(key{}).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	tx, err = p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}

		if err != nil {
			if errRB := tx.Rollback(ctx); errRB != nil {
				err = fmt.Errorf("rollback tx: %w", errRB)
			}
			return
		}

		if err = tx.Commit(ctx); err != nil {
			err = fmt.Errorf("commit tx: %w", err)
		}
	}()

	return fn(context.WithValue(ctx, key{}, tx))
}

func (p Pool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx, ok := ctx.Value(key{}).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, sql, args...)
	}

	return p.Pool.Exec(ctx, sql, args...)
}

func (p Pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	tx, ok := ctx.Value(key{}).(pgx.Tx)
	if ok {
		return tx.Query(ctx, sql, args...)
	}

	return p.Pool.Query(ctx, sql, args...)
}

func (p Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	tx, ok := ctx.Value(key{}).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, sql, args...)
	}

	return p.Pool.QueryRow(ctx, sql, args...)
}
