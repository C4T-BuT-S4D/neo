package cli

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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

func parseJobsFlag(cmd *cobra.Command, name string) int {
	jobs, err := cmd.Flags().GetInt(name)
	if err != nil {
		logrus.Fatalf("Could not get jobs number: %v", err)
	}
	if jobs < 0 {
		logrus.Fatal("run: job count should be non-negative")
	}
	return jobs
}

func NewRun(cmd *cobra.Command, _ []string, cfg *client.Config) NeoCLI {
	cli := &runCLI{
		baseCLI: &baseCLI{cfg: cfg},
	}
	neocli, err := cli.client()
	if err != nil {
		logrus.Fatalf("run: failed to create client: %v", err)
	}

	jobs := parseJobsFlag(cmd, "jobs")
	endlessJobs := parseJobsFlag(cmd, "endless-jobs")
	timeoutScaleTarget, err := cmd.Flags().GetFloat64("timeout-autoscale-target")
	if err != nil {
		logrus.Fatalf("Could not get timeout-autoscale-target flag: %v", err)
	}
	if timeoutScaleTarget < 0 {
		logrus.Fatalf("timeout-autoscale-target should be non-negative")
	}

	neocli.Weight = jobs
	cli.sender = joblogger.NewRemoteSender(neocli)
	cli.run = exploit.NewRunner(
		cli.ClientID(),
		jobs,
		endlessJobs,
		timeoutScaleTarget,
		cfg,
		neocli,
		cli.sender,
	)

	return cli
}

func (rc *runCLI) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rc.sender.Start(ctx)
		logrus.Info("log sender finished")
	}()

	return rc.run.Run(ctx) // nolint:wrapcheck
}
