package algos

import "testing"

func TestBinarySearch(t *testing.T) {
	inp := []int{2, 3, 5, 6, 8, 10}

	is(t, BinarySearch(inp, 420), -1)
	is(t, BinarySearch(inp, 6), 3)
}
