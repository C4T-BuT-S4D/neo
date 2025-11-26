package logs

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/c4t-but-s4d/neo/v2/internal/logstor"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

const bufSize = 1024 * 1024

func testServerWithClient(t *testing.T, storage *logstor.MockStorage) (*Server, logspb.ServiceClient) {
	t.Helper()

	s := New(storage)

	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	logspb.RegisterServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server exited: %v", err)
		}
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		lis.Close()
	})

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		conn.Close()
	})

	client := logspb.NewServiceClient(conn)
	return s, client
}

func TestNew(t *testing.T) {
	storage := logstor.NewMockStorage()
	s := New(storage)
	require.NotNil(t, s)
}

func TestServer_AddLogLines(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	lines := []*logspb.LogLine{
		{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "test message 1",
			Level:     "info",
			Team:      "team1",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "test message 2",
			Level:     "error",
			Team:      "team2",
			Timestamp: timestamppb.New(time.Now()),
		},
	}

	_, err := client.AddLogLines(ctx, &logspb.AddLogLinesRequest{Lines: lines})
	require.NoError(t, err)

	// Verify lines were stored
	stored := storage.Lines()
	require.Len(t, stored, 2)
	require.Equal(t, "test message 1", stored[0].GetMessage())
	require.Equal(t, "test message 2", stored[1].GetMessage())
}

func TestServer_AddLogLines_Empty(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	_, err := client.AddLogLines(ctx, &logspb.AddLogLinesRequest{Lines: nil})
	require.NoError(t, err)

	require.Empty(t, storage.Lines())
}

func TestServer_AddLogLines_Error(t *testing.T) {
	storage := logstor.NewMockStorage()
	storage.AddErr = errors.New("storage error")
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	lines := []*logspb.LogLine{
		{
			Exploit: "exploit1",
			Message: "test",
		},
	}

	_, err := client.AddLogLines(ctx, &logspb.AddLogLinesRequest{Lines: lines})
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestServer_SearchLogLines(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	// Add some lines first
	lines := []*logspb.LogLine{
		{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "message 1",
			Level:     "info",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "message 2",
			Level:     "error",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "exploit2",
			Version:   1,
			Message:   "message 3",
			Level:     "info",
			Timestamp: timestamppb.New(time.Now()),
		},
	}
	require.NoError(t, storage.Add(ctx, lines...))

	// Search for all lines from exploit1
	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "exploit1",
	})
	require.NoError(t, err)

	var results []*logspb.LogLine
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		results = append(results, resp.GetLines()...)
	}

	require.Len(t, results, 2)
	require.Equal(t, "message 1", results[0].GetMessage())
	require.Equal(t, "message 2", results[1].GetMessage())
}

func TestServer_SearchLogLines_ByVersion(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	lines := []*logspb.LogLine{
		{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "version 1 message",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "exploit1",
			Version:   2,
			Message:   "version 2 message",
			Timestamp: timestamppb.New(time.Now()),
		},
	}
	require.NoError(t, storage.Add(ctx, lines...))

	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "exploit1",
		Version: 2,
	})
	require.NoError(t, err)

	var results []*logspb.LogLine
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		results = append(results, resp.GetLines()...)
	}

	require.Len(t, results, 1)
	require.Equal(t, "version 2 message", results[0].GetMessage())
}

func TestServer_SearchLogLines_Empty(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "nonexistent",
	})
	require.NoError(t, err)

	var results []*logspb.LogLine
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		results = append(results, resp.GetLines()...)
	}

	require.Empty(t, results)
}

func TestServer_SearchLogLines_Error(t *testing.T) {
	storage := logstor.NewMockStorage()
	storage.SearchErr = errors.New("search error")
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "exploit1",
	})
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestServer_AddAndSearch(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	// Add via gRPC
	lines := []*logspb.LogLine{
		{
			Exploit:   "test-exploit",
			Version:   42,
			Message:   "hello world",
			Level:     "debug",
			Team:      "team-a",
			Timestamp: timestamppb.New(time.Now()),
		},
	}

	_, err := client.AddLogLines(ctx, &logspb.AddLogLinesRequest{Lines: lines})
	require.NoError(t, err)

	// Search via gRPC
	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "test-exploit",
		Version: 42,
	})
	require.NoError(t, err)

	var results []*logspb.LogLine
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		results = append(results, resp.GetLines()...)
	}

	require.Len(t, results, 1)
	require.Equal(t, "hello world", results[0].GetMessage())
	require.Equal(t, "test-exploit", results[0].GetExploit())
	require.Equal(t, int64(42), results[0].GetVersion())
}

func TestServer_SearchLogLines_WithLimit(t *testing.T) {
	storage := logstor.NewMockStorage()
	_, client := testServerWithClient(t, storage)
	ctx := context.Background()

	// Add many lines
	var lines []*logspb.LogLine
	for i := 0; i < 10; i++ {
		lines = append(lines, &logspb.LogLine{
			Exploit:   "exploit1",
			Version:   1,
			Message:   "message",
			Timestamp: timestamppb.New(time.Now()),
		})
	}
	require.NoError(t, storage.Add(ctx, lines...))

	stream, err := client.SearchLogLines(ctx, &logspb.SearchLogLinesRequest{
		Exploit: "exploit1",
		Limit:   5,
	})
	require.NoError(t, err)

	var results []*logspb.LogLine
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		results = append(results, resp.GetLines()...)
	}

	require.Len(t, results, 5)
}
