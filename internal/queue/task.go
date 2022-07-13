package queue

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"neo/pkg/tasklogger"
)

func NewTask(taskID, executable, dir, teamID, teamIP string, environ []string, timeout time.Duration, logger *tasklogger.TaskLogger) Task {
	return Task{
		name:       taskID,
		executable: executable,
		dir:        dir,
		teamID:     teamID,
		teamIP:     teamIP,
		timeout:    timeout,
		environ:    environ,
		logger:     logger,
	}
}

type Task struct {
	name       string
	executable string
	dir        string
	teamID     string
	teamIP     string
	environ    []string
	timeout    time.Duration
	logger     *tasklogger.TaskLogger
}

func (t Task) String() string {
	return fmt.Sprintf(
		"Exploit(path=%s, target=%s (%s), timeout=%v, environ=%+v)",
		t.executable,
		t.teamID,
		t.teamIP,
		t.timeout,
		t.environ,
	)
}

func (t Task) Command(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, t.executable, t.teamIP)
	if t.dir != "" {
		cmd.Dir = t.dir
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, t.environ...)

	// disable buffering in python scripts
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	// Disable terminal for pwntools.
	cmd.Env = append(cmd.Env, "PWNLIB_NOTERM=1")
	return cmd
}
