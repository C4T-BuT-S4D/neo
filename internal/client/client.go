package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"

	epb "github.com/c4t-but-s4d/neo/proto/go/exploits"

	fspb "github.com/c4t-but-s4d/neo/proto/go/fileserver"
	logspb "github.com/c4t-but-s4d/neo/proto/go/logs"

	"github.com/c4t-but-s4d/neo/pkg/filestream"

	"google.golang.org/grpc"
)

func New(cc grpc.ClientConnInterface, id string) *Client {
	return &Client{
		exploits: epb.NewServiceClient(cc),
		fs:       fspb.NewServiceClient(cc),
		logs:     logspb.NewServiceClient(cc),
		ID:       id,
	}
}

type Client struct {
	exploits epb.ServiceClient
	fs       fspb.ServiceClient
	logs     logspb.ServiceClient

	ID     string
	Weight int
}

func (nc *Client) GetServerState(ctx context.Context) (*epb.ServerState, error) {
	resp, err := nc.exploits.Ping(
		ctx,
		&epb.PingRequest{
			ClientId: nc.ID,
			Payload: &epb.PingRequest_ServerInfoRequest{
				ServerInfoRequest: &epb.PingRequest_ServerInfo{},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("making ping request: %w", err)
	}
	return resp.State, nil
}

func (nc *Client) Heartbeat(ctx context.Context) (*epb.ServerState, error) {
	resp, err := nc.exploits.Ping(
		ctx,
		&epb.PingRequest{
			ClientId: nc.ID,
			Payload: &epb.PingRequest_HeartbeatRequest{
				HeartbeatRequest: &epb.PingRequest_Heartbeat{
					Weight: int32(nc.Weight),
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("making ping request: %w", err)
	}
	return resp.State, nil
}

func (nc *Client) Leave(ctx context.Context) error {
	if _, err := nc.exploits.Ping(
		ctx,
		&epb.PingRequest{
			ClientId: nc.ID,
			Payload: &epb.PingRequest_LeaveRequest{
				LeaveRequest: &epb.PingRequest_Leave{},
			},
		},
	); err != nil {
		return fmt.Errorf("making ping request: %w", err)
	}
	return nil
}

func (nc *Client) Exploit(ctx context.Context, id string) (*epb.ExploitResponse, error) {
	req := &epb.ExploitRequest{
		ExploitId: id,
	}
	resp, err := nc.exploits.Exploit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("making exploit request: %w", err)
	}
	return resp, nil
}

func (nc *Client) UpdateExploit(ctx context.Context, state *epb.ExploitState) (*epb.ExploitState, error) {
	req := &epb.UpdateExploitRequest{State: state}
	resp, err := nc.exploits.UpdateExploit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("aking update exploit request: %w", err)
	}
	return resp.State, nil
}

func (nc *Client) DownloadFile(ctx context.Context, info *fspb.FileInfo, out io.Writer) error {
	resp, err := nc.fs.DownloadFile(ctx, info)
	if err != nil {
		return fmt.Errorf("making download file request: %w", err)
	}
	if err := filestream.Save(resp, out); err != nil {
		return fmt.Errorf("saving downloaded file: %w", err)
	}
	if err := resp.CloseSend(); err != nil {
		return fmt.Errorf("closing the stream: %w", err)
	}
	return nil
}

func (nc *Client) UploadFile(ctx context.Context, r io.Reader) (*fspb.FileInfo, error) {
	client, err := nc.fs.UploadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("making upload file request: %w", err)
	}
	if err := filestream.Load(r, client); err != nil {
		return nil, fmt.Errorf("loading filestream: %w", err)
	}
	fileInfo, err := client.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("closing & reading upload response: %w", err)
	}
	return fileInfo, nil
}

func (nc *Client) BroadcastCommand(ctx context.Context, command string) error {
	req := &epb.BroadcastRequest{Command: command}
	if _, err := nc.exploits.BroadcastCommand(ctx, req); err != nil {
		return fmt.Errorf("making broadcast command request: %w", err)
	}
	return nil
}

func (nc *Client) SingleRun(ctx context.Context, exploitID string) error {
	req := &epb.SingleRunRequest{ExploitId: exploitID}
	if _, err := nc.exploits.SingleRun(ctx, req); err != nil {
		return fmt.Errorf("making single run request: %w", err)
	}
	return nil
}

func (nc *Client) SetExploitDisabled(ctx context.Context, id string, disabled bool) error {
	resp, err := nc.Exploit(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching current exploit config: %w", err)
	}

	req := &epb.UpdateExploitRequest{State: resp.State}
	req.State.Disabled = disabled

	if _, err := nc.exploits.UpdateExploit(ctx, req); err != nil {
		return fmt.Errorf("making delete exploit request: %w", err)
	}
	return nil
}

func (nc *Client) ListenBroadcasts(ctx context.Context) (<-chan *epb.BroadcastSubscribeResponse, error) {
	stream, err := nc.exploits.BroadcastSubscribe(ctx, &epb.BroadcastSubscribeRequest{})
	if err != nil {
		return nil, fmt.Errorf("creating broadcast requests stream: %w", err)
	}

	results := make(chan *epb.BroadcastSubscribeResponse)
	go func() {
		defer close(results)
		for {
			cmd, err := stream.Recv()
			if !checkStreamError("broadcast", err, stream.Context().Err()) {
				return
			}
			select {
			case results <- cmd:
			case <-ctx.Done():
				logrus.Debugf("Broadcast context cancelled")
				return
			}
		}
	}()

	return results, nil
}

func (nc *Client) ListenSingleRuns(ctx context.Context) (<-chan *epb.SingleRunSubscribeResponse, error) {
	stream, err := nc.exploits.SingleRunSubscribe(ctx, &epb.SingleRunSubscribeRequest{})
	if err != nil {
		return nil, fmt.Errorf("creating single run requests stream: %w", err)
	}

	results := make(chan *epb.SingleRunSubscribeResponse)
	go func() {
		defer close(results)
		for {
			er, err := stream.Recv()
			if !checkStreamError("single runs", err, stream.Context().Err()) {
				return
			}
			select {
			case results <- er:
			case <-ctx.Done():
				logrus.Warningf("Single runs context cancelled")
				return
			}
		}
	}()

	return results, nil
}

func (nc *Client) AddLogLines(ctx context.Context, lines ...*logspb.LogLine) error {
	req := logspb.AddLogLinesRequest{Lines: lines}
	if _, err := nc.logs.AddLogLines(ctx, &req); err != nil {
		return fmt.Errorf("sending a batch of %d logs: %w", len(lines), err)
	}
	return nil
}

func (nc *Client) SearchLogLines(ctx context.Context, exploit string, version int64) (<-chan []*logspb.LogLine, error) {
	req := logspb.SearchLogLinesRequest{
		Exploit: exploit,
		Version: version,
	}
	stream, err := nc.logs.SearchLogLines(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("querying server: %w", err)
	}

	results := make(chan []*logspb.LogLine)
	go func() {
		defer close(results)
		for {
			resp, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					logrus.Errorf("Unexpected error reading log lines: %v", err)
				}
				return
			}
			select {
			case results <- resp.Lines:
			case <-ctx.Done():
				logrus.Debugf("Search logs context cancelled")
				return
			}
		}
	}()

	return results, nil
}
