package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var singleRunCmd = &cobra.Command{
	Use:   "single",
	Short: "Run an exploit once on all teams immediately",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c := cli.NewSingleRun(cmd, args, cfg)
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("running single exploit: %w", err)
		}
		zap.L().Debug("Single run finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(singleRunCmd)
}
