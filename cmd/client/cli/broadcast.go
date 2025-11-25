package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

type broadcastCLI struct {
	*baseCLI
	cmd string
}

func NewBroadcast(cmd *cobra.Command, _ []string, cfg *client.Config) (NeoCLI, error) {
	command, err := cmd.Flags().GetString("command")
	if err != nil {
		return nil, fmt.Errorf("parsing command flag: %w", err)
	}
	return &broadcastCLI{
		baseCLI: &baseCLI{cfg: cfg},
		cmd:     command,
	}, nil
}

func (bc *broadcastCLI) Run(ctx context.Context) error {
	zap.L().Info("Broadcasting command to all connected clients", zap.String("command", bc.cmd))
	c, err := bc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	if err := c.BroadcastCommand(ctx, bc.cmd); err != nil {
		return fmt.Errorf("making broadcast request: %w", err)
	}
	return nil
}
