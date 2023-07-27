package main

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"neo/internal/logger"
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
	logger.Init()
	if err := setupConfig(); err != nil {
		logrus.Fatalf("Error setting up config: %v", err)
	}
	setConfigDefaults()
	cfg, err := readConfig()
	if err != nil {
		logrus.Fatalf("Error reading config: %v", err)
	}

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
		logrus.Fatalf("Failed to listen: %v", err)
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go srv.HeartBeat(ctx)
	go func() {
		<-ctx.Done()
		logrus.Info("Received shutdown signal, stopping server")
		s.GracefulStop()
	}()
	logrus.Infof("Starting server on port %s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}
}

func setupConfig() error {
	pflag.BoolP("debug", "v", false, "Enable verbose logging")
	pflag.StringP("config", "c", "server_config.yml", "Path to config file")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("binding flags: %w", err)
	}
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("NEO")
	viper.AutomaticEnv()
	return nil
}

func setConfigDefaults() {
	viper.SetDefault("config", "server_config.yml")
	viper.SetDefault("ping_every", time.Second*5)
	viper.SetDefault("submit_every", time.Second*2)
}

func readConfig() (*server.Config, error) {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading yaml config: %w", err)
	}
	cfg := new(server.Config)
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	logrus.Infof("Parsed config: %+v", cfg)
	return cfg, nil
}

func setLogLevel(cfg *server.Config) {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
