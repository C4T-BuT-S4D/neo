package cli

import (
	"context"
	"runtime"

	"neo/internal/client"
	"neo/internal/exploit"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const JobsPerCPU = 5

type runCLI struct {
	*baseCLI
	run *exploit.Runner
}

func parseJobsFlag(cmd *cobra.Command) int {
	jobs, err := cmd.Flags().GetInt("jobs")
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

	jobs := parseJobsFlag(cmd)

	neocli.Weight = jobs

	cli.run = exploit.NewRunner(jobs, cfg.ExploitDir, neocli)
	return cli
}

func (rc *runCLI) Run(ctx context.Context) error {
	return rc.run.Run(ctx) // nolint:wrapcheck
}
