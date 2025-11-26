package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewTailCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Tail exploit logs by name",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().Int64("version", 0, "exploit version")
	cmd.Flags().IntP("count", "n", -1, "lines to show (-1 for all lines)")

	viperext.MustBindCommandFlags(cc.Viper, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewTail(cc, args, cfg)
		if err != nil {
			return fmt.Errorf("creating tail cli: %w", err)
		}
		return c.Run(cmd.Context())
	}

	return cmd
}
