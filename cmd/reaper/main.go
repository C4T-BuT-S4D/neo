package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"neo/internal/logger"
	"neo/pkg/neoproc"

	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()

	log := logrus.WithFields(logrus.Fields{
		"pid":       os.Getpid(),
		"component": "reaper-runner",
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("Starting reaper")
	neoproc.StartReaper(ctx)
	log.Info("Reaper finished")
}
