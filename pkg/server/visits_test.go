package server

import (
	"testing"
	"time"

	"neo/pkg/testutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestVisitsMap_Invalidate(t *testing.T) {
	generate := func(m map[string]time.Time) *visitsMap {
		vm := newVisitsMap()
		vm.visits = m
		return vm
	}
	for _, tc := range []struct {
		vm   *visitsMap
		dur  string
		now  time.Time
		want []string
	}{
		{
			vm:   generate(map[string]time.Time{"1": time.Unix(10, 0), "2": time.Unix(30, 0)}),
			dur:  "5s",
			now:  time.Unix(35, 0),
			want: []string{"1"},
		},
		{
			vm:   generate(map[string]time.Time{"1": time.Unix(10, 0), "2": time.Unix(30, 0)}),
			dur:  "50s",
			now:  time.Unix(35, 0),
			want: []string{},
		},
		{
			vm:   generate(map[string]time.Time{"1": time.Unix(10, 0), "2": time.Unix(30, 0)}),
			dur:  "1s",
			now:  time.Unix(35, 0),
			want: []string{"1", "2"},
		},
		{
			vm:   generate(map[string]time.Time{"1": time.Unix(10, 0), "2": time.Unix(30, 0)}),
			dur:  "5s",
			now:  time.Unix(15, 0),
			want: []string{},
		},
	} {
		d, _ := time.ParseDuration(tc.dur)
		invalid := tc.vm.Invalidate(tc.now, d)
		if diff := cmp.Diff(tc.want, invalid, cmpopts.SortSlices(testutils.LessString), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Invalidate() ids mismatch (-want +got):\n%s", diff)
		}
	}
}
