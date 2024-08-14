package SplitPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

// Test Result
// GOMAXCPU = 16
// Serial 2min43s, 163s
// Pipeline 1min53s, 113s
// SplitPipeline 1min45s, 105s

func Test4Groth16SerialRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewGroth16SerialRunner(&circuit)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}

func Test4Groth16PipelineRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewGroth16PipelineRunner(&circuit)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}

func Test4Groth16SplitPipelineRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewGroth16SplitPipelineRunner(&circuit, 2, 1)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}
