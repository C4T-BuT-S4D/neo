package cmd

import (
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Neo client",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewRun(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		logrus.Debugf("Run finished")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntP("jobs", "j", runtime.NumCPU()*cli.JobsPerCPU, "workers to run")
	runCmd.Flags().IntP("endless-jobs", "e", 0, "workers to run for endless mode. Default is 0 for no endless mode")
	runCmd.Flags().Bool("disable-timeout-scaling", false, "disable scaling recurrent exploit timeouts to fit the workers count")
}
