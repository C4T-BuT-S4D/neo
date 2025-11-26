package fs

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/c4t-but-s4d/neo/v2/internal/server/config"
	"github.com/c4t-but-s4d/neo/v2/pkg/filestream"
	fspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/fileserver"
)

const bufSize = 1024 * 1024

func testServerWithClient(t *testing.T) (*Server, fspb.ServiceClient) {
	t.Helper()

	cfg := &config.Config{
		BaseDir: t.TempDir(),
	}
	s, err := New(cfg)
	require.NoError(t, err)

	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	fspb.RegisterServiceServer(grpcServer, s)

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

	client := fspb.NewServiceClient(conn)
	return s, client
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		BaseDir: t.TempDir(),
	}
	s, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, s)
}

func TestNew_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "nested", "dir")
	cfg := &config.Config{
		BaseDir: newDir,
	}
	s, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, s)

	// Verify the directory was created
	info, err := os.Stat(newDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestServer_UploadFile(t *testing.T) {
	s, client := testServerWithClient(t)
	ctx := context.Background()
	content := "test file content"

	stream, err := client.UploadFile(ctx)
	require.NoError(t, err)

	err = filestream.Load(strings.NewReader(content), stream)
	require.NoError(t, err)

	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetUuid())

	// Verify the file was written correctly
	f, err := s.fs.Open(resp.GetUuid())
	require.NoError(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
}

func TestServer_UploadFile_LargeContent(t *testing.T) {
	s, client := testServerWithClient(t)
	ctx := context.Background()
	// Create content larger than typical chunk size
	content := strings.Repeat("abcdefghij", 100000)

	stream, err := client.UploadFile(ctx)
	require.NoError(t, err)

	err = filestream.Load(strings.NewReader(content), stream)
	require.NoError(t, err)

	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetUuid())

	// Verify the file was written correctly
	f, err := s.fs.Open(resp.GetUuid())
	require.NoError(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
}

func TestServer_DownloadFile(t *testing.T) {
	s, client := testServerWithClient(t)
	ctx := context.Background()
	content := "test file content for download"

	// First create a file directly
	of, err := s.fs.Create("test-uuid")
	require.NoError(t, err)
	_, err = of.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, of.Close())

	// Now download it via gRPC
	stream, err := client.DownloadFile(ctx, &fspb.FileInfo{Uuid: "test-uuid"})
	require.NoError(t, err)

	var buf bytes.Buffer
	err = filestream.Save(stream, &buf)
	require.NoError(t, err)
	require.Equal(t, content, buf.String())
}

func TestServer_DownloadFile_NotFound(t *testing.T) {
	_, client := testServerWithClient(t)
	ctx := context.Background()

	stream, err := client.DownloadFile(ctx, &fspb.FileInfo{Uuid: "nonexistent-uuid"})
	require.NoError(t, err)

	var buf bytes.Buffer
	err = filestream.Save(stream, &buf)
	require.Error(t, err)
}

func TestServer_UploadAndDownload(t *testing.T) {
	_, client := testServerWithClient(t)
	ctx := context.Background()
	content := "round trip content"

	// Upload
	uploadStream, err := client.UploadFile(ctx)
	require.NoError(t, err)

	err = filestream.Load(strings.NewReader(content), uploadStream)
	require.NoError(t, err)

	fileInfo, err := uploadStream.CloseAndRecv()
	require.NoError(t, err)

	// Download using the UUID from upload
	downloadStream, err := client.DownloadFile(ctx, fileInfo)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = filestream.Save(downloadStream, &buf)
	require.NoError(t, err)

	require.Equal(t, content, buf.String())
}

func TestServer_MultipleFiles(t *testing.T) {
	_, client := testServerWithClient(t)
	ctx := context.Background()

	files := map[string]string{
		"file1": "content of file 1",
		"file2": "content of file 2 with more data",
		"file3": "short",
	}

	fileInfos := make(map[string]*fspb.FileInfo)

	// Upload all files
	for name, content := range files {
		stream, err := client.UploadFile(ctx)
		require.NoError(t, err)

		err = filestream.Load(strings.NewReader(content), stream)
		require.NoError(t, err)

		info, err := stream.CloseAndRecv()
		require.NoError(t, err)
		fileInfos[name] = info
	}

	// Download and verify all files
	for name, content := range files {
		stream, err := client.DownloadFile(ctx, fileInfos[name])
		require.NoError(t, err)

		var buf bytes.Buffer
		err = filestream.Save(stream, &buf)
		require.NoError(t, err)

		require.Equal(t, content, buf.String(), "content mismatch for file %s", name)
	}
}
