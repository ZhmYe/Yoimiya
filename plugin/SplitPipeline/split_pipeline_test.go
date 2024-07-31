package SplitPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

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
	runner := NewGroth16SplitPipelineRunner(&circuit, 2)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}
