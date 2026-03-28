package algos

import (
	"reflect"
	"testing"
)

func TestQuicksort(t *testing.T) {
	inp := []int{9, 3, 7, 4, 69, 420, 42}
	out := []int{3, 4, 7, 9, 42, 69, 420}

	QuickSort(inp)
	if !reflect.DeepEqual((inp), out) {
		t.Fatalf("i fucked up")
	}
}
