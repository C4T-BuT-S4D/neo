package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	child, cf := context.WithCancel(ctx)
	fmt.Println(child.Err(), ctx.Err())
	cf()
	fmt.Println(child.Err(), ctx.Err())
}
