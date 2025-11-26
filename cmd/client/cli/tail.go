package cli

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

type tailCLI struct {
	*baseCLI
	exploitID string
	version   int64
	count     int
}

func NewTail(cc *Context, args []string, cfg *client.Config) (NeoCLI, error) {
	return &tailCLI{
		baseCLI:   &baseCLI{cc: cc, cfg: cfg},
		exploitID: args[0],
		version:   cc.Viper.GetInt64("version"),
		count:     cc.Viper.GetInt("count"),
	}, nil
}

func (tc *tailCLI) Run(ctx context.Context) error {
	c, err := tc.client()
	if err != nil {
		return err
	}
	state, err := c.GetServerState(ctx)
	if err != nil {
		return fmt.Errorf("making ping config request: %w", err)
	}
	found := false
	for _, ex := range state.GetExploits() {
		if ex.GetExploitId() == tc.exploitID {
			found = true
			if tc.version == 0 {
				tc.version = ex.GetVersion()
			}
			if ex.GetVersion() > tc.version {
				return fmt.Errorf("too fresh version requested, current is %v", ex.GetVersion())
			}
		}
	}
	if !found {
		return fmt.Errorf("could not locate exploit %v (v%v)", tc.exploitID, tc.version)
	}

	stream, err := c.SearchLogLines(ctx, tc.exploitID, tc.version)
	if err != nil {
		return fmt.Errorf("making search request: %w", err)
	}
	var lines []*logspb.LogLine
	for batch := range stream {
		lines = append(lines, batch...)
	}
	zap.L().Debug("Got log lines", zap.Int("count", len(lines)))
	if tc.count != -1 && len(lines) > tc.count {
		lines = lines[len(lines)-tc.count:]
	}

	for _, line := range lines {
		logger := zap.L().With(
			zap.String("exploit", line.GetExploit()),
			zap.Int64("version", line.GetVersion()),
			zap.String("team", line.GetTeam()),
		)
		switch line.GetLevel() {
		case "debug":
			logger.Debug(line.GetMessage())
		case "info":
			logger.Info(line.GetMessage())
		case "warning":
			logger.Warn(line.GetMessage())
		case "error":
			logger.Error(line.GetMessage())
		default:
			logger.Warn("Unexpected log level", zap.String("level", line.GetLevel()))
			logger.Warn(line.GetMessage())
		}
	}
	return nil
}
