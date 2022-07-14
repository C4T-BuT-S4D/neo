package cli

import (
	"context"
	"fmt"
	"os"
	"path"

	"neo/internal/client"
	"neo/internal/config"
	"neo/internal/exploit"
	"neo/internal/queue"
	"neo/pkg/tasklogger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	neopb "neo/lib/genproto/neo"
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
		baseCLI:   &baseCLI{c: cfg},
		exploitID: args[0],
		jobs:      parseJobsFlag(cmd),
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
	state, err := c.Ping(ctx, neopb.PingRequest_CONFIG_REQUEST)
	if err != nil {
		return fmt.Errorf("failed to get config from server: %w", err)
	}
	cfg, err := config.FromProto(state.Config)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	exists := false
	for _, v := range state.Exploits {
		if v.ExploitId == rc.exploitID {
			exists = true
			break
		}
	}
	if !exists {
		return fmt.Errorf("exploit %s does not exist, add it first", rc.exploitID)
	}

	storage := exploit.NewStorage(exploit.NewCache(), rc.baseCLI.c.ExploitDir, c)
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

	sender := tasklogger.NewDummySender()

	var tasks []queue.Task
	if rc.teamIP == "" && rc.teamID == "" {
		tasks = exploit.CreateExploitTasks(ex, allTeams, cfg.Environ, sender)
	} else {
		oneTeamMap := make(map[string]string)
		for k, v := range allTeams {
			if k == rc.teamID || v == rc.teamIP {
				oneTeamMap[k] = v
			}
		}
		tasks = exploit.CreateExploitTasks(ex, oneTeamMap, cfg.Environ, sender)
	}

	var q queue.Queue
	if ex.Endless {
		q = queue.NewEndlessQueue(rc.jobs)
	} else {
		q = queue.NewSimpleQueue(rc.jobs)
	}
	for _, t := range tasks {
		if err := q.Add(t); err != nil {
			logrus.Errorf("Failed to add task (%+v) to queue: %v", t, err)
		}
	}
	go q.Start(ctx)
	go func() {
		<-ctx.Done()
		q.Stop()
	}()

	tasksDone := 0
	for {
		select {
		case res, ok := <-q.Results():
			if !ok {
				logrus.Infof("Finished")
				return nil
			}
			tasksDone++
			logrus.Infof("Team = %v, Out = %v", res.Team, string(res.Out))
			if tasksDone == len(tasks) && !ex.Endless {
				q.Stop()
			}
		case <-ctx.Done():
			logrus.Infof("Got interrupt")
			return nil
		}
	}
}
