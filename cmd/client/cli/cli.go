package cli

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/pkg/grpcauth"
)

type NeoCLI interface {
	Run(ctx context.Context) error
}

type baseCLI struct {
	cfg *client.Config

	clientID string
}

func (cmd *baseCLI) client() (*client.Client, error) {
	clientID, err := cmd.ClientID()
	if err != nil {
		return nil, fmt.Errorf("getting client id: %w", err)
	}

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
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	conn, err := grpc.NewClient(cmd.cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing grpc: %w", err)
	}
	return client.New(conn, clientID), nil
}

func (cmd *baseCLI) ClientID() (string, error) {
	if cmd.clientID != "" {
		return cmd.clientID, nil
	}

	cmd.clientID = viper.GetString("client_id")
	if cmd.clientID == "" {
		var err error
		if cmd.clientID, err = machineid.ID(); err != nil {
			return "", fmt.Errorf("getting unique client name: %w", err)
		}
	}
	zap.L().Info("Detected client id", zap.String("client_id", cmd.clientID))
	return cmd.clientID, nil
}

func (cmd *baseCLI) Run(_ context.Context) error {
	return errors.New("unimplemented")
}
