package cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/internal/config"
	"github.com/c4t-but-s4d/neo/v2/internal/exploit"
	"github.com/c4t-but-s4d/neo/v2/internal/queue"
	"github.com/c4t-but-s4d/neo/v2/pkg/joblogger"
)

type dryRunCLI struct {
	*baseCLI
	jobs      int
	exploitID string
	teamID    string
	teamIP    string
}

func NewDryRun(cmd *cobra.Command, args []string, cfg *client.Config) (NeoCLI, error) {
	cfg.ExploitDir = path.Join(cfg.ExploitDir, "dry")
	if err := os.MkdirAll(cfg.ExploitDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating dry dir %s: %w", cfg.ExploitDir, err)
	}

	jobs, err := parseJobsFlagE(cmd, "jobs")
	if err != nil {
		return nil, err
	}

	teamID, err := cmd.Flags().GetString("team_id")
	if err != nil {
		return nil, fmt.Errorf("parsing team_id flag: %w", err)
	}
	teamIP, err := cmd.Flags().GetString("team_ip")
	if err != nil {
		return nil, fmt.Errorf("parsing team_ip flag: %w", err)
	}

	return &dryRunCLI{
		baseCLI:   &baseCLI{cfg: cfg},
		exploitID: args[0],
		jobs:      jobs,
		teamID:    teamID,
		teamIP:    teamIP,
	}, nil
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
	cfg, err := config.FromProto(state.GetConfig())
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if getExploitFromState(state, rc.exploitID) == nil {
		return fmt.Errorf("exploit %s does not exist, add it first", rc.exploitID)
	}

	storage := exploit.NewStorage(exploit.NewCache(), rc.cfg.ExploitDir, c)
	storage.UpdateExploits(ctx, state.GetExploits())
	ex, ok := storage.Exploit(rc.exploitID)
	if !ok {
		return fmt.Errorf("failed to find exploit '%s' in storage", rc.exploitID)
	}

	allTeams := make(map[string]string)
	for _, tbuck := range state.GetClientTeamMap() {
		for k, v := range tbuck.GetTeams() {
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

	wg.Go(func() {
		q.Start(runCtx)
		zap.L().Info("Queue finished")
	})

	for _, t := range tasks {
		if err := q.Add(t); err != nil {
			zap.L().Error("Failed to add task to queue", zap.Any("task", t), zap.Error(err))
		}
	}

	tasksDone := 0
loop:
	for {
		select {
		case res, ok := <-q.Results():
			if !ok || tasksDone+1 == len(tasks) && !ex.Endless {
				zap.L().Info("Finished running sploits, waiting for queue to finish")
				break loop
			}
			tasksDone++
			zap.L().Info("Result", zap.Any("target", res.Target), zap.String("output", string(res.Out)))
		case <-ctx.Done():
			zap.L().Info("Got interrupt")
			return nil
		}
	}
	return nil
}

func parseJobsFlagE(cmd *cobra.Command, name string) (int, error) {
	jobs, err := cmd.Flags().GetInt(name)
	if err != nil {
		return 0, fmt.Errorf("getting %s flag: %w", name, err)
	}
	if jobs < 0 {
		return 0, fmt.Errorf("%s should be non-negative", name)
	}
	return jobs, nil
}
