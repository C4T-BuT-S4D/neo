package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

func NewUpdateCLI(cmd *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	c := &updateCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
	}

	if cmd.Flags().Changed("interval") {
		runEvery, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			logrus.Fatalf("Could not parse run interval: %v", err)
		}
		c.runEvery = &runEvery
	}
	if cmd.Flags().Changed("timeout") {
		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			logrus.Fatalf("Could not parse run timeout: %v", err)
		}
		c.timeout = &timeout
	}
	if cmd.Flags().Changed("endless") {
		endless, err := cmd.Flags().GetBool("endless")
		if err != nil {
			logrus.Fatalf("Could not parse endless: %v", err)
		}
		c.endless = &endless
	}
	if cmd.Flags().Changed("disabled") {
		disabled, err := cmd.Flags().GetBool("disabled")
		if err != nil {
			logrus.Fatalf("Could not parse disabled: %v", err)
		}
		c.disabled = &disabled
	}

	return c
}

func (uc *updateCLI) Run(ctx context.Context) error {
	logrus.Infof("Going to update config for exploit with id = %s", uc.exploitID)

	c, err := uc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := c.Exploit(ctx, uc.exploitID)
	if err != nil {
		logrus.Fatalf("Exploit with id = %s does not exist", uc.exploitID)
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
	logrus.Infof("Updated exploit state: %v", ns)
	return nil
}
