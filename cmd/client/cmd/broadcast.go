package cmd

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewBroadcastCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast",
		Short: "Run a command on all connected clients",
	}

	cmd.Flags().StringP("command", "r", "", "command to run")

	lo.Must0(cmd.MarkFlagRequired("command"))

	viperext.MustBindCommandFlags(cc.Viper, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c, err := cli.NewBroadcast(cc, args, cfg)
		if err != nil {
			return fmt.Errorf("creating broadcast cli: %w", err)
		}
		return c.Run(cmd.Context())
	}

	return cmd
}
