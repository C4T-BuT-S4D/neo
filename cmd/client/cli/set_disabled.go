package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	epb "github.com/c4t-but-s4d/neo/proto/go/exploits"

	"github.com/c4t-but-s4d/neo/internal/client"
)

type setDisabledCli struct {
	*baseCLI
	exploitID string
	disabled  bool
}

func NewSetDisabled(_ *cobra.Command, args []string, cfg *client.Config, disabled bool) NeoCLI {
	return &setDisabledCli{
		baseCLI:   &baseCLI{cfg: cfg},
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

	var spl *epb.ExploitState
	for _, v := range state.Exploits {
		if v.ExploitId == sc.exploitID {
			spl = v
			break
		}
	}

	if spl == nil {
		return fmt.Errorf("exploit %s does not exist", sc.exploitID)
	}
	if err := c.SetExploitDisabled(ctx, spl.ExploitId, sc.disabled); err != nil {
		return fmt.Errorf("set disabled failed: %w", err)
	}

	return nil
}
