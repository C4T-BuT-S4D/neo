package cli

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

type broadcastCLI struct {
	*baseCLI
	cmd string
}

func NewBroadcast(cc *Context, _ []string, cfg *client.Config) (NeoCLI, error) {
	return &broadcastCLI{
		baseCLI: &baseCLI{cc: cc, cfg: cfg},
		cmd:     cc.Viper.GetString("command"),
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
