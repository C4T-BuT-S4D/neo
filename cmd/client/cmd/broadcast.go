package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var broadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "Run a command on all connected clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewBroadcast(cmd, args, cfg)
		if err != nil {
			return fmt.Errorf("creating broadcast cli: %w", err)
		}
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("broadcasting command: %w", err)
		}
		zap.L().Debug("Broadcast finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(broadcastCmd)
	broadcastCmd.Flags().StringP("command", "r", "", "command to run")
}
