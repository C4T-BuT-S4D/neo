package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an exploit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewAdd(cmd, args, cfg)
		if err != nil {
			return fmt.Errorf("creating add cli: %w", err)
		}
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("adding exploit: %w", err)
		}
		zap.L().Debug("Add finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.PersistentFlags().String("id", "", "exploit name")
	addCmd.PersistentFlags().BoolP("dir", "d", false, "add exploit as a directory")
	addCmd.PersistentFlags().DurationP("interval", "i", time.Second*30, "run interval")
	addCmd.PersistentFlags().DurationP("timeout", "t", time.Second*30, "timeout for a single run")
	addCmd.PersistentFlags().BoolP("endless", "e", false, "mark exploit as endless")
	addCmd.PersistentFlags().Bool("disabled", false, "mark exploit as disabled")
}
