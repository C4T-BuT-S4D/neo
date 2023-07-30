package cli

import (
	"context"
	"errors"
	"fmt"

	"neo/internal/client"
	"neo/pkg/grpcauth"

	"github.com/denisbrodbeck/machineid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/encoding/gzip"

	"google.golang.org/grpc"
)

type NeoCLI interface {
	Run(ctx context.Context) error
}

type baseCLI struct {
	c *client.Config
}

func (cmd *baseCLI) client() (*client.Client, error) {
	var opts []grpc.DialOption
	opts = append(
		opts,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor(gzip.Name),
		),
	)
	if cmd.c.GrpcAuthKey != "" {
		interceptor := grpcauth.NewClientInterceptor(cmd.c.GrpcAuthKey)
		opts = append(
			opts,
			grpc.WithUnaryInterceptor(interceptor.Unary()),
			grpc.WithStreamInterceptor(interceptor.Stream()),
		)
	}
	conn, err := grpc.Dial(cmd.c.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing grpc: %w", err)
	}
	return client.New(conn, cmd.ClientID()), nil
}

func (cmd *baseCLI) ClientID() string {
	id, err := machineid.ID()
	if err != nil {
		logrus.Fatalf("Failed to get unique client name: %v", err)
	}
	return id
}

func (cmd *baseCLI) Run(_ context.Context) error {
	return errors.New("unimplemented")
}
