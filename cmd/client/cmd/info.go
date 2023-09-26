package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/internal/client"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print current state",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewInfo(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		logrus.Debugf("Info finished")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
