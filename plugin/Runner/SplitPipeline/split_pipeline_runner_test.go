package SplitPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"runtime"
	"testing"
)

func Test4SolverEngine(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	solver := NewSolverEngine(&circuit, 2, 20, 1, 4)
	//solver.InjectTasks(20)
	solver.Start()
}

func Test4ProverEngine(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	prover := NewProverEngine(&circuit, 2, 20, 1, runtime.NumCPU()-4)
	prover.Start()
}
