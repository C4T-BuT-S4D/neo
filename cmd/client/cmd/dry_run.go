package cmd

import (
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

// runCmd represents the run command
var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Start Neo client",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewDryRun(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		logrus.Debugf("Dry run finished")
	},
}

func init() {
	rootCmd.AddCommand(dryRunCmd)
	dryRunCmd.Flags().StringP("team_ip", "p", "", "ip of team to run")
	dryRunCmd.Flags().StringP("team_id", "d", "", "id of team to run")
	dryRunCmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "number of workers to run")
}
