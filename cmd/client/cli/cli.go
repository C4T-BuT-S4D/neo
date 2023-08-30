package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/c4t-but-s4d/neo/internal/client"
	"github.com/c4t-but-s4d/neo/pkg/grpcauth"
)

type NeoCLI interface {
	Run(ctx context.Context) error
}

type baseCLI struct {
	cfg *client.Config

	clientID string
}

func (cmd *baseCLI) client() (*client.Client, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor(gzip.Name),
		),
	}
	if cmd.cfg.GrpcAuthKey != "" {
		interceptor := grpcauth.NewClientInterceptor(cmd.cfg.GrpcAuthKey)
		opts = append(
			opts,
			grpc.WithUnaryInterceptor(interceptor.Unary()),
			grpc.WithStreamInterceptor(interceptor.Stream()),
		)
	}
	if !cmd.cfg.UseTLS {
		opts = append(
			opts,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}
	conn, err := grpc.Dial(cmd.cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing grpc: %w", err)
	}
	return client.New(conn, cmd.ClientID()), nil
}

func (cmd *baseCLI) ClientID() string {
	if cmd.clientID != "" {
		return cmd.clientID
	}

	cmd.clientID = viper.GetString("client_id")
	if cmd.clientID == "" {
		var err error
		if cmd.clientID, err = machineid.ID(); err != nil {
			logrus.Fatalf("Failed to get unique client name: %v", err)
		}
	}
	logrus.Infof("Detected client id: %s", cmd.clientID)
	return cmd.clientID
}

func (cmd *baseCLI) Run(_ context.Context) error {
	return errors.New("unimplemented")
}
