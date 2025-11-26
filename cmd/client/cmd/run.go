package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewRunCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start Neo client",
	}

	cmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "workers to run")
	cmd.Flags().IntP("endless-jobs", "e", 0, "workers to run for endless mode. Default is 0 for no endless mode")
	cmd.Flags().Float64(
		"timeout-autoscale-target",
		1.5,
		"target upper bound for recurrent exploit worker utilization by scaling timeouts. Setting this to 0 disables scaling",
	)

	viperext.MustBindCommandFlags(cc.Viper, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewRun(cc, args, cfg)
		if err != nil {
			return fmt.Errorf("creating run cli: %w", err)
		}
		return c.Run(cmd.Context())
	}

	return cmd
}
