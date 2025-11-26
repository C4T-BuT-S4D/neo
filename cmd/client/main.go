package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/cmd/client/cli"
	"github.com/c4t-but-s4d/neo/v2/cmd/client/cmd"
	"github.com/c4t-but-s4d/neo/v2/pkg/logging"
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func main() {
	logSync := logging.Init("info")

	cobra.EnableTraverseRunHooks = true

	v, err := viperext.NewViper("NEO")
	if err != nil {
		zap.L().Fatal("Error creating viper", zap.Error(err))
	}

	rootCmd := &cobra.Command{
		Use:              "neo",
		Short:            "Neo client",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			logSync.Close()
		},
	}

	rootCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringP("config", "c", "client_config.yml", "config file")
	rootCmd.Flags().String("host", "127.0.0.1:5005", "server host")

	rootCmd.PersistentPreRun = func(*cobra.Command, []string) {
		viperext.RunBindCommandFlags(v, rootCmd)

		v.SetConfigFile(v.GetString("config"))
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err == nil {
			zap.L().Info("Using config file", zap.String("path", v.ConfigFileUsed()))
		}

		logging.SetLogLevel(v.GetString("log_level"))
	}

	cc := cli.NewContext(v)

	rootCmd.AddCommand(
		cmd.NewRunCommand(cc),
		cmd.NewDryRunCommand(cc),
		cmd.NewInfoCommand(cc),
		cmd.NewAddCommand(cc),
		cmd.NewUpdateCommand(cc),
		cmd.NewEnableCommand(cc),
		cmd.NewDisableCommand(cc),
		cmd.NewSingleCommand(cc),
		cmd.NewBroadcastCommand(cc),
		cmd.NewTailCommand(cc),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		zap.L().Fatal("Running app", zap.Error(err))
	}
}
