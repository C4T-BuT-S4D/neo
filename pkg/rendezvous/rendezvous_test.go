package rendezvous

import (
	"testing"
)

func TestRendezvous_Calculate(t *testing.T) {
	tests := []struct {
		name   string
		nodeID string
		value  string
	}{
		{
			"simple_test",
			"kek",
			"lol",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			hash := r.Calculate(tt.nodeID, tt.value)
			if hash == 0 {
				t.Error("Zero hash returned")
			}

			k := r.getKey(tt.nodeID, tt.value)
			rhash, ok := r.checkCache(k)
			if !ok {
				t.Error("Key is not cached after add")
			} else if rhash != hash {
				t.Errorf("Cached %d, but returned %d", rhash, hash)
			}
		})
	}
}
