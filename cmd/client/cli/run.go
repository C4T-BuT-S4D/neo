package cli

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/internal/client"
	"github.com/c4t-but-s4d/neo/internal/exploit"
	"github.com/c4t-but-s4d/neo/pkg/joblogger"
)

const JobsPerCPU = 5

type runCLI struct {
	*baseCLI
	run    *exploit.Runner
	sender *joblogger.RemoteSender
}

func parseJobsFlag(cmd *cobra.Command, name string) int {
	jobs, err := cmd.Flags().GetInt(name)
	if err != nil {
		logrus.Fatalf("Could not get jobs number: %v", err)
	}
	if jobs < 0 {
		logrus.Fatal("run: job count should be non-negavtive")
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

	neocli.Weight = jobs
	cli.sender = joblogger.NewRemoteSender(neocli)
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
