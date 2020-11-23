package rendezvous

import (
	"testing"
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

			k := CombineKey(tt.nodeID, tt.value)
			rhash, ok := r.checkCache(k)
			if !ok {
				t.Error("Key is not cached after add")
			} else {
				needHash := WeightHash(rhash, tt.nodeWeight)
				if needHash != hash {
					t.Errorf("Cached %f, but returned %f", needHash, hash)
				}
			}
		})
	}
}
