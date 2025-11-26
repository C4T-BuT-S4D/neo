package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewDryRunCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dry-run",
		Short: "Run exploit locally without server",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("team_ip", "p", "", "ip of team to run")
	cmd.Flags().StringP("team_id", "d", "", "id of team to run")
	cmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "number of workers to run")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewDryRun(cc, args, cfg)
		if err != nil {
			return fmt.Errorf("creating dry-run cli: %w", err)
		}
		return c.Run(cmd.Context())
	}

	return cmd
}
