package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail exploit logs by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewTail(cmd, args, cfg)
		if err != nil {
			return fmt.Errorf("creating tail cli: %w", err)
		}
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("tailing logs: %w", err)
		}
		zap.L().Debug("Tail finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(tailCmd)

	tailCmd.PersistentFlags().Int64("version", 0, "exploit version")
	tailCmd.PersistentFlags().IntP("count", "n", -1, "lines to show (-1 for all lines)")
}
