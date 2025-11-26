package cli

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
)

type updateCLI struct {
	*baseCLI
	cmd       *cobra.Command
	exploitID string
}

func NewUpdateCLI(cc *Context, cmd *cobra.Command, args []string, cfg *client.Config) (NeoCLI, error) {
	return &updateCLI{
		baseCLI:   &baseCLI{cc: cc, cfg: cfg},
		cmd:       cmd,
		exploitID: args[0],
	}, nil
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
	if uc.cmd.Flags().Changed("interval") {
		runEvery = durationpb.New(uc.cc.Viper.GetDuration("interval"))
	}
	timeout := escfg.GetTimeout()
	if uc.cmd.Flags().Changed("timeout") {
		timeout = durationpb.New(uc.cc.Viper.GetDuration("timeout"))
	}

	var endless *bool
	if uc.cmd.Flags().Changed("endless") {
		endless = lo.ToPtr(uc.cc.Viper.GetBool("endless"))
	}
	var disabled *bool
	if uc.cmd.Flags().Changed("disabled") {
		disabled = lo.ToPtr(uc.cc.Viper.GetBool("disabled"))
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
			Endless:    lo.FromPtrOr(endless, escfg.GetEndless()),
			Disabled:   lo.FromPtrOr(disabled, escfg.GetDisabled()),
		},
	}

	ns, err := c.UpdateExploit(ctx, newState)
	if err != nil {
		return fmt.Errorf("failed to update exploit: %w", err)
	}
	zap.L().Info("Updated exploit state", zap.Any("state", ns))
	return nil
}
