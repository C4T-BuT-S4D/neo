package testutils

import (
	"context"
	"time"
)

func LessString(v1, v2 string) bool {
	return v1 < v2
}

func CanceledContext() context.Context {
	ctx, c := context.WithCancel(context.Background())
	defer c()
	return ctx
}

func TimeoutedContext() context.Context {
	ctx, c := context.WithTimeout(context.Background(), time.Second*0)
	defer c()
	return ctx
}
