package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"neo/pkg/grpc_auth"
	"net"
	"os"
	"os/signal"

	"neo/internal/server"

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

func main() {
	flag.Parse()
	cfg := &server.Configuration{}
	if err := load(*configFile, cfg); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	st, err := server.NewBoltStorage(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to create bolt storage: %v", err)
	}
	if cfg.RunEvery <= 0 {
		log.Fatalf("run_every should be positive")
	}
	if cfg.PingEvery <= 0 {
		log.Fatalf("ping_every should be positive")
	}
	if cfg.Timeout <= 0 {
		log.Fatalf("timeout should be positive")
	}
	logrus.Infof("Config: %+v", cfg)
	srv := server.New(cfg, st)
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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

	go srv.HeartBeat(ctx)
	go func() {
		<-c
		cancel()
		s.GracefulStop()
	}()
	logrus.Infof("Starting server on port %s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
