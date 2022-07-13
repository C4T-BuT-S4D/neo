package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"neo/internal/server"
	"neo/pkg/grpcauth"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"

	neopb "neo/lib/genproto/neo"
)

func main() {
	setupConfig()
	initLogger()
	setConfigDefaults()
	cfg := readConfig()

	setLogLevel(cfg)

	ctx := context.Background()
	fc := server.NewFarmClient(cfg.FarmConfig)
	if err := fc.FillConfig(ctx, &cfg.FarmConfig); err != nil {
		logrus.Fatalf("Failed to fetch config from farm: %v", err)
	}

	st, err := server.NewBoltStorage(cfg.DBPath)
	if err != nil {
		logrus.Fatalf("Failed to create bolt storage: %v", err)
	}

	logStore, err := server.NewLogStorage(ctx, cfg.RedisURL)
	if err != nil {
		logrus.Fatalf("Failed to create log storage: %v", err)
	}

	if cfg.PingEvery <= 0 {
		logrus.Fatalf("ping_every should be positive")
	}
	logrus.Infof("Config: %+v", cfg)
	srv, err := server.New(cfg, st, logStore)
	if err != nil {
		logrus.Fatalf("Failed to create server: %v", err)
	}
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if cfg.GrpcAuthKey != "" {
		authInterceptor := grpcauth.NewServerInterceptor(cfg.GrpcAuthKey)
		opts = append(opts, grpc.UnaryInterceptor(authInterceptor.Unary()))
		opts = append(opts, grpc.StreamInterceptor(authInterceptor.Stream()))
	}

	s := grpc.NewServer(opts...)
	neopb.RegisterExploitManagerServer(s, srv)
	reflection.Register(s)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go srv.HeartBeat(ctx)
	go func() {
		<-c
		logrus.Info("Received shutdown signal, stopping server")
		cancel()
		s.GracefulStop()
	}()
	logrus.Infof("Starting server on port %s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}

func setupConfig() {
	pflag.BoolP("debug", "v", false, "Enable verbose logging")
	pflag.StringP("config", "c", "server_config.yml", "Path to config file")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		logrus.Fatalf("Error binding flags: %v", err)
	}
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("NEO")
	viper.AutomaticEnv()
}

func setConfigDefaults() {
	viper.SetDefault("config", "server_config.yml")
	viper.SetDefault("ping_every", time.Second*5)
	viper.SetDefault("submit_every", time.Second*2)
}

func readConfig() *server.Config {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("Error reading config from yaml: ", err)
	}
	cfg := new(server.Config)
	if err := viper.Unmarshal(cfg); err != nil {
		logrus.Fatal("Error parsing config: ", err)
	}
	logrus.Infof("Parsed config: %+v", cfg)
	return cfg
}

func initLogger() {
	mainFormatter := &logrus.TextFormatter{}
	mainFormatter.FullTimestamp = true
	mainFormatter.ForceColors = true
	mainFormatter.PadLevelText = true
	mainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(mainFormatter)
}

func setLogLevel(cfg *server.Config) {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
