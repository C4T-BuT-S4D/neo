package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/c4t-but-s4d/neo/internal/logger"
	"github.com/c4t-but-s4d/neo/pkg/neoproc"
)

func main() {
	logger.Init()

	log := logrus.WithFields(logrus.Fields{
		"pid":       os.Getpid(),
		"component": "reaper-runner",
	})

	if err := neoproc.SetSubreaper(); err != nil {
		log.Fatalf("Failed to set subreaper: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("Starting reaper")
	neoproc.StartReaper(ctx)
	log.Info("Reaper finished")
}
