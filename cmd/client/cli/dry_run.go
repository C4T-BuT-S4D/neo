package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

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

func NewDryRun(cc *Context, args []string, cfg *client.Config) (NeoCLI, error) {
	cfg.ExploitDir = path.Join(cfg.ExploitDir, "dry")
	if err := os.MkdirAll(cfg.ExploitDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating dry dir %s: %w", cfg.ExploitDir, err)
	}

	jobs := cc.Viper.GetInt("jobs")
	if jobs < 0 {
		return nil, errors.New("jobs should be non-negative")
	}

	return &dryRunCLI{
		baseCLI:   &baseCLI{cc: cc, cfg: cfg},
		exploitID: args[0],
		jobs:      jobs,
		teamID:    cc.Viper.GetString("team_id"),
		teamIP:    cc.Viper.GetString("team_ip"),
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
	for {
		select {
		case res := <-q.Results():
			if tasksDone+1 == len(tasks) && !ex.Endless {
				zap.L().Info("Finished running sploits, waiting for queue to finish")
				runCancel()
				return nil
			}
			tasksDone++
			zap.L().Info("Result", zap.Any("target", res.Target), zap.String("output", string(res.Out)))
		case <-ctx.Done():
			zap.L().Info("Got interrupt")
			return nil
		}
	}
}
