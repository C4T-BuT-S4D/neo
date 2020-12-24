package rendezvous

import (
	"hash"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spaolacci/murmur3"
)

type Rendezvous struct {
	hcache map[string]uint64
	mu     sync.RWMutex
	h      hash.Hash64
}

func New() *Rendezvous {
	return &Rendezvous{
		hcache: make(map[string]uint64),
		h:      murmur3.New64WithSeed(1337),
	}
}

func (r *Rendezvous) checkCache(key string) (uint64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	raw, ok := r.hcache[key]
	return raw, ok
}

func (r *Rendezvous) calcHash(key string) uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.h.Reset()
	if _, err := r.h.Write([]byte(key)); err != nil {
		logrus.Fatalf("Error calculating hash: %v", err)
	}
	rawVal := r.h.Sum64()
	r.hcache[key] = rawVal
	return rawVal
}

func (r *Rendezvous) Calculate(node string, weight int, value string) float64 {
	key := combineKey(node, value)
	if raw, ok := r.checkCache(key); ok {
		return weightHash(raw, weight)
	}
	raw := r.calcHash(key)
	return weightHash(raw, weight)
}
