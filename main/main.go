package main

import (
	"S-gnark/frontend/cs/r1cs"
	"fmt"
	"math/rand"
	"time"
)

const MAX_N int = 10

func main() {
	rand.Seed(time.Now().Unix())
	//test.Debug()
	fmt.Println(r1cs.R1csStrs[:10])
}
