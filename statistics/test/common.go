package test

import (
	"math/rand"
)

const MaxN int = 10

func generateSamples(N int) []int {
	result := make([]int, 0)
	for i := 0; i < N; i++ {
		result = append(result, rand.Intn(100))
	}
	return result
}
