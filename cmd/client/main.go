package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/c4t-but-s4d/neo/cmd/client/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan interface{}, 1)
	go func() {
		cmd.Execute(ctx)
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
