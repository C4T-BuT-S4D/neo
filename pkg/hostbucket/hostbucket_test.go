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
			tc.b.AddNode(u, 1)
		}
		b := tc.b.Buckets()
		var gotIps []string
		for cid, wantn := range tc.want {
			ipb := b[cid]
			if len(ipb.GetTeamIps()) != wantn {
				t.Errorf("HostBucket.AddNode(): incorrent number of gotIps in user bucket want=%d, got=%d",
					wantn, len(ipb.GetTeamIps()),
				)
			}
			gotIps = append(gotIps, ipb.GetTeamIps()...)
		}
		less := func(s1, s2 string) bool {
			return s1 < s2
		}
		if diff := cmp.Diff(tc.b.ips, gotIps, cmpopts.SortSlices(less)); diff != "" && len(tc.want) > 0 {
			t.Errorf("HostBucket.AddNode() summary ips mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestHostBucket_Add_Distribution(t *testing.T) {
	populate := func(ipCount, idCount int, weightMax int) *HostBucket {
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
			w := testutils.RandomInt(1, weightMax+1)
			hb.AddNode(id, w)
		}
		return hb
	}

	for _, tc := range []struct {
		ipCount       int
		idCount       int
		maxDeviation  float64
		maxStdDev     float64
		maxMeanStdDev float64
		weightMax     int
		runs          int
	}{
		{
			100,
			5,
			1.5,
			0.7,
			0.3,
			32,
			30,
		},
		{
			1000,
			10,
			1,
			0.5,
			0.2,
			32,
			30,
		},
	} {
		meanStdDev := 0.0
		for i := 0; i < tc.runs; i += 1 {
			b := populate(tc.ipCount, tc.idCount, tc.weightMax)
			sizes := make([]float64, tc.idCount)
			meanSize := 0.0
			for i := range sizes {
				id := b.nodes[i].ID
				ips := b.buck[id].GetTeamIps()
				sizes[i] = float64(len(ips)) / float64(b.nodes[i].Weight)
				meanSize += sizes[i]
			}
			meanSize /= float64(len(sizes))

			stdDev := 0.0
			for i := range sizes {
				id := b.nodes[i].ID
				deviation := math.Abs((sizes[i] - meanSize) / meanSize)
				if deviation > tc.maxDeviation {
					t.Errorf(
						"Deviation for bucket %s too large: %f > %f, target size: %f, weight %d, got size: %f",
						id,
						deviation,
						tc.maxDeviation,
						meanSize,
						b.nodes[i].Weight,
						sizes[i],
					)
				}
				curDev := math.Abs(sizes[i] - meanSize)
				stdDev += curDev * curDev
			}
			stdDev = math.Sqrt(stdDev/float64(len(sizes))) / meanSize
			if stdDev > tc.maxStdDev {
				t.Errorf("Std too large: %f > %f", stdDev, tc.maxStdDev)
			}
			meanStdDev += stdDev
		}
		meanStdDev /= float64(tc.runs)
		if meanStdDev > tc.maxMeanStdDev {
			t.Errorf("Mean std too large: %f > %f", meanStdDev, tc.maxMeanStdDev)
		}
		t.Logf("Mean std dev: %f", meanStdDev)
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
			hb.AddNode(id, 1)
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
		for _, n := range b.nodes {
			ips := b.buck[n.ID].GetTeamIps()
			for _, ip := range ips {
				beforeByIP[ip] = n.ID
			}
		}

		getCntMoved := func() int {
			cntMoved := 0
			for _, n := range b.nodes {
				ips := b.buck[n.ID].GetTeamIps()
				for _, ip := range ips {
					if n.ID != beforeByIP[ip] {
						cntMoved += 1
					}
					beforeByIP[ip] = n.ID
				}
			}
			return cntMoved
		}

		for i := 0; i < tc.cntDelete; i += 1 {
			toDelete := testutils.RandomInt(0, tc.idCount)
			b.DeleteNode(b.nodes[toDelete].ID)
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
			b.AddNode(id, 1)
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
			hb.AddNode(id, 1)
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
			tc.b.DeleteNode(tid)
		}
		var gotIps []string
		for cid, wantn := range tc.want {
			ipb := tc.b.buck[cid]
			if len(ipb.GetTeamIps()) != wantn {
				t.Errorf("HostBucket.DeleteNode(): incorrent number of gotIps in user bucket want=%d, got=%d",
					wantn, len(ipb.GetTeamIps()),
				)
			}
			gotIps = append(gotIps, ipb.GetTeamIps()...)
		}
		if diff := cmp.Diff(tc.b.ips, gotIps, cmpopts.SortSlices(testutils.LessString)); diff != "" && len(tc.want) > 0 {
			t.Errorf("HostBucket.DeleteNode() summary ips mismatch (-want +got):\n%s", diff)
		}
	}
}
