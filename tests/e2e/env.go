package e2e

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/c4t-but-s4d/neo/v2/internal/client"
	"github.com/c4t-but-s4d/neo/v2/internal/logstor"
	serverConfig "github.com/c4t-but-s4d/neo/v2/internal/server/config"
	"github.com/c4t-but-s4d/neo/v2/internal/server/exploits"
	"github.com/c4t-but-s4d/neo/v2/internal/server/fs"
	"github.com/c4t-but-s4d/neo/v2/internal/server/logs"
	serverMetrics "github.com/c4t-but-s4d/neo/v2/internal/server/metrics"
	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
	fspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/fileserver"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

const bufSize = 1024 * 1024

type Env struct {
	T *testing.T

	Config         *serverConfig.Config
	ExploitsServer *exploits.Server
	FSServer       *fs.Server
	LogsServer     *logs.Server
	LogsStorage    *logstor.MockStorage

	GRPCServer    *grpc.Server
	MetricsServer *httptest.Server

	bufLis     *bufconn.Listener
	ClientConn *grpc.ClientConn
	Client     *client.Client
}

type EnvOption func(*envOptions)

type envOptions struct {
	teams map[string]string
}

func WithTeams(teams map[string]string) EnvOption {
	return func(o *envOptions) {
		o.teams = teams
	}
}

func NewEnv(t *testing.T, opts ...EnvOption) *Env {
	t.Helper()

	options := &envOptions{
		teams: map[string]string{
			"team1": "10.0.0.1",
			"team2": "10.0.0.2",
		},
	}
	for _, opt := range opts {
		opt(options)
	}

	tmpDir := t.TempDir()

	cfg := &serverConfig.Config{
		BaseDir:          filepath.Join(tmpDir, "exploits"),
		MetricsNamespace: "test_" + uuid.NewString()[:8],
		PingEvery:        time.Second,
		SubmitEvery:      time.Second,
		FarmConfig: serverConfig.FarmConfig{
			Teams: options.teams,
		},
	}

	// Create bolt DB for exploits storage
	db := testDB(t, tmpDir)
	storage, err := exploits.NewStorage(db)
	require.NoError(t, err)

	exploitsServer := exploits.New(cfg, storage)

	fsServer, err := fs.New(cfg)
	require.NoError(t, err)

	logsStorage := logstor.NewMockStorage()
	logsServer := logs.New(logsStorage)

	// Setup gRPC server with bufconn
	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	epb.RegisterServiceServer(grpcServer, exploitsServer)
	fspb.RegisterServiceServer(grpcServer, fsServer)
	logspb.RegisterServiceServer(grpcServer, logsServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server exited: %v", err)
		}
	}()

	// Setup HTTP server for metrics proxy (mock backend)
	metricsBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/api/metrics/", serverMetrics.NewProxyHandler(
		http.DefaultClient, metricsBackend.URL, "",
	))
	metricsServer := httptest.NewServer(metricsMux)

	// Create client connection via bufconn
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	neoClient := client.New(conn, "test-client")

	t.Cleanup(func() {
		conn.Close()
		grpcServer.Stop()
		lis.Close()
		metricsServer.Close()
		metricsBackend.Close()
	})

	return &Env{
		T:              t,
		Config:         cfg,
		ExploitsServer: exploitsServer,
		FSServer:       fsServer,
		LogsServer:     logsServer,
		LogsStorage:    logsStorage,
		GRPCServer:     grpcServer,
		MetricsServer:  metricsServer,
		bufLis:         lis,
		ClientConn:     conn,
		Client:         neoClient,
	}
}

func (e *Env) NewClient(id string) *client.Client {
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return e.bufLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(e.T, err)
	e.T.Cleanup(func() {
		conn.Close()
	})
	return client.New(conn, id)
}

func testDB(t *testing.T, tmpDir string) *bolt.DB {
	t.Helper()
	db, err := bolt.Open(filepath.Join(tmpDir, "db.bolt"), 0o755, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, db.Close())
	})
	return db
}
