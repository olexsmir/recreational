package algos

func BubbleSort(inp []int) {
	for i := 0; i < len(inp); i++ {
		for j := 0; j < len(inp)-1-i; j++ {
			if inp[j] > inp[j+1] {
				inp[j], inp[j+1] = inp[j+1], inp[j]
			}
		}
	}
}
