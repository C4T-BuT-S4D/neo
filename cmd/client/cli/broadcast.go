package cli

import (
	"context"
	"fmt"

	"neo/internal/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type broadcastCLI struct {
	*baseCLI
	cmd string
}

func NewBroadcast(_ *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	return &broadcastCLI{
		baseCLI: &baseCLI{cfg},
		cmd:     args[0],
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
