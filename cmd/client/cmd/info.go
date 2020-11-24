package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"neo/cmd/client/cli"
	"neo/internal/client"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print current state",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewInfo(cmd, args, cfg)
		ctx := context.Background()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		logrus.Debugf("Info finished")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
