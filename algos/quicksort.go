package algos

func partition(inp []int, lo, hi int) int {
	pivot := inp[hi] // might result in O(n^2)

	idx := lo - 1
	for i := lo; i < hi; i++ {
		if inp[i] <= pivot {
			idx++
			inp[i], inp[idx] = inp[idx], inp[i]
		}
	}

	idx++
	inp[hi], inp[idx] = inp[idx], pivot
	return idx
}

func qs(inp []int, lo, hi int) {
	if lo >= hi {
		return
	}

	pivot := partition(inp, lo, hi)
	qs(inp, lo, pivot-1)
	qs(inp, pivot+1, hi)
}

func QuickSort(input []int) { qs(input, 0, len(input) - 1) }
