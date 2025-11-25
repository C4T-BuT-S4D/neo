package main

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cmd"
	"github.com/c4t-but-s4d/neo/v2/pkg/logging"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute(ctx)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var err error
	select {
	case <-c:
		cancel()
		err = <-done
	case err = <-done:
	}

	if err != nil {
		zap.L().Error("Command failed", zap.Error(err))
		logging.Sync()
		os.Exit(1)
	}
	logging.Sync()
}
