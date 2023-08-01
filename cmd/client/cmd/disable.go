package cmd

import (
	"github.com/c4t-but-s4d/neo/internal/client"

	"github.com/c4t-but-s4d/neo/cmd/client/cli"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// disableCmd represents the disable command
var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable an exploit by id",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewSetDisabled(cmd, args, cfg, true)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error disabling exploit: %v", err)
		}
		logrus.Debugf("Disable finished")
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}
