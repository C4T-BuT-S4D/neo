package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func NewInfoCommand(cc *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print current state",
	}

	viperext.MustBindCommandFlags(cc.Viper, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := viperext.GetConfig(cc.Viper, &client.Config{})
		if err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
		c := cli.NewInfo(cc, args, cfg)
		return c.Run(cmd.Context())
	}

	return cmd
}
