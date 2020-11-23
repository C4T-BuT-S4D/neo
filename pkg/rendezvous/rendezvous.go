package rendezvous

import (
	"github.com/sirupsen/logrus"
	"hash/maphash"
	"sync"
)

type Rendezvous struct {
	hcache map[string]uint64
	mu     sync.RWMutex
	seed   maphash.Seed
}

func New() *Rendezvous {
	return &Rendezvous{
		hcache: make(map[string]uint64),
		seed:   maphash.MakeSeed(),
	}
}

func (r *Rendezvous) getKey(nodeId, value string) string {
	return nodeId + ":" + value
}

func (r *Rendezvous) checkCache(key string) (uint64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	hash, ok := r.hcache[key]
	return hash, ok
}

func (r *Rendezvous) setCache(key string, value uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hcache[key] = value
}

func (r *Rendezvous) calcHash(key string) uint64 {
	h := new(maphash.Hash)
	h.SetSeed(r.seed)
	h.SetSeed(maphash.MakeSeed())
	if _, err := h.WriteString(key); err != nil {
		logrus.Fatalf("Error calculating hash: %v", err)
	}
	hash := h.Sum64()
	r.setCache(key, hash)
	return hash
}

func (r *Rendezvous) Calculate(nodeId, value string) uint64 {
	key := r.getKey(nodeId, value)
	if hash, ok := r.checkCache(key); ok {
		return hash
	}
	return r.calcHash(key)
}
