package cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c4t-but-s4d/neo/internal/client"
	"github.com/c4t-but-s4d/neo/internal/config"
	"github.com/c4t-but-s4d/neo/internal/exploit"
	"github.com/c4t-but-s4d/neo/internal/queue"
	"github.com/c4t-but-s4d/neo/pkg/joblogger"
)

type dryRunCLI struct {
	*baseCLI
	jobs      int
	exploitID string
	teamID    string
	teamIP    string
}

func NewDryRun(cmd *cobra.Command, args []string, cfg *client.Config) NeoCLI {
	cfg.ExploitDir = path.Join(cfg.ExploitDir, "dry")
	if err := os.MkdirAll(cfg.ExploitDir, os.ModePerm); err != nil {
		logrus.Fatalf("failed to create dry dir (%v): %v", cfg.ExploitDir, err)
	}

	cli := &dryRunCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
		jobs:      parseJobsFlag(cmd, "jobs"),
	}
	cli.teamID, _ = cmd.Flags().GetString("team_id")
	cli.teamIP, _ = cmd.Flags().GetString("team_ip")
	return cli
}

func (rc *dryRunCLI) Run(ctx context.Context) error {
	c, err := rc.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	state, err := c.GetServerState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config from server: %w", err)
	}
	cfg, err := config.FromProto(state.Config)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if getExploitFromState(state, rc.exploitID) == nil {
		return fmt.Errorf("exploit %s does not exist, add it first", rc.exploitID)
	}

	storage := exploit.NewStorage(exploit.NewCache(), rc.baseCLI.cfg.ExploitDir, c)
	storage.UpdateExploits(ctx, state.Exploits)
	ex, ok := storage.Exploit(rc.exploitID)
	if !ok {
		return fmt.Errorf("failed to find exploit '%s' in storage", rc.exploitID)
	}

	allTeams := make(map[string]string)
	for _, tbuck := range state.ClientTeamMap {
		for k, v := range tbuck.Teams {
			allTeams[k] = v
		}
	}

	sender := joblogger.NewDummySender()

	var tasks []*queue.Job
	if rc.teamIP == "" && rc.teamID == "" {
		tasks = exploit.CreateExploitJobs(ex, allTeams, cfg.Environ, sender)
	} else {
		oneTeamMap := make(map[string]string)
		for k, v := range allTeams {
			if k == rc.teamID || v == rc.teamIP {
				oneTeamMap[k] = v
			}
		}
		tasks = exploit.CreateExploitJobs(ex, oneTeamMap, cfg.Environ, sender)
	}

	var q queue.Queue
	if ex.Endless {
		q = queue.NewEndlessQueue(rc.jobs)
	} else {
		q = queue.NewSimpleQueue(rc.jobs)
	}

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		q.Start(runCtx)
		logrus.Info("Queue finished")
	}()

	for _, t := range tasks {
		if err := q.Add(t); err != nil {
			logrus.Errorf("Failed to add task (%+v) to queue: %v", t, err)
		}
	}

	tasksDone := 0
loop:
	for {
		select {
		case res, ok := <-q.Results():
			if !ok || tasksDone+1 == len(tasks) && !ex.Endless {
				logrus.Info("Finished running sploits, waiting for queue to finish")
				break loop
			}
			tasksDone++
			logrus.Infof("Target = %v, Out = %v", res.Target, string(res.Out))
		case <-ctx.Done():
			logrus.Info("Got interrupt")
			return nil
		}
	}
	return nil
}
