package main

import (
	"Yoimiya/evaluate"
	"math/rand"
	"time"
)

const MAX_N int = 10

func main() {
	rand.Seed(time.Now().Unix())
	//test.Debug()
	//evaluate.TestRunTimeInDifferentSplitMethod()
	evaluate.TestMemoryReduce()
}
