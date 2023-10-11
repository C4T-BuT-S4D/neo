package cmd

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

// tailCmd represents the tail command
var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update exploit configuration by name",
	Example: "neo update exploit_name -i 2m -t 2m",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewUpdateCLI(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error updating exploit config: %v", err)
		}
		logrus.Debugf("Update finished")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().DurationP("interval", "i", time.Second*30, "run interval")
	updateCmd.PersistentFlags().DurationP("timeout", "t", time.Second*30, "timeout for a single run")
	updateCmd.PersistentFlags().BoolP("endless", "e", false, "mark exploit as endless")
	updateCmd.PersistentFlags().Bool("disabled", false, "mark exploit as disabled")
}
