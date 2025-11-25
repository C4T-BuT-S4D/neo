package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
)

type updateCLI struct {
	*baseCLI
	exploitID string
	runEvery  *time.Duration
	timeout   *time.Duration
	endless   *bool
	disabled  *bool
}

func NewUpdateCLI(cmd *cobra.Command, args []string, cfg *client.Config) (NeoCLI, error) {
	c := &updateCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
	}

	if cmd.Flags().Changed("interval") {
		runEvery, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			return nil, fmt.Errorf("parsing run interval: %w", err)
		}
		c.runEvery = &runEvery
	}
	if cmd.Flags().Changed("timeout") {
		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			return nil, fmt.Errorf("parsing run timeout: %w", err)
		}
		c.timeout = &timeout
	}
	if cmd.Flags().Changed("endless") {
		endless, err := cmd.Flags().GetBool("endless")
		if err != nil {
			return nil, fmt.Errorf("parsing endless flag: %w", err)
		}
		c.endless = &endless
	}
	if cmd.Flags().Changed("disabled") {
		disabled, err := cmd.Flags().GetBool("disabled")
		if err != nil {
			return nil, fmt.Errorf("parsing disabled flag: %w", err)
		}
		c.disabled = &disabled
	}

	return c, nil
}

func (uc *updateCLI) Run(ctx context.Context) error {
	zap.L().Info("Going to update config for exploit", zap.String("exploit_id", uc.exploitID))

	c, err := uc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := c.Exploit(ctx, uc.exploitID)
	if err != nil {
		return fmt.Errorf("exploit %s does not exist: %w", uc.exploitID, err)
	}
	es := resp.GetState()
	escfg := es.GetConfig()

	runEvery := escfg.GetRunEvery()
	if uc.runEvery != nil {
		runEvery = durationpb.New(*uc.runEvery)
	}
	timeout := escfg.GetTimeout()
	if uc.timeout != nil {
		timeout = durationpb.New(*uc.timeout)
	}

	newState := &epb.ExploitState{
		ExploitId: es.GetExploitId(),
		File:      es.GetFile(),
		Version:   es.GetVersion(),
		Config: &epb.ExploitConfiguration{
			Entrypoint: escfg.GetEntrypoint(),
			IsArchive:  escfg.GetIsArchive(),
			RunEvery:   runEvery,
			Timeout:    timeout,
			Endless:    lo.FromPtrOr(uc.endless, escfg.GetEndless()),
			Disabled:   lo.FromPtrOr(uc.disabled, escfg.GetDisabled()),
		},
	}

	ns, err := c.UpdateExploit(ctx, newState)
	if err != nil {
		return fmt.Errorf("failed to update exploit: %w", err)
	}
	zap.L().Info("Updated exploit state", zap.Any("state", ns))
	return nil
}
