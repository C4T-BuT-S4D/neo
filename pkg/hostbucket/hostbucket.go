package hostbucket

import (
	"neo/pkg/rendezvous"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	neopb "neo/lib/genproto/neo"
)

func New(ips []string) *HostBucket {
	return &HostBucket{
		buck: make(map[string]*neopb.TeamBucket),
		ids:  nil,
		ips:  ips,
		r:    rendezvous.New(),
	}
}

type HostBucket struct {
	m    sync.RWMutex
	buck map[string]*neopb.TeamBucket
	ids  []string
	ips  []string
	r    *rendezvous.Rendezvous
}

func (hb *HostBucket) UpdateIPS(ips []string) {
	lessFunc := func(s1, s2 string) bool {
		return s1 < s2
	}
	if !cmp.Equal(ips, hb.ips, cmpopts.SortSlices(lessFunc)) {
		hb.m.Lock()
		defer hb.m.Unlock()
		hb.ips = ips
		hb.rehash()
	}
}

func (hb *HostBucket) Buckets() map[string]*neopb.TeamBucket {
	hb.m.RLock()
	defer hb.m.RUnlock()
	return hb.buck
}

func (hb *HostBucket) Exists(tid string) (exists bool) {
	hb.m.RLock()
	defer hb.m.RUnlock()
	_, exists = hb.buck[tid]
	return
}

func (hb *HostBucket) Add(tid string) {
	hb.m.Lock()
	defer hb.m.Unlock()
	hb.buck[tid] = &neopb.TeamBucket{}
	hb.ids = append(hb.ids, tid)
	hb.rehash()
}

func (hb *HostBucket) Delete(tid string) bool {
	hb.m.Lock()
	defer hb.m.Unlock()
	if _, ok := hb.buck[tid]; !ok {
		return false
	}
	for i, v := range hb.ids {
		if v == tid {
			hb.ids[i] = hb.ids[len(hb.ids)-1]
			hb.ids[len(hb.ids)-1] = ""
			hb.ids = hb.ids[:len(hb.ids)-1]
			delete(hb.buck, tid)
			hb.rehash()
			return true
		}
	}
	return false
}

func (hb *HostBucket) rehash() {
	for _, v := range hb.buck {
		v.Reset()
	}
	if len(hb.ids) == 0 {
		return
	}
	for _, ip := range hb.ips {
		bestHash := uint64(0)
		bestNode := ""

		for _, id := range hb.ids {
			hash := hb.r.Calculate(id, ip)
			if bestNode == "" || hash > bestHash {
				bestNode = id
				bestHash = hash
			}
		}

		hb.buck[bestNode].TeamIps = append(hb.buck[bestNode].TeamIps, ip)
	}
}
