package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewAddCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an exploit",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().String("id", "", "exploit name")
	cmd.Flags().BoolP("dir", "d", false, "add exploit as a directory")
	cmd.Flags().DurationP("interval", "i", time.Second*30, "run interval")
	cmd.Flags().DurationP("timeout", "t", time.Second*30, "timeout for a single run")
	cmd.Flags().BoolP("endless", "e", false, "mark exploit as endless")
	cmd.Flags().Bool("disabled", false, "mark exploit as disabled")

	viperext.MustBindCommandFlags(cc.Viper, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewAdd(cc, args, cfg)
		if err != nil {
			return fmt.Errorf("creating add cli: %w", err)
		}
		return c.Run(cmd.Context())
	}

	return cmd
}
