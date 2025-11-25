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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "google.golang.org/grpc/encoding/gzip"

	"github.com/c4t-but-s4d/neo/v2/internal/logstor"
	"github.com/c4t-but-s4d/neo/v2/internal/server/config"
	"github.com/c4t-but-s4d/neo/v2/internal/server/exploits"
	"github.com/c4t-but-s4d/neo/v2/internal/server/fs"
	"github.com/c4t-but-s4d/neo/v2/internal/server/logs"
	serverMetrics "github.com/c4t-but-s4d/neo/v2/internal/server/metrics"
	"github.com/c4t-but-s4d/neo/v2/pkg/grpcauth"
	"github.com/c4t-but-s4d/neo/v2/pkg/logging"
	"github.com/c4t-but-s4d/neo/v2/pkg/mu"
	"github.com/c4t-but-s4d/neo/v2/pkg/neohttp"
	"github.com/c4t-but-s4d/neo/v2/pkg/neosync"
	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
	fspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/fileserver"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

func main() {
	if err := setupConfig(); err != nil {
		// Can't use zap yet, not initialized.
		panic(fmt.Sprintf("Error setting up config: %v", err))
	}

	cfg, err := readConfig()
	if err != nil {
		panic(fmt.Sprintf("Error reading config: %v", err))
	}

	logging.Init(cfg.Debug)
	defer logging.Sync()

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fc := exploits.NewFarmClient(cfg.FarmConfig)
	if err := fc.FillConfig(initCtx, &cfg.FarmConfig); err != nil {
		zap.L().Fatal("Failed to fetch config from farm", zap.Error(err))
	}

	st, err := exploits.NewBoltStorage(cfg.DBPath)
	if err != nil {
		zap.L().Fatal("Failed to create bolt storage", zap.Error(err))
	}

	zap.L().Info("Using VictoriaLogs storage", zap.String("url", cfg.VictoriaLogsURL))
	logStore, err := logstor.NewVictoriaLogsStorage(initCtx, cfg.VictoriaLogsURL)
	if err != nil {
		zap.L().Fatal("Failed to create victorialogs storage", zap.Error(err))
	}

	if cfg.PingEvery <= 0 {
		zap.L().Fatal("ping_every should be positive")
	}
	zap.L().Info("Config loaded", zap.Any("config", cfg))

	exploitsServer := exploits.New(cfg, st)
	fsServer, err := fs.New(cfg)
	if err != nil {
		zap.L().Fatal("Failed to create file server", zap.Error(err))
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
	httpMux.Handle("/", neohttp.StaticHandler(cfg.StaticDir))
	httpMux.Handle("/api/metrics/", serverMetrics.NewProxyHandler("http://victoria:8428", cfg.GrpcAuthKey))

	muHandler := mu.NewHandler(s, mu.WithHTTPHandler(httpMux))
	httpServer := &http.Server{
		Handler: muHandler,
		Addr:    cfg.Address,
	}

	// Separate server to make it private.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Handler: metricsMux,
		Addr:    cfg.MetricsAddress,
	}

	runCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Go(func() {
		exploitsServer.HeartBeat(runCtx)
	})
	wg.Go(func() {
		exploitsServer.UpdateMetrics(runCtx)
	})
	wg.Go(func() {
		<-runCtx.Done()
		zap.L().Info("Received shutdown signal, stopping server")

		shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer shutdownCancel()
		shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, 5*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("Failed to shutdown http server", zap.Error(err))
		}
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("Failed to shutdown metrics server", zap.Error(err))
		}
	})
	wg.Go(func() {
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Failed to serve metrics", zap.Error(err))
		}
	})

	zap.L().Info("Starting multiproto server", zap.String("address", cfg.Address))
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.L().Fatal("Failed to serve", zap.Error(err))
	}

	select {
	case <-neosync.AwaitWG(&wg):
		zap.L().Info("Shutdown finished")
	case <-time.After(10 * time.Second):
		zap.L().Warn("Shutdown timeout")
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
	viper.MustBindEnv("victorialogs_url")
	viper.MustBindEnv("base_dir")

	viper.SetDefault("config", "server_config.yml")
	viper.SetDefault("ping_every", time.Second*5)
	viper.SetDefault("submit_every", time.Second*2)
	viper.SetDefault("address", ":5005")
	viper.SetDefault("metrics_address", ":3000")
	viper.SetDefault("static_dir", "front/dist")
	viper.SetDefault("victorialogs_url", "http://127.0.0.1:9428")
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

	return cfg, nil
}
