package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	bolt "go.etcd.io/bbolt"

	neopb "neo/lib/genproto/neo"
)

const (
	stateBucketKey         = "states"
	configurationBucketKey = "configuration"
)

func NewBoltStorage(path string) (*CachedStorage, error) {
	db, err := bolt.Open(path, 0755, nil)
	if err != nil {
		return nil, fmt.Errorf("opening bolt db %s: %w", path, err)
	}
	return NewStorage(db)
}

func NewStorage(db *bolt.DB) (*CachedStorage, error) {
	cs := &CachedStorage{
		stateCache:  nil,
		configCache: nil,
		bdb:         db,
	}
	if err := cs.initDB(); err != nil {
		return nil, err
	}
	cs.initCache()
	return cs, nil
}

type CachedStorage struct {
	stateCache  map[string]*neopb.ExploitState
	configCache map[string]*neopb.ExploitConfiguration
	m           sync.RWMutex
	bdb         *bolt.DB
}

func (cs *CachedStorage) States() []*neopb.ExploitState {
	cs.m.RLock()
	defer cs.m.RUnlock()
	res := make([]*neopb.ExploitState, 0, len(cs.stateCache))
	for _, v := range cs.stateCache {
		res = append(res, v)
	}
	return res
}

func (cs *CachedStorage) State(exploitID string) (*neopb.ExploitState, bool) {
	cs.m.RLock()
	defer cs.m.RUnlock()
	val, ok := cs.stateCache[exploitID]
	return val, ok
}

func (cs *CachedStorage) Configuration(s *neopb.ExploitState) (*neopb.ExploitConfiguration, bool) {
	cs.m.RLock()
	defer cs.m.RUnlock()
	val, ok := cs.configCache[cs.configCacheKey(s)]
	return val, ok
}

func (cs *CachedStorage) UpdateExploitVersion(newState *neopb.ExploitState, cfg *neopb.ExploitConfiguration) error {
	cs.m.Lock()
	defer cs.m.Unlock()
	if state, ok := cs.stateCache[newState.ExploitId]; ok {
		newState.Version = state.Version + 1
	} else {
		newState.Version = 1
	}

	if err := cs.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stateBucketKey))
		key := []byte(fmt.Sprintf("%s:%d", newState.ExploitId, newState.Version))
		stateBytes, err := proto.Marshal(newState)
		if err != nil {
			return fmt.Errorf("marshalling state: %w", err)
		}
		if err := b.Put(key, stateBytes); err != nil {
			return fmt.Errorf("setting state in db: %w", err)
		}
		b = tx.Bucket([]byte(configurationBucketKey))
		confBytes, err := proto.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshalling config: %w", err)
		}
		if err := b.Put(key, confBytes); err != nil {
			return fmt.Errorf("setting config in db: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("updating db state: %w", err)
	}

	cs.stateCache[newState.ExploitId] = newState
	cs.configCache[cs.configCacheKey(newState)] = cfg
	return nil
}

func (cs *CachedStorage) configCacheKey(s *neopb.ExploitState) string {
	return fmt.Sprintf("%s:%d", s.ExploitId, s.Version)
}

func (cs *CachedStorage) initCache() {
	cs.m.Lock()
	defer cs.m.Unlock()
	if cs.stateCache == nil || cs.configCache == nil {
		cs.stateCache = make(map[string]*neopb.ExploitState)
		cs.configCache = make(map[string]*neopb.ExploitConfiguration)
		if err := cs.readDB(); err != nil {
			logrus.Errorf("Failed to read exploit data from DB: %v", err)
		}
	}
}

func (cs *CachedStorage) readDB() error {
	if err := cs.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stateBucketKey))
		if err := b.ForEach(func(k, v []byte) error {
			key := string(k)
			eID := strings.Split(key, ":")[0]
			es := new(neopb.ExploitState)
			if err := proto.Unmarshal(v, es); err != nil {
				return fmt.Errorf("unmarshalling exploit state: %w", err)
			}
			if v, ok := cs.stateCache[eID]; !ok || es.Version > v.Version {
				cs.stateCache[eID] = es
			}
			return nil
		}); err != nil {
			return fmt.Errorf("reading exploit states: %w", err)
		}
		b = tx.Bucket([]byte(configurationBucketKey))
		if err := b.ForEach(func(k, v []byte) error {
			cfg := new(neopb.ExploitConfiguration)
			if err := proto.Unmarshal(v, cfg); err != nil {
				return fmt.Errorf("unmarshalling exploit config: %w", err)
			}
			cs.configCache[string(k)] = cfg
			return nil
		}); err != nil {
			return fmt.Errorf("reading exploit configs: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reading state from db: %w", err)
	}
	return nil
}

func (cs *CachedStorage) initDB() error {
	if err := cs.bdb.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(stateBucketKey)); err != nil {
			return fmt.Errorf("creating state bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(configurationBucketKey)); err != nil {
			return fmt.Errorf("creating config bucket: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("initializing db: %w", err)
	}
	return nil
}
