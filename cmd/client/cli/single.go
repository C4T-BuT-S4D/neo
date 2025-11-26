package cli

import (
	"context"
	"fmt"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

type singleRunCLI struct {
	*baseCLI
	exploitID string
}

func NewSingleRun(cc *Context, args []string, cfg *client.Config) NeoCLI {
	return &singleRunCLI{
		baseCLI:   &baseCLI{cc: cc, cfg: cfg},
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

	if getExploitFromState(state, sc.exploitID) == nil {
		return fmt.Errorf("exploit %s does not exist, add it first", sc.exploitID)
	}
	if err := c.SingleRun(ctx, sc.exploitID); err != nil {
		return fmt.Errorf("single run failed: %w", err)
	}

	return nil
}
