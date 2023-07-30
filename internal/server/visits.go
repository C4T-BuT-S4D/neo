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

// MarkInvalid resets the client timestamp, so the next Invalidate pass will remove the client
func (vm *visitsMap) MarkInvalid(cid string) {
	vm.m.Lock()
	defer vm.m.Unlock()
	if _, ok := vm.visits[cid]; !ok {
		return
	}
	vm.visits[cid] = time.Unix(0, 0)
}

func (vm *visitsMap) Invalidate(now time.Time, pingEvery time.Duration) (alive, dead []string) {
	vm.m.Lock()
	defer vm.m.Unlock()

	for k, v := range vm.visits {
		if v.Add(pingEvery * beatThreshold).Before(now) {
			delete(vm.visits, k)
			dead = append(dead, k)
		} else {
			alive = append(alive, k)
		}
	}
	return alive, dead
}

func (vm *visitsMap) Size() int {
	vm.m.Lock()
	defer vm.m.Unlock()
	return len(vm.visits)
}
