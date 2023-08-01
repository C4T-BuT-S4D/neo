package cli

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/internal/client"
	logspb "github.com/c4t-but-s4d/neo/proto/go/logs"
)

type tailCLI struct {
	*baseCLI
	exploitID string
	version   int64
	count     int
}

func NewTail(cmd *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	c := &tailCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
	}

	var err error
	if c.version, err = cmd.Flags().GetInt64("version"); err != nil {
		logrus.Fatalf("Could not get exploit version: %v", err)
	}
	if c.count, err = cmd.Flags().GetInt("count"); err != nil {
		logrus.Fatalf("Could not get count: %v", err)
	}
	return c
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
	for _, ex := range state.Exploits {
		if ex.ExploitId == tc.exploitID {
			found = true
			if tc.version == 0 {
				tc.version = ex.Version
			}
			if ex.Version > tc.version {
				return fmt.Errorf("too fresh version requested, current is %v", ex.Version)
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
	logrus.Debugf("Got %d log lines", len(lines))
	if tc.count != -1 && len(lines) > tc.count {
		lines = lines[len(lines)-tc.count:]
	}

	logrus.SetLevel(logrus.DebugLevel)
	for _, line := range lines {
		logger := logrus.WithFields(logrus.Fields{
			"exploit": line.Exploit,
			"version": line.Version,
			"team":    line.Team,
		})
		switch line.Level {
		case "debug":
			logger.Debug(line.Message)
		case "info":
			logger.Info(line.Message)
		case "warning":
			logger.Warning(line.Message)
		case "error":
			logger.Error(line.Message)
		default:
			logger.Warningf("Unexpected log level: %v", line.Level)
			logger.Warning(line.Message)
		}
	}
	return nil
}
