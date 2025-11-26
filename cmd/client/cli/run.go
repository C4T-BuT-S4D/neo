package cli

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

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

func NewRun(cc *Context, _ []string, cfg *client.Config) (NeoCLI, error) {
	cli := &runCLI{
		baseCLI: &baseCLI{cc: cc, cfg: cfg},
	}
	neocli, err := cli.client()
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	jobs := cc.Viper.GetInt("jobs")
	if jobs < 0 {
		return nil, errors.New("jobs should be non-negative")
	}

	endlessJobs := cc.Viper.GetInt("endless_jobs")
	if endlessJobs < 0 {
		return nil, errors.New("endless-jobs should be non-negative")
	}

	timeoutScaleTarget := cc.Viper.GetFloat64("timeout_autoscale_target")
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
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		rc.sender.Start(gctx)
		zap.L().Info("Log sender finished")
		return nil
	})

	g.Go(func() error {
		if err := rc.run.Run(gctx); err != nil {
			return fmt.Errorf("running exploit runner: %w", err)
		}
		zap.L().Info("Exploit runner finished")
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("running client: %w", err)
	}
	return nil
}
