package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan any, 1)
	go func() {
		if err := cmd.Execute(ctx); err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		done <- nil
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		cancel()
		<-done
	case <-done:
	}
}
