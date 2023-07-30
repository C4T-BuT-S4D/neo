package cli

import (
	"context"
	"runtime"

	"neo/internal/client"
	"neo/internal/exploit"
	"neo/pkg/tasklogger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const JobsPerCPU = 5

type runCLI struct {
	*baseCLI
	run    *exploit.Runner
	sender *tasklogger.RemoteSender
}

func parseJobsFlag(cmd *cobra.Command, name string) int {
	jobs, err := cmd.Flags().GetInt(name)
	if err != nil {
		logrus.Fatalf("Could not get jobs number: %v", err)
	}
	if jobs == 0 {
		jobs = runtime.NumCPU() * JobsPerCPU
	}
	if jobs <= 0 {
		logrus.Fatal("run: job count should be positive")
	}
	return jobs
}

func NewRun(cmd *cobra.Command, _ []string, cfg *client.Config) NeoCLI {
	cli := &runCLI{
		baseCLI: &baseCLI{c: cfg},
	}
	neocli, err := cli.client()
	if err != nil {
		logrus.Fatalf("run: failed to create client: %v", err)
	}

	jobs := parseJobsFlag(cmd, "jobs")
	endlessJobs := parseJobsFlag(cmd, "endless-jobs")

	neocli.Weight = jobs
	cli.sender = tasklogger.NewRemoteSender(neocli)
	cli.run = exploit.NewRunner(
		cli.ClientID(),
		jobs,
		endlessJobs,
		cfg,
		neocli,
		cli.sender,
	)
	return cli
}

func (rc *runCLI) Run(ctx context.Context) error {
	go rc.sender.Start(ctx)
	return rc.run.Run(ctx) // nolint:wrapcheck
}
