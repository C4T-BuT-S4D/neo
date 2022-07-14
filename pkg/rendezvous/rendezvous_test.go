package rendezvous

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRendezvous_Calculate(t *testing.T) {
	tests := []struct {
		name       string
		nodeID     string
		nodeWeight int
		value      string
	}{
		{
			"simple_test",
			"kek",
			1,
			"lol",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			hash := r.Calculate(tt.nodeID, tt.nodeWeight, tt.value)
			if hash == 0 {
				t.Error("Zero hash returned")
			}

			k := combineKey(tt.nodeID, tt.value)
			rhash, ok := r.checkCache(k)
			require.True(t, ok)
			needHash := weightHash(rhash, tt.nodeWeight)
			require.Equal(t, needHash, hash)
		})
	}
}
