package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a disabled exploit by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c := cli.NewSetDisabled(cmd, args, cfg, false)
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("enabling exploit: %w", err)
		}
		zap.L().Debug("Enable finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(enableCmd)
}
