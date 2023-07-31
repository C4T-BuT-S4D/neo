package server

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"neo/pkg/hostbucket"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	neopb "neo/lib/genproto/neo"
)

func testServer() (*ExploitManagerServer, func()) {
	db, cleanupDB := testDB()
	st, err := NewStorage(db)
	if err != nil {
		panic(err)
	}
	dir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		BaseDir:          dir,
		MetricsNamespace: fmt.Sprintf("tests_%s", uuid.NewString()[:8]),
	}
	es, err := New(cfg, st, nil)
	if err != nil {
		panic(err)
	}
	return es, func() {
		cleanupDB()
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
}

func TestExploitManagerServer_UpdateExploit(t *testing.T) {
	es, clean := testServer()
	defer clean()
	cfg := &neopb.ExploitConfiguration{
		Entrypoint: "bin",
		IsArchive:  false,
	}
	r := &neopb.UpdateExploitRequest{
		State: &neopb.ExploitState{
			ExploitId: "1",
			File:      &neopb.FileInfo{Uuid: "1"},
			Config:    cfg,
		},
	}
	resp, err := es.UpdateExploit(context.Background(), r)
	require.NoError(t, err)
	want := &neopb.ExploitState{
		ExploitId: "1",
		Version:   1,
		File:      r.State.File,
		Config:    cfg,
	}
	if diff := cmp.Diff(want, resp.State, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploit() mismatch (-want +got):\n%s", diff)
	}
}

func TestExploitManagerServer_Exploit(t *testing.T) {
	es, clean := testServer()
	defer clean()
	cfg := &neopb.ExploitConfiguration{
		Entrypoint: "bin",
		IsArchive:  false,
	}
	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
		Config:    cfg,
	}
	req := &neopb.UpdateExploitRequest{State: state}
	ctx := context.Background()
	_, err := es.UpdateExploit(ctx, req)
	require.NoError(t, err)
	resp, err := es.Exploit(ctx, &neopb.ExploitRequest{ExploitId: state.ExploitId})
	require.NoError(t, err)
	wantState := &neopb.ExploitState{
		ExploitId: "1",
		Version:   1,
		File:      state.File,
		Config:    cfg,
	}
	if diff := cmp.Diff(wantState, resp.State, protocmp.Transform()); diff != "" {
		t.Errorf("Exploit() state mismatch (-want +got):\n%s", diff)
	}
}

func TestExploitManagerServer_Ping(t *testing.T) {
	es, clean := testServer()
	defer clean()
	es.buckets = hostbucket.New(map[string]string{"id1": "ip1", "id2": "ip2"})
	es.config.FarmURL = "test"
	ctx := context.Background()
	cfg := &neopb.ExploitConfiguration{
		Entrypoint: "bin",
		IsArchive:  false,
	}
	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
		Config:    cfg,
	}
	r := &neopb.UpdateExploitRequest{State: state}
	updateResp, err := es.UpdateExploit(ctx, r)
	require.NoError(t, err)

	req := &neopb.PingRequest{ClientId: "id1", Type: neopb.PingRequest_HEARTBEAT}
	resp, err := es.Ping(ctx, req)
	require.NoError(t, err)
	want := []*neopb.ExploitState{updateResp.State}
	if diff := cmp.Diff(want, resp.State.Exploits, protocmp.Transform()); diff != "" {
		t.Errorf("Ping() states mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(es.buckets.Buckets(), resp.State.ClientTeamMap, protocmp.Transform()); diff != "" {
		t.Errorf("Ping() bucket mismatch (-want +got):\n%s", diff)
	}
	require.NotEmpty(t, es.buckets.Buckets()[req.ClientId].Teams)
	require.Equal(t, es.config.FarmURL, resp.State.Config.FarmUrl)
	require.True(t, es.visits.visits["id1"].Before(time.Now()))
}

func TestExploitManagerServer_BroadcastCommand(t *testing.T) {
	es, clean := testServer()
	defer clean()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var received *neopb.Command
	signal := make(chan struct{})
	handler := func(msg *neopb.Command) error {
		received = msg
		close(signal)
		return nil
	}
	testSub := es.broadcastPubSub.Subscribe(handler)
	defer es.broadcastPubSub.Unsubscribe(testSub)
	go testSub.Run(ctx)

	r := &neopb.Command{Command: "echo 123"}
	_, err := es.BroadcastCommand(ctx, r)
	require.NoError(t, err)

	select {
	case <-signal:
		break
	case <-time.After(time.Millisecond * 100):
		t.Errorf("Handler was not called in time")
	}
	require.Equal(t, r, received)
}
