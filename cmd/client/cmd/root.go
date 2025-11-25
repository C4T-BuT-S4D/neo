package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/c4t-but-s4d/neo/v2/pkg/logging"
)

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Neo client",
}

func Execute(ctx context.Context) error {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("executing root command: %w", err)
	}
	return nil
}

//nolint:gochecknoinits // cli init
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("config", "c", "client_config.yml", "config file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", true, "enable debug logging")
	rootCmd.PersistentFlags().String("host", "127.0.0.1:5005", "server host")

	lo.Must0(viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")))
	lo.Must0(viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host")))
	lo.Must0(viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")))
}

func initConfig() {
	logging.Init(viper.GetBool("verbose"))

	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("NEO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		zap.L().Info("Using config file", zap.String("config", viper.ConfigFileUsed()))
	}

	zap.L().Debug("Got configuration", zap.Any("settings", viper.AllSettings()))
}
