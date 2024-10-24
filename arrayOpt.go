package researd

import "math/rand"

func removeValue(arr []string, value string) []string {
	result := []string{}
	for _, val := range arr {
		if val != value {
			result = append(result, value)
		}
	}
	return result
}

func shuffleArray(arr []string) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i] // 交换元素
	})
}
