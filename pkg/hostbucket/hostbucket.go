package hostbucket

import (
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/c4t-but-s4d/neo/v2/pkg/rendezvous"
	epb "github.com/c4t-but-s4d/neo/v2/proto/go/exploits"
)

func New(teams map[string]string) *HostBucket {
	return &HostBucket{
		buck:  make(map[string]*epb.TeamBucket),
		nodes: nil,
		teams: teams,
		r:     rendezvous.New(),
	}
}

type HostBucket struct {
	m     sync.RWMutex
	buck  map[string]*epb.TeamBucket
	nodes []*node
	teams map[string]string
	r     *rendezvous.Rendezvous
}

// TODO: effective ip addition & deletion
func (hb *HostBucket) UpdateTeams(teams map[string]string) {
	lessFunc := func(s1, s2 string) bool {
		return s1 < s2
	}
	if !cmp.Equal(teams, hb.teams, cmpopts.SortSlices(lessFunc)) {
		hb.m.Lock()
		defer hb.m.Unlock()
		hb.teams = teams
		hb.rehash()
	}
}

func (hb *HostBucket) Buckets() map[string]*epb.TeamBucket {
	hb.m.RLock()
	defer hb.m.RUnlock()

	clone := make(map[string]*epb.TeamBucket, len(hb.buck))
	for k, v := range hb.buck {
		clone[k] = v
	}
	return clone
}

func (hb *HostBucket) Exists(id string) (exists bool) {
	hb.m.RLock()
	defer hb.m.RUnlock()
	_, exists = hb.buck[id]
	return
}

func (hb *HostBucket) AddNode(id string, weight int) {
	hb.m.Lock()
	defer hb.m.Unlock()

	if _, ok := hb.buck[id]; ok {
		return
	}

	hb.buck[id] = &epb.TeamBucket{}
	n := &node{
		id:     id,
		weight: weight,
	}
	hb.nodes = append(hb.nodes, n)
	// TODO: more effective node addition
	hb.rehash()
}

func (hb *HostBucket) DeleteNode(id string) bool {
	hb.m.Lock()
	defer hb.m.Unlock()
	if _, ok := hb.buck[id]; !ok {
		return false
	}
	for i, n := range hb.nodes {
		if n.id == id {
			last := len(hb.nodes) - 1
			hb.nodes[i] = hb.nodes[last]
			hb.nodes[last] = nil
			hb.nodes = hb.nodes[:last]
			delete(hb.buck, id)

			// TODO: more effective node deletion
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
	if len(hb.nodes) == 0 {
		return
	}
	for id, ip := range hb.teams {
		bestHash := 0.0
		bestNode := ""

		for _, n := range hb.nodes {
			hash := hb.r.Calculate(n.id, n.weight, id)
			if bestNode == "" || hash > bestHash {
				bestNode = n.id
				bestHash = hash
			}
		}

		if hb.buck[bestNode].Teams == nil {
			hb.buck[bestNode].Teams = make(map[string]string)
		}
		hb.buck[bestNode].Teams[id] = ip
	}
}
