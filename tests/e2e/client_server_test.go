package e2e

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

func TestHeartbeat(t *testing.T) {
	env := NewEnv(t)
	ctx := context.Background()

	state, err := env.Client.Heartbeat(ctx)
	require.NoError(t, err)
	require.NotNil(t, state)

	// Client should be assigned teams
	clientTeams, ok := state.GetClientTeamMap()["test-client"]
	require.True(t, ok)
	require.NotEmpty(t, clientTeams.GetTeams())
}

func TestHeartbeat_MultipleClients(t *testing.T) {
	env := NewEnv(t, WithTeams(map[string]string{
		"team1": "10.0.0.1",
		"team2": "10.0.0.2",
		"team3": "10.0.0.3",
		"team4": "10.0.0.4",
	}))
	ctx := context.Background()

	// First client heartbeat
	_, err := env.Client.Heartbeat(ctx)
	require.NoError(t, err)

	// Create second client and heartbeat
	client2 := env.NewClient("test-client-2")
	_, err = client2.Heartbeat(ctx)
	require.NoError(t, err)

	// Get fresh state after both clients have registered
	state, err := env.Client.Heartbeat(ctx)
	require.NoError(t, err)

	clientMap := state.GetClientTeamMap()
	require.Contains(t, clientMap, "test-client")
	require.Contains(t, clientMap, "test-client-2")

	// Combined teams should cover all teams
	var totalTeams int
	for _, bucket := range clientMap {
		totalTeams += len(bucket.GetTeams())
	}
	require.Equal(t, 4, totalTeams)
}

func TestFileUploadDownload(t *testing.T) {
	env := NewEnv(t)
	ctx := context.Background()

	content := "#!/bin/bash\necho 'Hello from exploit'"

	// Upload file
	fileInfo, err := env.Client.UploadFile(ctx, strings.NewReader(content))
	require.NoError(t, err)
	require.NotEmpty(t, fileInfo.GetUuid())

	// Download file
	var buf bytes.Buffer
	err = env.Client.DownloadFile(ctx, fileInfo, &buf)
	require.NoError(t, err)
	require.Equal(t, content, buf.String())
}

func TestExploitLifecycle(t *testing.T) {
	env := NewEnv(t)
	ctx := context.Background()

	// 1. Upload exploit file
	exploitContent := "#!/bin/bash\necho $1"
	fileInfo, err := env.Client.UploadFile(ctx, strings.NewReader(exploitContent))
	require.NoError(t, err)

	// 2. Register exploit
	exploitState := &epb.ExploitState{
		ExploitId: "test-exploit",
		File:      fileInfo,
		Config: &epb.ExploitConfiguration{
			Entrypoint: "exploit.sh",
			Timeout:    durationpb.New(time.Minute),
			RunEvery:   durationpb.New(time.Minute),
		},
	}
	updatedState, err := env.Client.UpdateExploit(ctx, exploitState)
	require.NoError(t, err)
	require.Equal(t, int64(1), updatedState.GetVersion())

	// 3. Verify exploit is returned in heartbeat
	serverState, err := env.Client.Heartbeat(ctx)
	require.NoError(t, err)
	require.Len(t, serverState.GetExploits(), 1)
	require.Equal(t, "test-exploit", serverState.GetExploits()[0].GetExploitId())

	// 4. Get exploit by ID
	exploitResp, err := env.Client.Exploit(ctx, "test-exploit")
	require.NoError(t, err)
	require.Equal(t, int64(1), exploitResp.GetState().GetVersion())

	// 5. Update exploit (new version)
	exploitState.Config.Timeout = durationpb.New(2 * time.Minute)
	updatedState, err = env.Client.UpdateExploit(ctx, exploitState)
	require.NoError(t, err)
	require.Equal(t, int64(2), updatedState.GetVersion())

	// 6. Disable exploit
	err = env.Client.SetExploitDisabled(ctx, "test-exploit", true)
	require.NoError(t, err)

	exploitResp, err = env.Client.Exploit(ctx, "test-exploit")
	require.NoError(t, err)
	require.True(t, exploitResp.GetState().GetConfig().GetDisabled())
}

func TestLogsAddAndSearch(t *testing.T) {
	env := NewEnv(t)
	ctx := context.Background()

	// Add log lines
	lines := []*logspb.LogLine{
		{
			Exploit:   "test-exploit",
			Version:   1,
			Message:   "Starting exploit",
			Level:     "info",
			Team:      "team1",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "test-exploit",
			Version:   1,
			Message:   "Got flag: FLAG{test}",
			Level:     "info",
			Team:      "team1",
			Timestamp: timestamppb.New(time.Now()),
		},
		{
			Exploit:   "other-exploit",
			Version:   1,
			Message:   "Different exploit",
			Level:     "info",
			Team:      "team2",
			Timestamp: timestamppb.New(time.Now()),
		},
	}
	err := env.Client.AddLogLines(ctx, lines...)
	require.NoError(t, err)

	// Search logs
	resultsCh, err := env.Client.SearchLogLines(ctx, "test-exploit", 1)
	require.NoError(t, err)

	var results []*logspb.LogLine
	for batch := range resultsCh {
		results = append(results, batch...)
	}

	require.Len(t, results, 2)
	require.Equal(t, "Starting exploit", results[0].GetMessage())
}

func TestBroadcastCommand(t *testing.T) {
	env := NewEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start listening for broadcasts
	broadcastCh, err := env.Client.ListenBroadcasts(ctx)
	require.NoError(t, err)

	// Small delay to ensure subscription is established
	time.Sleep(50 * time.Millisecond)

	// Send broadcast command
	err = env.Client.BroadcastCommand(ctx, "echo 'test broadcast'")
	require.NoError(t, err)

	// Receive broadcast
	select {
	case cmd := <-broadcastCh:
		require.Equal(t, "echo 'test broadcast'", cmd.GetCommand())
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive broadcast in time")
	}
}

func TestSingleRun(t *testing.T) {
	env := NewEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Register an exploit first
	fileInfo, err := env.Client.UploadFile(ctx, strings.NewReader("#!/bin/bash\necho test"))
	require.NoError(t, err)

	_, err = env.Client.UpdateExploit(ctx, &epb.ExploitState{
		ExploitId: "single-run-exploit",
		File:      fileInfo,
		Config: &epb.ExploitConfiguration{
			Entrypoint: "exploit.sh",
			Timeout:    durationpb.New(time.Minute),
			RunEvery:   durationpb.New(time.Hour), // Don't auto-run
		},
	})
	require.NoError(t, err)

	// Start listening for single runs
	singleRunCh, err := env.Client.ListenSingleRuns(ctx)
	require.NoError(t, err)

	// Small delay to ensure subscription is established
	time.Sleep(50 * time.Millisecond)

	// Trigger single run
	err = env.Client.SingleRun(ctx, "single-run-exploit")
	require.NoError(t, err)

	// Receive single run request
	select {
	case req := <-singleRunCh:
		require.Equal(t, "single-run-exploit", req.GetExploitId())
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive single run request in time")
	}
}

func TestServerState_ConfigPropagation(t *testing.T) {
	env := NewEnv(t)
	ctx := context.Background()

	// Get initial state
	state, err := env.Client.GetServerState(ctx)
	require.NoError(t, err)
	require.NotNil(t, state.GetConfig())

	// Verify config fields are propagated
	config := state.GetConfig()
	require.NotNil(t, config.GetPingEvery())
	require.NotNil(t, config.GetSubmitEvery())
}
