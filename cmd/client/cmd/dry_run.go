package cmd

import (
	"runtime"

	"neo/cmd/client/cli"
	"neo/internal/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Start Neo client",
	Args:  cobra.MinimumNArgs(1),
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
	dryRunCmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "exploit jobs multiplier")
}