package cli

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"neo/internal/client"
	"neo/internal/exploit"
	"runtime"
)

type runCLI struct {
	*baseCLI
	run *exploit.Runner
}

func NewRun(args []string, cfg *client.Config) *runCLI {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	jc := flags.Int("j", 0, "Number of parallel exploits to run. Will use runtime.NumCPU() by default")
	if err := flags.Parse(args); err != nil {
		logrus.Fatalf("run: failed to parse cli flags: %v", err)
	}
	if *jc == 0 {
		*jc = runtime.NumCPU()
	}
	if *jc <= 0 {
		logrus.Fatal("run: job count should be positive")
	}
	cli := &runCLI{
		baseCLI: &baseCLI{c: cfg},
	}
	neocli, err := cli.client()
	if err != nil {
		logrus.Fatalf("run: failed to create client: %v", err)
	}
	neocli.Weight = *jc

	runner := exploit.NewRunner(*jc, cfg.ExploitDir, neocli)
	cli.run = runner
	return cli
}

func (rc *runCLI) Run(ctx context.Context) error {
	return rc.run.Run(ctx)
}
