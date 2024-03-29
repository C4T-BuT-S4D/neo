package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

// singleRunCmd represents the single command
var singleRunCmd = &cobra.Command{
	Use:   "single",
	Short: "Run an exploit once on all teams immediately",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewSingleRun(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error in single run: %v", err)
		}
		logrus.Debugf("Single run finished")
	},
}

func init() {
	rootCmd.AddCommand(singleRunCmd)
}
