package neoproc

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func StartReaper(ctx context.Context) {
	logger := logrus.WithField("component", "reaper")

	if os.Getpid() != 1 {
		logger.Warn("Not PID 1, not reaping children")
		return
	}

	c := make(chan os.Signal, 5)
	signal.Notify(c, syscall.SIGCHLD)
	notify := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-c:
				logger.Debugf("Received signal %v", sig)
				select {
				case notify <- struct{}{}:
				default:
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-notify:
			for {
				var wstatus syscall.WaitStatus

				pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
				for errors.Is(err, syscall.EINTR) {
					pid, err = syscall.Wait4(pid, &wstatus, 0, nil)
				}
				if errors.Is(err, syscall.ECHILD) {
					break
				}

				logger.Debugf("Reaped child pid=%d, wstatus=%+v", pid, wstatus)
			}
		}
	}
}
