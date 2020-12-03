package client

import (
	"context"
	"io"

	"neo/pkg/filestream"

	"google.golang.org/grpc"

	neopb "neo/lib/genproto/neo"
)

func New(cc grpc.ClientConnInterface, id string) *Client {
	return &Client{
		c:  neopb.NewExploitManagerClient(cc),
		ID: id,
	}
}

type Client struct {
	c      neopb.ExploitManagerClient
	ID     string
	Weight int
}

func (nc *Client) Ping(ctx context.Context, t neopb.PingRequest_PingType) (*neopb.ServerState, error) {
	req := &neopb.PingRequest{ClientId: nc.ID, Type: t}
	if t == neopb.PingRequest_HEARTBEAT {
		req.Type = neopb.PingRequest_HEARTBEAT
		req.Weight = int32(nc.Weight)
	}
	resp, err := nc.c.Ping(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetState(), nil
}

func (nc *Client) ExploitConfig(ctx context.Context, id string) (*neopb.ExploitConfiguration, error) {
	req := &neopb.ExploitRequest{
		ExploitId: id,
	}
	resp, err := nc.c.Exploit(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetConfig(), nil
}

func (nc *Client) UpdateExploit(ctx context.Context, req *neopb.UpdateExploitRequest) error {
	_, err := nc.c.UpdateExploit(ctx, req)
	return err
}

func (nc *Client) DownloadFile(ctx context.Context, info *neopb.FileInfo, out io.Writer) error {
	resp, err := nc.c.DownloadFile(ctx, info)
	if err != nil {
		return err
	}
	if err := filestream.Save(resp, out); err != nil {
		return err
	}
	return resp.CloseSend()
}

func (nc *Client) UploadFile(ctx context.Context, r io.Reader) (*neopb.FileInfo, error) {
	client, err := nc.c.UploadFile(ctx)
	if err != nil {
		return nil, err
	}
	if err := filestream.Load(r, client); err != nil {
		return nil, err
	}
	return client.CloseAndRecv()
}
