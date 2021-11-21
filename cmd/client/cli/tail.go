package cli

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"neo/internal/client"

	neopb "neo/lib/genproto/neo"
)

type tailCLI struct {
	*baseCLI
	exploitID string
	version   int64
	tail      int
}

func NewTail(cmd *cobra.Command, _ []string, cfg *client.Config) NeoCLI {
	c := &tailCLI{baseCLI: &baseCLI{cfg}}

	var err error
	if c.exploitID, err = cmd.Flags().GetString("id"); err != nil {
		logrus.Fatalf("Could not get exploit id: %v", err)
	}
	if c.version, err = cmd.Flags().GetInt64("version"); err != nil {
		logrus.Fatalf("Could not get exploit version: %v", err)
	}
	if c.tail, err = cmd.Flags().GetInt("tail"); err != nil {
		logrus.Fatalf("Could not get tail: %v", err)
	}
	return c
}

func (tc *tailCLI) Run(ctx context.Context) error {
	c, err := tc.client()
	if err != nil {
		return err
	}
	state, err := c.Ping(ctx, neopb.PingRequest_CONFIG_REQUEST)
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
	lines, err := c.SearchLogLines(ctx, tc.exploitID, tc.version)
	if err != nil {
		return fmt.Errorf("searching logs: %w", err)
	}
	logrus.Debugf("Got %d log lines", len(lines))
	if len(lines) > tc.tail {
		lines = lines[len(lines)-tc.tail:]
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
