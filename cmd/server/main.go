package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/c4t-but-s4d/neo/internal/logger"
	"github.com/c4t-but-s4d/neo/internal/server/config"
	"github.com/c4t-but-s4d/neo/internal/server/exploit_manager"
	"github.com/c4t-but-s4d/neo/internal/server/fs"
	logs "github.com/c4t-but-s4d/neo/internal/server/logs"
	"github.com/c4t-but-s4d/neo/pkg/grpcauth"
	"github.com/c4t-but-s4d/neo/pkg/neosync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"

	empb "github.com/c4t-but-s4d/neo/proto/go/exploit_manager"
	fspb "github.com/c4t-but-s4d/neo/proto/go/fileserver"
	logspb "github.com/c4t-but-s4d/neo/proto/go/logs"
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

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fc := exploit_manager.NewFarmClient(cfg.FarmConfig)
	if err := fc.FillConfig(initCtx, &cfg.FarmConfig); err != nil {
		logrus.Fatalf("Failed to fetch config from farm: %v", err)
	}

	st, err := exploit_manager.NewBoltStorage(cfg.DBPath)
	if err != nil {
		logrus.Fatalf("Failed to create bolt storage: %v", err)
	}

	logStore, err := logs.NewLogStorage(initCtx, cfg.RedisURL)
	if err != nil {
		logrus.Fatalf("Failed to create log storage: %v", err)
	}

	if cfg.PingEvery <= 0 {
		logrus.Fatalf("ping_every should be positive")
	}
	logrus.Infof("Config: %+v", cfg)

	exploitsServer := exploit_manager.New(cfg, st)
	fsServer, err := fs.New(cfg)
	if err != nil {
		logrus.Fatalf("Failed to create file server: %v", err)
	}
	logsServer := logs.New(logStore)

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
	empb.RegisterServiceServer(s, exploitsServer)
	fspb.RegisterServiceServer(s, fsServer)
	logspb.RegisterServiceServer(s, logsServer)
	reflection.Register(s)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logrus.Infof("Starting metrics server on %s", viper.GetString("metrics.address"))
		if err := http.ListenAndServe(viper.GetString("metrics.address"), http.DefaultServeMux); err != nil {
			logrus.Fatalf("Failed to serve metrics: %v", err)
		}
	}()

	runCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	wg := neosync.NewWG()

	wg.Add(3)
	go func() {
		defer wg.Done()
		exploitsServer.HeartBeat(runCtx)
	}()
	go func() {
		defer wg.Done()
		exploitsServer.UpdateMetrics(runCtx)
	}()
	go func() {
		defer wg.Done()
		<-runCtx.Done()
		logrus.Info("Received shutdown signal, stopping server")
		s.GracefulStop()
	}()

	logrus.Infof("Starting server on port %s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}
	select {
	case <-wg.Await():
		logrus.Info("Shutdown finished")
	case <-time.After(10 * time.Second):
		logrus.Warn("Shutdown timeout")
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
	viper.SetDefault("metrics.address", ":3000")
}

func readConfig() (*config.Config, error) {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading yaml config: %w", err)
	}
	cfg := new(config.Config)
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	logrus.Infof("Parsed config: %+v", cfg)
	return cfg, nil
}

func setLogLevel(cfg *config.Config) {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
