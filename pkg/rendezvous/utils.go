package rendezvous

import "math"

const maxUInt64 = float64(^uint64(0))

func CombineKey(nodeId, value string) string {
	return nodeId + ":" + value
}

func WeightHash(hash uint64, weight int) float64 {
	scale := 1.0 / -math.Log(float64(hash)/maxUInt64)
	return scale * float64(weight)
}
