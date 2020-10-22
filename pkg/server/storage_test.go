package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	neopb "neo/lib/genproto/neo"
)

func testDB() (*bolt.DB, func()) {
	tmpFile, err := ioutil.TempFile("", "db")
	if err != nil {
		panic(err)
	}
	db, err := bolt.Open(tmpFile.Name(), 0755, nil)
	if err != nil {
		panic(err)
	}
	return db, func() {
		defer os.Remove(tmpFile.Name())
		if err := db.Close(); err != nil {
			panic(err)
		}
	}
}

func TestCachedStorage_States(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	if err != nil {
		t.Fatalf("NewStorage() failed with unexpected error = %v", err)
	}
	for i := 0; i < 5; i++ {
		state := &neopb.ExploitState{
			ExploitId: fmt.Sprintf("%d", i),
			File:      &neopb.FileInfo{Uuid: "1"},
		}
		config := &neopb.ExploitConfiguration{
			Entrypoint: "./kek",
			IsArchive:  false,
		}
		if err := cs.UpdateExploitVersion(state, config); err != nil {
			t.Errorf("UpdateExploitVersion(): got unexpected error = %v", err)
		}
	}
	states := cs.States()
	if len(states) != 5 {
		t.Errorf("States(): wrong number of states returned, want: %d, got %d", 5, len(states))
	}
}

func TestCachedStorage_UpdateStates(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	if err != nil {
		t.Fatalf("NewStorage() failed with unexpected error = %v", err)
	}
	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
	}
	config := &neopb.ExploitConfiguration{
		Entrypoint: "./kek",
		IsArchive:  false,
	}
	if err := cs.UpdateExploitVersion(state, config); err != nil {
		t.Fatalf("UpdateExploitVersion(): got unexpected error = %v", err)
	}
	if state.Version != 1 {
		t.Errorf("UpdateExploitVersion(): wrong version returned: want: 1, got: %d", state.Version)
	}
	s, _ := cs.State(state.ExploitId)
	if diff := cmp.Diff(state, s, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected state diff: (-want +got):\n%s", diff)
	}
	c, _ := cs.Configuration(s)
	if diff := cmp.Diff(config, c, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected config diff: (-want +got):\n%s", diff)
	}
	state = &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "2"},
	}
	config = &neopb.ExploitConfiguration{
		Entrypoint: "./kek2",
		IsArchive:  true,
	}
	if err := cs.UpdateExploitVersion(state, config); err != nil {
		t.Fatalf("UpdateExploitVersion(): got unexpected error = %v", err)
	}
	if state.Version != 2 {
		t.Errorf("UpdateExploitVersion(): wrong version returned: want: 2, got: %d", state.Version)
	}
	s, _ = cs.State(state.ExploitId)
	if diff := cmp.Diff(state, s, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected state diff: (-want +got):\n%s", diff)
	}
	c, _ = cs.Configuration(s)
	if diff := cmp.Diff(config, c, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateExploitVersion(): unexpected config diff: (-want +got):\n%s", diff)
	}
	if len(cs.States()) != 1 {
		t.Errorf("States(): want: %d, got: %d", 1, len(cs.States()))
	}
}

func TestCachedStorage_UpdateExploitVersionDB(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	cs, err := NewStorage(db)
	if err != nil {
		t.Fatalf("NewStorage() failed with unexpected error = %v", err)
	}
	state := &neopb.ExploitState{
		ExploitId: "1",
		File:      &neopb.FileInfo{Uuid: "1"},
	}
	config := &neopb.ExploitConfiguration{
		Entrypoint: "./kek",
		IsArchive:  false,
	}
	if err := cs.UpdateExploitVersion(state, config); err != nil {
		t.Fatalf("UpdateExploitVersion(): got unexpected error = %v", err)
	}
	if err := cs.readDB(); err != nil {
		t.Fatalf("readDB(): got unexpected error: %v", err)
	}
	if len(cs.States()) != 1 {
		t.Errorf("States(): want: %d, got: %d", 1, len(cs.States()))
	}
	if len(cs.configCache) != 1 {
		t.Errorf("configs: want: %d, got: %d", 1, len(cs.States()))
	}
}

// TestCachedStorage_readDB test the implementation of the readDB function.
func TestCachedStorage_readDB(t *testing.T) {
	db, cleanup := testDB()
	defer cleanup()
	state := &neopb.ExploitState{
		ExploitId: "1",
		Version:   1,
	}
	config := &neopb.ExploitConfiguration{
		Entrypoint: "./kek",
		IsArchive:  false,
	}
	cs, err := NewStorage(db)
	if err != nil {
		t.Fatalf("NewStorage() failed with unexpected error = %v", err)
	}
	if err := cs.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stateBucketKey))
		stateBytes, err := proto.Marshal(state)
		if err != nil {
			t.Fatalf("proto.Marshall(): failed with error: %v", err)
		}
		stateKey := []byte(fmt.Sprintf("%s:%d", state.ExploitId, state.Version))
		if err := b.Put(stateKey, stateBytes); err != nil {
			return err
		}
		b = tx.Bucket([]byte(configurationBucketKey))
		confBytes, err := proto.Marshal(config)
		if err != nil {
			return err
		}
		if err := b.Put(stateKey, confBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("db.Update() failed with error = %v", err)
	}
	if err := cs.readDB(); err != nil {
		t.Fatalf("readDB() failed with unexpected error = %v", err)
	}
	if diff := cmp.Diff(state, cs.stateCache["1"], protocmp.Transform()); diff != "" {
		t.Errorf("readDB(): unexpected diff for exploit with id = 1")
	}
	if diff := cmp.Diff(config, cs.configCache["1:1"], protocmp.Transform()); diff != "" {
		t.Errorf("readDB(): unexpected diff for exploit with id = 1")
	}
}
