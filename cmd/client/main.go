package main

import (
	"context"
	"neo/cmd/client/cmd"
	"os"
	"os/signal"
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
	<-c
	cancel()
	<-done
}
