package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/internal/exploit"
	"github.com/c4t-but-s4d/neo/v2/pkg/joblogger"
)

const JobsPerCPU = 5

type runCLI struct {
	*baseCLI
	run    *exploit.Runner
	sender joblogger.Sender
}

func NewRun(cmd *cobra.Command, _ []string, cfg *client.Config) (NeoCLI, error) {
	cli := &runCLI{
		baseCLI: &baseCLI{cfg: cfg},
	}
	neocli, err := cli.client()
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	jobs, err := parseJobsFlagE(cmd, "jobs")
	if err != nil {
		return nil, err
	}
	endlessJobs, err := parseJobsFlagE(cmd, "endless-jobs")
	if err != nil {
		return nil, err
	}
	timeoutScaleTarget, err := cmd.Flags().GetFloat64("timeout-autoscale-target")
	if err != nil {
		return nil, fmt.Errorf("getting timeout-autoscale-target flag: %w", err)
	}
	if timeoutScaleTarget < 0 {
		return nil, errors.New("timeout-autoscale-target should be non-negative")
	}

	clientID, err := cli.ClientID()
	if err != nil {
		return nil, err
	}

	neocli.Weight = jobs
	cli.sender = joblogger.NewRemoteSender(neocli)
	cli.run = exploit.NewRunner(
		clientID,
		jobs,
		endlessJobs,
		timeoutScaleTarget,
		cfg,
		neocli,
		cli.sender,
	)

	return cli, nil
}

func (rc *runCLI) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Go(func() {
		rc.sender.Start(ctx)
		zap.L().Info("log sender finished")
	})

	if err := rc.run.Run(ctx); err != nil {
		return fmt.Errorf("running exploit runner: %w", err)
	}
	return nil
}
