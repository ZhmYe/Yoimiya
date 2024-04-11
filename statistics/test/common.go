package test

import (
	"Yoimiya/frontend"
	"math/rand"
)

const MaxN int = 10

func generateSamples(N int) ([10000]frontend.Variable, []int) {
	resultVariable := new([10000]frontend.Variable)
	resultInt := make([]int, 0)
	for i := 0; i < N; i++ {
		element := rand.Intn(100)
		resultVariable[i] = element
		resultInt = append(resultInt, element)
	}
	return *resultVariable, resultInt
}
