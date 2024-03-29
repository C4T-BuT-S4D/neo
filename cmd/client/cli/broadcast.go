package cli

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

type broadcastCLI struct {
	*baseCLI
	cmd string
}

func NewBroadcast(cmd *cobra.Command, _ []string, cfg *client.Config) NeoCLI {
	command, err := cmd.Flags().GetString("command")
	if err != nil {
		logrus.Fatalf("Could not parse command: %v", cmd)
	}
	return &broadcastCLI{
		baseCLI: &baseCLI{cfg: cfg},
		cmd:     command,
	}
}

func (bc *broadcastCLI) Run(ctx context.Context) error {
	logrus.Infof("Broadcasting command %s to all connected clients", bc.cmd)
	c, err := bc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	if err := c.BroadcastCommand(ctx, bc.cmd); err != nil {
		return fmt.Errorf("making broadcast request: %w", err)
	}
	return nil
}
