package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print current state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c := cli.NewInfo(cmd, args, cfg)
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("getting info: %w", err)
		}
		zap.L().Debug("Info finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(infoCmd)
}
