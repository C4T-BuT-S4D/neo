package cli

import (
	"context"
	"fmt"
	"neo/internal/config"

	"neo/internal/client"
)

type infoCLI struct {
	*baseCLI
}

func NewInfo(args []string, cfg *client.Config) *infoCLI {
	return &infoCLI{&baseCLI{cfg}}
}

func (ic *infoCLI) Run(ctx context.Context) error {
	c, err := ic.client()
	if err != nil {
		return err
	}
	state, err := c.Ping(ctx)
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
		fmt.Print(k, ": ")
		for _, ip := range v.GetTeamIps() {
			fmt.Print(ip, ", ")
		}
		fmt.Println()
	}
	fmt.Println("Exploits: ")
	for _, e := range state.GetExploits() {
		fmt.Printf("%+v\n", e)
	}
	return nil
}
