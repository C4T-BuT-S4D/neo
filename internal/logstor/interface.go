package logstor

import (
	"context"
)

const defaultLimit = 10000

type SearchConfig struct {
	limit     int
	lastToken string
}

type SearchOption func(cfg *SearchConfig)

func SearchWithLimit(limit int) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.limit = limit
	}
}

func SearchWithLastToken(lastToken string) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.lastToken = lastToken
	}
}

func GetSearchConfig(opts ...SearchOption) *SearchConfig {
	cfg := &SearchConfig{
		limit: defaultLimit,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type Storage interface {
	Add(ctx context.Context, lines ...*Line) error
	Search(ctx context.Context, exploit string, version int64, opts ...SearchOption) ([]*Line, error)
}
