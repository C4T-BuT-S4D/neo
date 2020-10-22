package hostbucket

import (
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
			b:     New([]string{"ip1", "ip2"}),
			users: []string{"u1", "u2"},
			want:  map[string]int{"u1": 1, "u2": 1},
		},
		{
			b:     New([]string{"ip1", "ip2", "ip3"}),
			users: []string{"u1", "u2"},
			want:  map[string]int{"u1": 2, "u2": 1},
		},
		{
			b:     New([]string{"ip1", "ip2", "ip3", "ip4"}),
			users: []string{"u1", "u2"},
			want:  map[string]int{"u1": 2, "u2": 2},
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
