package neoproc

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func StartReaper(ctx context.Context) {
	logger := logrus.WithField("component", "reaper")

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

				pid, err := syscall.Wait4(-1, &wstatus, syscall.WNOHANG, nil)
				if err != nil {
					logger.Warnf("Wait4 failed: %v", err)
				} else if pid == 0 {
					break
				}

				logger.Debugf("Reaped child pid=%d, wstatus=%+v", pid, wstatus)
			}
		}
	}
}
