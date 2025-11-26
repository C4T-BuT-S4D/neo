package queue

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/c4t-but-s4d/neo/v2/internal/models"
	"github.com/c4t-but-s4d/neo/v2/pkg/joblogger"
)

func NewJob(
	exploit *models.Exploit,
	target *models.Target,
	executable, dir string,
	environ []string,
	timeout time.Duration,
	logger *joblogger.JobLogger,
) *Job {
	return &Job{
		Exploit: exploit,
		Target:  target,

		executable: executable,
		dir:        dir,
		timeout:    timeout,
		environ:    environ,
		logger:     logger,
	}
}

type Job struct {
	Exploit *models.Exploit
	Target  *models.Target

	executable string
	dir        string
	environ    []string
	timeout    time.Duration
	logger     *joblogger.JobLogger
}

func (t *Job) String() string {
	return fmt.Sprintf(
		"Job(exploit=%v, target=%v, path=%s, timeout=%v, environ=%+v)",
		t.Exploit,
		t.Target,
		t.executable,
		t.timeout,
		t.environ,
	)
}

func (t *Job) Command(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, t.executable, t.Target.IP)
	if t.dir != "" {
		cmd.Dir = t.dir
	}

	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			return fmt.Errorf("killing process group %d: %w", cmd.Process.Pid, err)
		}
		return nil
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, t.environ...)

	// Disable buffering in python scripts
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	// Disable terminal for pwntools.
	cmd.Env = append(cmd.Env, "PWNLIB_NOTERM=1")
	return cmd
}
