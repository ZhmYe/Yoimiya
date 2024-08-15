package NormalPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"runtime"
	"testing"
)

func Test4YoimiyaPipelineRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewYoimiyaPipelineRunner(&circuit)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}

func Test4SolverEngine(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	solver := NewSolverEngine(&circuit, 20, 2, 2)
	//solver.InjectTasks(20)
	solver.Start()
}
func Test4ProverEngine(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	prover := NewProverEngine(&circuit, 20, 1, runtime.NumCPU()-2)
	prover.Start()
}
