package cli

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"neo/pkg/grpc_auth"

	"neo/internal/client"

	"google.golang.org/grpc"

	"github.com/denisbrodbeck/machineid"
)

type NeoCLI interface {
	Run(ctx context.Context) error
}

type Creator func(cmd *cobra.Command, cfg *client.Config) NeoCLI

type baseCLI struct {
	c *client.Config
}

func (cmd *baseCLI) client() (*client.Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	if cmd.c.GrpcAuthKey != "" {
		interceptor := grpc_auth.NewClientInterceptor(cmd.c.GrpcAuthKey)
		opts = append(opts, grpc.WithUnaryInterceptor(interceptor.Unary()))
		opts = append(opts, grpc.WithStreamInterceptor(interceptor.Stream()))
	}
	conn, err := grpc.Dial(cmd.c.Host, opts...)
	if err != nil {
		return nil, err
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
