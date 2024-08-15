package Serial

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"testing"
)

func Test4YoimiyaSerialRunner(t *testing.T) {
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	runner := NewYoimiyaSerialRunner(&circuit)
	runner.InjectTasks(20)
	runner.Process()
	runner.Record()
}
