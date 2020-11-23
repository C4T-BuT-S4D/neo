package hostbucket

import (
	"github.com/denisbrodbeck/machineid"
	"math"
	"testing"

	"neo/pkg/testutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestHostBucket_Add(t *testing.T) {
	for _, tc := range []struct {
		b     *HostBucket
		users []string
		want  map[string]int
	}{
		{
			b:     New([]string{"ip1"}),
			users: []string{"u1"},
			want:  map[string]int{"u1": 1},
		},
		{
			b:     New([]string{"ip1", "ip2", "ip3"}),
			users: []string{},
			want:  map[string]int{},
		},
	} {
		for _, u := range tc.users {
			tc.b.Add(u)
		}
		b := tc.b.Buckets()
		var gotIps []string
		for cid, wantn := range tc.want {
			ipb := b[cid]
			if len(ipb.GetTeamIps()) != wantn {
				t.Errorf("HostBucket.Add(): incorrent number of gotIps in user bucket want=%d, got=%d",
					wantn, len(ipb.GetTeamIps()),
				)
			}
			gotIps = append(gotIps, ipb.GetTeamIps()...)
		}
		less := func(s1, s2 string) bool {
			return s1 < s2
		}
		if diff := cmp.Diff(tc.b.ips, gotIps, cmpopts.SortSlices(less)); diff != "" && len(tc.want) > 0 {
			t.Errorf("HostBucket.Add() summary ips mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestHostBucket_Add_Distribution(t *testing.T) {
	populate := func(ipCount, idCount int) *HostBucket {
		ips := make([]string, ipCount)
		for i := range ips {
			ips[i] = testutils.RandomIP()
		}
		hb := New(ips)

		mid, err := machineid.ID()
		if err != nil {
			t.Fatalf("Could not get machine id: %v", err)
		}
		for i := 0; i < idCount; i += 1 {
			id := testutils.RandomString(len(mid))
			hb.Add(id)
		}
		return hb
	}

	for _, tc := range []struct {
		ipCount      int
		idCount      int
		maxDeviation float64
	}{
		{
			100,
			5,
			0.3,
		},
		{
			1000,
			10,
			0.3,
		},
	} {
		b := populate(tc.ipCount, tc.idCount)
		sizes := make([]float64, tc.idCount)
		meanSize := 0.0
		for i := range sizes {
			id := b.ids[i]
			ips := b.buck[id].GetTeamIps()
			sizes[i] = float64(len(ips))
			meanSize += sizes[i]
		}
		meanSize /= float64(len(sizes))

		for i := range sizes {
			id := b.ids[i]
			deviation := math.Abs((sizes[i] - meanSize) / meanSize)
			if deviation > tc.maxDeviation {
				t.Errorf(
					"Deviation for bucket %s too large: %f > %f, target size: %f, got size: %f",
					id,
					deviation,
					tc.maxDeviation,
					meanSize,
					sizes[i],
				)
			}
		}
	}
}

func TestHostBucket_Balancing(t *testing.T) {
	populate := func(ipCount, idCount int) *HostBucket {
		ips := make([]string, ipCount)
		for i := range ips {
			ips[i] = testutils.RandomIP()
		}
		hb := New(ips)

		mid, err := machineid.ID()
		if err != nil {
			t.Fatalf("Could not get machine id: %v", err)
		}
		for i := 0; i < idCount; i += 1 {
			id := testutils.RandomString(len(mid))
			hb.Add(id)
		}
		return hb
	}

	for _, tc := range []struct {
		ipCount   int
		idCount   int
		maxMoved  float64
		cntAdd    int
		cntDelete int
	}{
		{
			100,
			5,
			0.3,
			0,
			1,
		},
		{
			1000,
			10,
			0.2,
			1,
			0,
		},
	} {
		b := populate(tc.ipCount, tc.idCount)

		beforeByIP := make(map[string]string)
		for _, id := range b.ids {
			ips := b.buck[id].GetTeamIps()
			for _, ip := range ips {
				beforeByIP[ip] = id
			}
		}

		getCntMoved := func() int {
			cntMoved := 0
			for _, id := range b.ids {
				ips := b.buck[id].GetTeamIps()
				for _, ip := range ips {
					if id != beforeByIP[ip] {
						cntMoved += 1
					}
					beforeByIP[ip] = id
				}
			}
			return cntMoved
		}

		for i := 0; i < tc.cntDelete; i += 1 {
			toDelete := testutils.RandomInt(0, tc.idCount)
			b.Delete(b.ids[toDelete])
		}
		cntMoved := getCntMoved()
		movedFraction := float64(cntMoved) / float64(tc.ipCount)
		if movedFraction > tc.maxMoved {
			t.Errorf("Too many ips moved after delete: %f%%, %d of %d", movedFraction*100, cntMoved, tc.ipCount)
		}

		mid, err := machineid.ID()
		if err != nil {
			t.Fatalf("Could not get machine id: %v", err)
		}
		for i := 0; i < tc.cntAdd; i += 1 {
			id := testutils.RandomString(len(mid))
			b.Add(id)
		}
		cntMoved = getCntMoved()
		movedFraction = float64(cntMoved) / float64(tc.ipCount)
		if movedFraction > tc.maxMoved {
			t.Errorf("Too many ips moved after add: %f%%, %d of %d", movedFraction*100, cntMoved, tc.ipCount)
		}
	}
}

func TestHostBucket_Delete(t *testing.T) {
	populate := func(ips []string, ids []string) *HostBucket {
		hb := New(ips)
		for _, id := range ids {
			hb.Add(id)
		}
		return hb
	}
	for _, tc := range []struct {
		b      *HostBucket
		delete []string
		want   map[string]int
	}{
		{
			b:      populate([]string{"ip1", "ip2"}, []string{"id1", "id2"}),
			delete: []string{"id1"},
			want:   map[string]int{"id2": 2},
		},
		{
			b:      populate([]string{"ip1", "ip2"}, []string{"id1"}),
			delete: []string{"id1"},
			want:   map[string]int{},
		},
	} {
		for _, tid := range tc.delete {
			tc.b.Delete(tid)
		}
		var gotIps []string
		for cid, wantn := range tc.want {
			ipb := tc.b.buck[cid]
			if len(ipb.GetTeamIps()) != wantn {
				t.Errorf("HostBucket.Delete(): incorrent number of gotIps in user bucket want=%d, got=%d",
					wantn, len(ipb.GetTeamIps()),
				)
			}
			gotIps = append(gotIps, ipb.GetTeamIps()...)
		}
		if diff := cmp.Diff(tc.b.ips, gotIps, cmpopts.SortSlices(testutils.LessString)); diff != "" && len(tc.want) > 0 {
			t.Errorf("HostBucket.Delete() summary ips mismatch (-want +got):\n%s", diff)
		}
	}
}
