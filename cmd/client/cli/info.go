package cli

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"neo/internal/config"
	neopb "neo/lib/genproto/neo"
	"strings"

	"neo/internal/client"
)

type infoCLI struct {
	*baseCLI
}

func NewInfo(_ *cobra.Command, _ []string, cfg *client.Config) *infoCLI {
	return &infoCLI{&baseCLI{cfg}}
}

func (ic *infoCLI) Run(ctx context.Context) error {
	c, err := ic.client()
	if err != nil {
		return err
	}
	state, err := c.Ping(ctx, neopb.PingRequest_CONFIG_REQUEST)
	if err != nil {
		return err
	}
	cfg, err := config.FromProto(state.GetConfig())
	if err != nil {
		return err
	}
	fmt.Printf("config: %+v\n", cfg)
	fmt.Println("IPs buckets: ")
	for k, v := range state.GetClientTeamMap() {
		fmt.Print(k, ": [")
		fmt.Print(strings.Join(v.GetTeamIps(), ", "))
		fmt.Println("]")
	}
	fmt.Println("Exploits: ")
	for _, e := range state.GetExploits() {
		fmt.Printf("%+v\n", e)
	}
	return nil
}
