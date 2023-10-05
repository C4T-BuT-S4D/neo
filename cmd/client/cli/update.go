package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/c4t-but-s4d/neo/internal/client"

	epb "github.com/c4t-but-s4d/neo/proto/go/exploits"
)

type updateConfigCLI struct {
	*baseCLI
	exploitID string
	runEvery  *time.Duration
	timeout   *time.Duration
	endless   *bool
	disabled  *bool
}

func NewUpdateConfig(cmd *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	c := &updateConfigCLI{
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

func (ac *updateConfigCLI) Run(ctx context.Context) error {
	logrus.Infof("Going to update config for exploit with id = %s", ac.exploitID)

	c, err := ac.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := c.Exploit(ctx, ac.exploitID)
	if err != nil {
		logrus.Fatalf("Exploit with id = %s does not exist", ac.exploitID)
	}
	es := resp.GetState()
	escfg := es.GetConfig()

	runEvery := escfg.GetRunEvery()
	if ac.runEvery != nil {
		runEvery = durationpb.New(*ac.runEvery)
	}
	timeout := escfg.GetTimeout()
	if ac.timeout != nil {
		timeout = durationpb.New(*ac.timeout)
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
			Endless:    updateIfSet(escfg.GetEndless(), ac.endless),
			Disabled:   updateIfSet(escfg.GetEndless(), ac.disabled),
		},
	}

	ns, err := c.UpdateExploit(ctx, newState)
	if err != nil {
		return fmt.Errorf("failed to update exploit: %w", err)
	}
	logrus.Infof("Updated exploit state: %v", ns)
	return nil
}

func updateIfSet[T any](e T, n *T) T {
	if n != nil {
		return *n
	}
	return e
}
