package cli

import (
	"context"

	"neo/internal/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type broadcastCLI struct {
	*baseCLI
	cmd string
}

func NewBroadcast(cmd *cobra.Command, _ []string, cfg *client.Config) *broadcastCLI {
	command, err := cmd.Flags().GetString("command")
	if err != nil {
		logrus.Fatalf("Could not get command: %v", err)
	}
	return &broadcastCLI{
		baseCLI: &baseCLI{cfg},
		cmd:     command,
	}
}

func (bc *broadcastCLI) Run(_ context.Context) error {
	logrus.Infof("Broadcasting command %s to all connected clients", bc.cmd)
	return nil
}
