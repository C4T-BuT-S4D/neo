package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable an exploit by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c := cli.NewSetDisabled(cmd, args, cfg, true)
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("disabling exploit: %w", err)
		}
		zap.L().Debug("Disable finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(disableCmd)
}
