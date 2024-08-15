package NormalPipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

func Test4YoimiyaPipelineRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewYoimiyaPipelineRunner(&circuit)
	runner.InjectTasks(4)
	runner.Process()
	runner.Record()
}
