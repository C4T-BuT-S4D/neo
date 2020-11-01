package server

import (
	"sync"
	"time"
)

const beatThreshold = 3

func newVisitsMap() *visitsMap {
	return &visitsMap{
		visits: make(map[string]time.Time),
	}
}

type visitsMap struct {
	visits map[string]time.Time
	m      sync.Mutex
}

func (vm *visitsMap) Add(cid string) {
	vm.m.Lock()
	defer vm.m.Unlock()
	vm.visits[cid] = time.Now()
}

func (vm *visitsMap) Invalidate(now time.Time, pingEvery time.Duration) (res []string) {
	vm.m.Lock()
	defer vm.m.Unlock()
	for k, v := range vm.visits {
		if v.Add(pingEvery * beatThreshold).Before(now) {
			delete(vm.visits, k)
			res = append(res, k)
		}
	}
	return res
}
