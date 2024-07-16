package Split

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

func Test_Groth16_Normal_Runner(t *testing.T) {
	runner := NewGroth16NormalRunner()
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	_, err := runner.Process(&circuit)
	if err != nil {
		panic(err)
	}
	runner.Record()
}