package cli

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"

	"neo/internal/client"

	"google.golang.org/grpc"

	"github.com/denisbrodbeck/machineid"
)

type NeoCLI interface {
	Run(ctx context.Context) error
}

type baseCLI struct {
	c *client.Config
}

func (cmd *baseCLI) client() (*client.Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
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

func (cmd *baseCLI) Run(ctx context.Context) error {
	return errors.New("unimplemented")
}
