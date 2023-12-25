package logstor

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	maxRedisStreamLength = 100000
)

var (
	_ Storage = (*RedisStorage)(nil)
)

type RedisStorage struct {
	rdb *redis.Client
}

func NewRedisStorage(ctx context.Context, url string) (*RedisStorage, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}
	c := redis.NewClient(opts)
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	return &RedisStorage{rdb: c}, nil
}

func (s *RedisStorage) Add(ctx context.Context, lines ...*Line) error {
	if _, err := s.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, line := range lines {
			key := getRedisStream(line.Exploit, line.Version)
			vals, err := line.ToRedis()
			if err != nil {
				return fmt.Errorf("serializing %v: %w", line, err)
			}
			args := redis.XAddArgs{
				Stream: key,
				MaxLen: maxRedisStreamLength,
				Approx: true,
				Values: vals,
			}
			if err := pipe.XAdd(ctx, &args).Err(); err != nil {
				return fmt.Errorf("adding %v: %w", line, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("running pipeline: %w", err)
	}
	return nil
}

func (s *RedisStorage) Search(ctx context.Context, exploit string, version int64, opts ...SearchOption) ([]*Line, error) {
	cfg := GetSearchConfig(opts...)
	key := getRedisStream(exploit, version)

	start := "-"
	if cfg.lastToken != "" {
		start = fmt.Sprintf("(%s", cfg.lastToken)
	}
	res, err := s.rdb.XRangeN(ctx, key, start, "+", int64(cfg.limit)).Result()
	if err != nil {
		return nil, fmt.Errorf("querying stream %s: %w", key, err)
	}
	lines := make([]*Line, 0, len(res))
	for _, msg := range res {
		line, err := NewLineFromRedis(msg.Values)
		if err != nil {
			return nil, fmt.Errorf("decoding line from %+v: %w", msg.Values, err)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func getRedisStream(exploit string, version int64) string {
	return fmt.Sprintf("logs:%s:%d", exploit, version)
}
