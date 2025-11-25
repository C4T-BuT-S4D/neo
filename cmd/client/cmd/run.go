package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Neo client",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.UnmarshalConfig()
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewRun(cmd, args, cfg)
		if err != nil {
			return fmt.Errorf("creating run cli: %w", err)
		}
		if err := c.Run(cmd.Context()); err != nil {
			return fmt.Errorf("running client: %w", err)
		}
		zap.L().Debug("Run finished")
		return nil
	},
}

//nolint:gochecknoinits // cli init
func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "workers to run")
	runCmd.Flags().IntP("endless-jobs", "e", 0, "workers to run for endless mode. Default is 0 for no endless mode")
	runCmd.Flags().Float64(
		"timeout-autoscale-target",
		1.5,
		"target upper bound for recurrent exploit worker utilization by scaling timeouts."+
			" Setting this to 0 disables scaling",
	)
}
