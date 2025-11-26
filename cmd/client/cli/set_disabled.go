package cli

import (
	"context"
	"fmt"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
)

type setDisabledCli struct {
	*baseCLI
	exploitID string
	disabled  bool
}

func NewSetDisabled(cc *Context, args []string, cfg *client.Config, disabled bool) NeoCLI {
	return &setDisabledCli{
		baseCLI:   &baseCLI{cc: cc, cfg: cfg},
		exploitID: args[0],
		disabled:  disabled,
	}
}

func (sc *setDisabledCli) Run(ctx context.Context) error {
	c, err := sc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	state, err := c.GetServerState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config from server: %w", err)
	}

	if spl := getExploitFromState(state, sc.exploitID); spl == nil {
		return fmt.Errorf("exploit %s does not exist", sc.exploitID)
	} else if err := c.SetExploitDisabled(ctx, spl.GetExploitId(), sc.disabled); err != nil {
		return fmt.Errorf("set disabled failed: %w", err)
	}

	return nil
}
