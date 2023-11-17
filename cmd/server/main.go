package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"

	"github.com/c4t-but-s4d/neo/v2/internal/logger"
	"github.com/c4t-but-s4d/neo/v2/internal/server/config"
	"github.com/c4t-but-s4d/neo/v2/internal/server/exploits"
	"github.com/c4t-but-s4d/neo/v2/internal/server/fs"
	logs "github.com/c4t-but-s4d/neo/v2/internal/server/logs"
	"github.com/c4t-but-s4d/neo/v2/pkg/grpcauth"
	"github.com/c4t-but-s4d/neo/v2/pkg/mu"
	"github.com/c4t-but-s4d/neo/v2/pkg/neohttp"
	"github.com/c4t-but-s4d/neo/v2/pkg/neosync"
	epb "github.com/c4t-but-s4d/neo/v2/proto/go/exploits"
	fspb "github.com/c4t-but-s4d/neo/v2/proto/go/fileserver"
	logspb "github.com/c4t-but-s4d/neo/v2/proto/go/logs"
)

func main() {
	logger.Init()
	if err := setupConfig(); err != nil {
		logrus.Fatalf("Error setting up config: %v", err)
	}

	cfg, err := readConfig()
	if err != nil {
		logrus.Fatalf("Error reading config: %v", err)
	}

	setLogLevel(cfg)

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fc := exploits.NewFarmClient(cfg.FarmConfig)
	if err := fc.FillConfig(initCtx, &cfg.FarmConfig); err != nil {
		logrus.Fatalf("Failed to fetch config from farm: %v", err)
	}

	st, err := exploits.NewBoltStorage(cfg.DBPath)
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

	exploitsServer := exploits.New(cfg, st)
	fsServer, err := fs.New(cfg)
	if err != nil {
		logrus.Fatalf("Failed to create file server: %v", err)
	}
	logsServer := logs.New(logStore)

	var opts []grpc.ServerOption
	if cfg.GrpcAuthKey != "" {
		authInterceptor := grpcauth.NewServerInterceptor(cfg.GrpcAuthKey)
		opts = append(opts, grpc.UnaryInterceptor(authInterceptor.Unary()))
		opts = append(opts, grpc.StreamInterceptor(authInterceptor.Stream()))
	}

	s := grpc.NewServer(opts...)
	epb.RegisterServiceServer(s, exploitsServer)
	fspb.RegisterServiceServer(s, fsServer)
	logspb.RegisterServiceServer(s, logsServer)
	reflection.Register(s)

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/", neohttp.StaticHandler(cfg.StaticDir))

	muHandler := mu.NewHandler(s, mu.WithHTTPHandler(httpMux))
	httpServer := &http.Server{
		Handler: muHandler,
		Addr:    cfg.Address,
	}

	runCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	wg := sync.WaitGroup{}

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

		shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer shutdownCancel()
		shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	logrus.Infof("Starting multiproto server on %s", cfg.Address)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Fatalf("Failed to serve: %v", err)
	}

	select {
	case <-neosync.AwaitWG(&wg):
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

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("NEO")
	viper.AutomaticEnv()

	viper.MustBindEnv("grpc_auth_key")
	viper.MustBindEnv("farm.password")
	viper.MustBindEnv("farm.url")
	viper.MustBindEnv("db_path")
	viper.MustBindEnv("redis_url")
	viper.MustBindEnv("base_dir")

	viper.SetDefault("config", "server_config.yml")
	viper.SetDefault("ping_every", time.Second*5)
	viper.SetDefault("submit_every", time.Second*2)
	viper.SetDefault("address", ":5005")
	viper.SetDefault("static_dir", "front/dist")
	viper.SetDefault("redis_url", "redis://127.0.0.1:6379/0")
	viper.SetDefault("db_path", "data/db.db")
	viper.SetDefault("base_dir", "data/exploits")

	return nil
}

func readConfig() (*config.Config, error) {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading yaml config: %w", err)
	}

	cfg := &config.Config{}
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
