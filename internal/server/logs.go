package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	linesPerSploitLimit = 1000
)

func NewLogStorage(ctx context.Context, redisURL string) (*LogStorage, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}
	c := redis.NewClient(opts)
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	return &LogStorage{rdb: c}, nil
}

type LogStorage struct {
	rdb *redis.Client
}

func (s *LogStorage) Add(ctx context.Context, lines []LogLine) error {
	if _, err := s.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, line := range lines {
			key := getRedisStream(line.Exploit, line.Version)
			vals, err := line.DumpValues()
			if err != nil {
				return fmt.Errorf("serializing %v: %w", line, err)
			}
			args := redis.XAddArgs{
				Stream: key,
				MaxLen: linesPerSploitLimit,
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

func (s *LogStorage) Get(ctx context.Context, opts GetOptions) ([]LogLine, error) {
	key := getRedisStream(opts.Exploit, strconv.FormatInt(opts.Version, 10))
	res, err := s.rdb.XRange(ctx, key, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("querying stream %s: %w", key, err)
	}
	lines := make([]LogLine, 0, len(res))
	for _, msg := range res {
		line, err := NewLogLineFromValues(msg.Values)
		if err != nil {
			return nil, fmt.Errorf("decoding line from %+v: %w", msg.Values, err)
		}
		lines = append(lines, *line)
	}
	return lines, nil
}

type GetOptions struct {
	Exploit string
	Version int64
}

func getRedisStream(exploit string, version string) string {
	return fmt.Sprintf("logs:%s:%s", exploit, version)
}
