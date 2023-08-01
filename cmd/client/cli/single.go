package cli

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/internal/client"
)

type singleRunCLI struct {
	*baseCLI
	exploitID string
}

func NewSingleRun(_ *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	return &singleRunCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
	}
}

func (sc *singleRunCLI) Run(ctx context.Context) error {
	c, err := sc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	state, err := c.GetServerState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config from server: %w", err)
	}
	exists := false
	for _, v := range state.Exploits {
		if v.ExploitId == sc.exploitID {
			exists = true
			break
		}
	}

	if !exists {
		logrus.Fatalf("Exploit %s does not exist. Please, add it first.", sc.exploitID)
	}
	if err := c.SingleRun(ctx, sc.exploitID); err != nil {
		return fmt.Errorf("single run failed: %w", err)
	}

	return nil
}
