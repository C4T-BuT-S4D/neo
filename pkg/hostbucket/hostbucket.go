package hostbucket

import (
	"sync"

	neopb "neo/lib/genproto/neo"
)

func New(ips []string) *HostBucket {
	return &HostBucket{
		buck: make(map[string]*neopb.TeamBucket),
		ids:  nil,
		ips:  ips,
	}
}

type HostBucket struct {
	m    sync.RWMutex
	buck map[string]*neopb.TeamBucket
	ids  []string
	ips  []string
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
	for i, ip := range hb.ips {
		teamId := hb.ids[i%len(hb.ids)]
		hb.buck[teamId].TeamIps = append(hb.buck[teamId].TeamIps, ip)
	}
}
