package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a disabled exploit by id",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewSetDisabled(cmd, args, cfg, false)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error enabling exploit: %v", err)
		}
		logrus.Debugf("Enable finished")
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
