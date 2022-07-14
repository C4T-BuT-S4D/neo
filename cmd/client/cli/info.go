package cli

import (
	"context"
	"fmt"

	"neo/internal/client"
	"neo/internal/config"

	"github.com/spf13/cobra"

	neopb "neo/lib/genproto/neo"
)

type infoCLI struct {
	*baseCLI
}

func NewInfo(_ *cobra.Command, _ []string, cfg *client.Config) NeoCLI {
	return &infoCLI{&baseCLI{cfg}}
}

func (ic *infoCLI) Run(ctx context.Context) error {
	c, err := ic.client()
	if err != nil {
		return err
	}
	state, err := c.Ping(ctx, neopb.PingRequest_CONFIG_REQUEST)
	if err != nil {
		return fmt.Errorf("making ping config request: %w", err)
	}
	cfg, err := config.FromProto(state.Config)
	if err != nil {
		return fmt.Errorf("unmarshalling config: %w", err)
	}
	fmt.Printf("config: %+v\n", cfg)
	fmt.Println("IPs buckets: ")
	for k, v := range state.ClientTeamMap {
		fmt.Print(k, ": [")
		fmt.Printf("%+v", v.Teams)
		fmt.Println("]")
	}
	fmt.Println("Exploits: ")
	for _, e := range state.Exploits {
		fmt.Printf("%+v\n", e)
	}
	return nil
}
