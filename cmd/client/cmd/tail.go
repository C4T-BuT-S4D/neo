package cmd

import (
	"neo/cmd/client/cli"
	"neo/internal/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// tailCmd represents the add command
var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail exploit logs",
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

	tailCmd.PersistentFlags().String("id", "", "exploit name")
	tailCmd.PersistentFlags().Int64("version", 0, "exploit version")
	tailCmd.PersistentFlags().IntP("tail", "t", 0, "lines to show")
}