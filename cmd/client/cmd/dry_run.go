package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Start Neo client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewDryRun(cmd, args, cfg)
		if err != nil {
			return fmt.Errorf("creating dry-run cli: %w", err)
		}
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("running dry-run: %w", err)
		}
		zap.L().Debug("Dry run finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(dryRunCmd)
	dryRunCmd.Flags().StringP("team_ip", "p", "", "ip of team to run")
	dryRunCmd.Flags().StringP("team_id", "d", "", "id of team to run")
	dryRunCmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "number of workers to run")
}
