package cmd

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Neo client",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yml)")
	rootCmd.PersistentFlags().String("host", "", "server host")
	rootCmd.PersistentFlags().BoolP("verbose", "v", true, "enable debug logging (default true)")

	mustBindPersistent(rootCmd, "host")
	mustBindPersistent(rootCmd, "verbose")
}

func mustBindPersistent(c *cobra.Command, flag string) {
	if err := viper.BindPFlag(flag, c.PersistentFlags().Lookup(flag)); err != nil {
		logrus.Fatalf("Error binding flag %s: %v", flag, err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("NEO")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file:", viper.ConfigFileUsed())
	}

	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Debugf("Got configuration: %+v", viper.AllSettings())
}
