package Split

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

func Test_Groth16_Normal_Runner(t *testing.T) {
	runner := NewGroth16NormalRunner()
	//circuit := Circuit4MatrixMultiplication.NewInterfaceMatrixMultiplicationCircuit()
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	_, err := runner.Process(&circuit)
	if err != nil {
		panic(err)
	}
	runner.Record()

}
func Test_Groth16_Split_Runner(t *testing.T) {
	runner := NewGroth16SplitRunner(2)
	//circuit := Circuit4MatrixMultiplication.NewInterfaceMatrixMultiplicationCircuit()
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	_, err := runner.Process(&circuit)
	if err != nil {
		panic(err)
	}
	runner.Record()
}
