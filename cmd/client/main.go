package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"neo/cmd/client/cli"
	"os"

	"neo/internal/client"
)

var (
	configPath = flag.String("config", "config.yml", "yaml config file to read")
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Printf("Usage: (%[1]s run | %[1]s list | %[1]s add)\n", os.Args[0])
		os.Exit(1)
	}

	cfg, err := client.ReadConfig(*configPath)
	if err != nil {
		logrus.Fatalf("Failed to read config: %v", err)
	}

	var cmd cli.NeoCLI

	subCmd := os.Args[1]
	switch subCmd {
	case "run":
		cmd = cli.NewRun(os.Args[2:], cfg)
	case "info":
		cmd = cli.NewInfo(os.Args[2:], cfg)
	case "add":
		cmd = cli.NewAdd(os.Args[2:], cfg)
	default:
		fmt.Printf("Usage: (%[1]s run | %[1]s list | %[1]s add)\n", os.Args[0])
		os.Exit(1)
	}
	ctx := context.Background()
	if err := cmd.Run(ctx); err != nil {
		logrus.Fatalf("Error: %v", err)
	}
	logrus.Println("Finished")
}
