package NormalPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
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
	solver := NewSolverEngine(&circuit)
	solver.InjectTasks(20)
	solver.Start()
}
func Test4ProverEngine(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	prover := NewProverEngine(&circuit)
	prover.Start()
}
