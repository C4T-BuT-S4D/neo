package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	bolt "go.etcd.io/bbolt"

	neopb "neo/lib/genproto/neo"
)

func testDB() (*bolt.DB, func()) {
	tmpFile, err := os.CreateTemp("", "db")
	if err != nil {
		panic(err)
	}
	db, err := bolt.Open(tmpFile.Name(), 0755, nil)
	if err != nil {
		panic(err)
	}
	return db, func() {
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				logrus.Errorf("Error removing file: %v", err)
			}
		}(tmpFile.Name())
		if err := db.Close(); err != nil {
			panic(err)
		}
	}
}

func TestCachedStorage_States(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		state := &neopb.ExploitState{
			ExploitId: fmt.Sprintf("%d", i),
			File:      &neopb.FileInfo{Uuid: "1"},
			Config: &neopb.ExploitConfiguration{
				Entrypoint: "./kek",
				IsArchive:  false,
			},
		}
		_, err := cs.UpdateExploitVersion(state)
		require.NoError(t, err)
	}
	require.Len(t, cs.States(), 5)
}

func TestCachedStorage_UpdateStates(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	require.NoError(t, err)
	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
		Config: &neopb.ExploitConfiguration{
			Entrypoint: "./kek",
			IsArchive:  false,
		},
	}
	_, err = cs.UpdateExploitVersion(state)
	require.NoError(t, err)
	require.EqualValues(t, 1, state.Version)
	s, _ := cs.GetState(state.ExploitId)
	if diff := cmp.Diff(state, s, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected state diff: (-want +got):\n%s", diff)
	}
	state = &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "2"},
		Config: &neopb.ExploitConfiguration{
			Entrypoint: "./kek2",
			IsArchive:  true,
		},
	}
	_, err = cs.UpdateExploitVersion(state)
	require.NoError(t, err)
	require.Len(t, cs.States(), 1)
	require.EqualValues(t, 2, state.Version)

	s, _ = cs.GetState(state.ExploitId)
	if diff := cmp.Diff(state, s, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected state diff: (-want +got):\n%s", diff)
	}
}

func TestCachedStorage_UpdateExploitVersionDB(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	require.NoError(t, err)

	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
		Config: &neopb.ExploitConfiguration{
			Entrypoint: "./kek",
			IsArchive:  false,
		},
	}
	_, err = cs.UpdateExploitVersion(state)
	require.NoError(t, err)
	require.NoError(t, cs.readDB())
	require.Len(t, cs.States(), 1)
}

// TestCachedStorage_readDB test the implementation of the readDB function.
func TestCachedStorage_readDB(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	state := &neopb.ExploitState{
		ExploitId: "1",
		Version:   1,
		Config: &neopb.ExploitConfiguration{
			Entrypoint: "./kek",
			IsArchive:  false,
		},
	}
	cs, err := NewStorage(db)
	require.NoError(t, err)
	require.NoError(t, cs.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stateBucketKey))
		stateBytes, err := proto.Marshal(state)
		require.NoError(t, err)
		stateKey := []byte(fmt.Sprintf("%s:%d", state.ExploitId, state.Version))
		if err := b.Put(stateKey, stateBytes); err != nil {
			return fmt.Errorf("setting state in db: %w", err)
		}
		return nil
	}))
	require.NoError(t, cs.readDB())
	if diff := cmp.Diff(state, cs.stateCache["1"], protocmp.Transform()); diff != "" {
		t.Errorf("readDB(): unexpected diff for exploit with id = 1")
	}
}
