package cmd

import (
	"neo/cmd/client/cli"
	"neo/internal/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// tailCmd represents the tail command
var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail exploit logs by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := client.MustUnmarshalConfig()
		cli := cli.NewTail(cmd, args, cfg)
		ctx := cmd.Context()
		if err := cli.Run(ctx); err != nil {
			logrus.Fatalf("Error tailing logs: %v", err)
		}
		logrus.Debugf("Tail finished")
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)

	tailCmd.PersistentFlags().Int64("version", 0, "exploit version")
	tailCmd.PersistentFlags().IntP("count", "n", -1, "lines to show (-1 for all lines)")
}
