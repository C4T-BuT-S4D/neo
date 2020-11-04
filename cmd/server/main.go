package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	"neo/internal/server"
	"neo/pkg/grpc_auth"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	neopb "neo/lib/genproto/neo"
)

var (
	configFile = flag.String("config", "config.yml", "yaml config file to read")
)

func load(path string, cfg *server.Configuration) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return server.ReadConfig(data, cfg)
}

func watchConfig(ctx context.Context, srv *server.ExploitManagerServer) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				watcher.Close()
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
					cfg := &server.Configuration{}
					err := load(*configFile, cfg)
					if err != nil {
						logrus.Errorf("Failed to reload read configuration: %v", err)
					} else {
						logrus.Infof("Reloaded config: %v", cfg)
						srv.UpdateConfig(cfg)
					}
				}
			}
		}
	}()
	return watcher.Add(*configFile)
}

func main() {
	flag.Parse()
	cfg := &server.Configuration{}
	if err := load(*configFile, cfg); err != nil {
		logrus.Fatalf("Failed to read config: %v", err)
	}
	st, err := server.NewBoltStorage(cfg.DBPath)
	if err != nil {
		logrus.Fatalf("Failed to create bolt storage: %v", err)
	}
	if cfg.RunEvery <= 0 {
		logrus.Fatalf("run_every should be positive")
	}
	if cfg.PingEvery <= 0 {
		logrus.Fatalf("ping_every should be positive")
	}
	if cfg.Timeout <= 0 {
		logrus.Fatalf("timeout should be positive")
	}
	logrus.Infof("Config: %+v", cfg)
	srv := server.New(cfg, st)
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if cfg.GrpcAuthKey != "" {
		authInterceptor := grpc_auth.NewServerInterceptor(cfg.GrpcAuthKey)
		opts = append(opts, grpc.UnaryInterceptor(authInterceptor.Unary()))
		opts = append(opts, grpc.StreamInterceptor(authInterceptor.Stream()))
	}
	s := grpc.NewServer(opts...)
	neopb.RegisterExploitManagerServer(s, srv)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	if err := watchConfig(ctx, srv); err != nil {
		logrus.Errorf("Failed to start config auto-reload: %v", err)
	}
	go srv.HeartBeat(ctx)
	go func() {
		<-c
		cancel()
		s.GracefulStop()
	}()
	logrus.Infof("Starting server on port %s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}
