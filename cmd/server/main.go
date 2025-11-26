package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
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
	"github.com/c4t-but-s4d/neo/v2/pkg/viperext"
)

func main() {
	if err := run(); err != nil {
		zap.L().Fatal("Running server", zap.Error(err))
	}
}

func loadConfig() (*config.Config, error) {
	pflag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pflag.StringP("config", "c", "server_config.yml", "Path to config file")
	pflag.String("address", ":5005", "Server address")
	pflag.String("metrics-address", ":3000", "Metrics server address")
	pflag.Parse()

	v, err := viperext.NewViper("NEO")
	if err != nil {
		return nil, fmt.Errorf("creating viper: %w", err)
	}

	viperext.BindPFlags(v, pflag.CommandLine)

	v.SetConfigFile(v.GetString("config"))
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg, err := viperext.GetConfig(v, &config.Config{})
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	defer logging.Init(cfg.LogLevel).Close()

	zap.L().Info("Config loaded", zap.Any("config", cfg))

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fc := exploits.NewFarmClient(cfg.FarmConfig)
	if err := fc.FillConfig(initCtx, &cfg.FarmConfig); err != nil {
		return fmt.Errorf("fetching config from farm: %w", err)
	}

	st, err := exploits.NewBoltStorage(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("creating bolt storage: %w", err)
	}

	zap.L().Info("Using VictoriaLogs storage", zap.String("url", cfg.VictoriaLogsURL))
	logStore, err := logstor.NewVictoriaLogsStorage(cfg.VictoriaLogsURL)
	if err != nil {
		return fmt.Errorf("creating victorialogs storage: %w", err)
	}

	if cfg.PingEvery <= 0 {
		return errors.New("ping_every should be positive")
	}
	if cfg.SubmitEvery <= 0 {
		return errors.New("submit_every should be positive")
	}

	exploitsServer := exploits.New(cfg, st)
	fsServer, err := fs.New(cfg)
	if err != nil {
		return fmt.Errorf("creating file server: %w", err)
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
	httpMux.Handle("/api/metrics/", serverMetrics.NewProxyHandler(http.DefaultClient, cfg.VictoriaMetricsURL, cfg.GrpcAuthKey))

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

	var wg sync.WaitGroup
	wg.Go(func() {
		exploitsServer.HeartBeat(runCtx)
	})
	wg.Go(func() {
		exploitsServer.UpdateMetrics(runCtx)
	})
	wg.Go(func() {
		zap.L().Info("Starting metrics server", zap.String("address", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Failed to serve metrics", zap.Error(err))
		}
	})
	wg.Go(func() {
		zap.L().Info("Starting multiproto server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Failed to serve http", zap.Error(err))
		}
	})

	<-runCtx.Done()
	zap.L().Info("Received shutdown signal, stopping servers")

	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer shutdownCancel()
	shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, 5*time.Second)
	defer shutdownCancel()

	var finalErr error
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		finalErr = errors.Join(finalErr, fmt.Errorf("shutting down http server: %w", err))
	}
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		finalErr = errors.Join(finalErr, fmt.Errorf("shutting down metrics server: %w", err))
	}

	select {
	case <-neosync.AwaitWG(&wg):
		zap.L().Info("Shutdown finished")
	case <-time.After(10 * time.Second):
		finalErr = errors.Join(finalErr, errors.New("shutdown timeout"))
	}

	return finalErr
}
