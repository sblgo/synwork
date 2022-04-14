package utils

import (
	"testing"
)

func TestMapArray01(t *testing.T) {
	r := MapArray[int, float32]([]int{1, 2, 3}, []float32{}, func(i int) float32 { return 2.3 * float32(i) })
	if len(r) != 3 {
		t.Fatalf("result has unexpected length %d", len(r))
	}
	t.Logf("maparray %v", r)
}
