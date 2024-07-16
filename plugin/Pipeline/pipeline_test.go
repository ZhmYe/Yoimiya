package Pipeline

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"runtime"
	"testing"
)

func Test4Pipeline(t *testing.T) {
	pipelineMaster := NewCoordinator(2, 1)
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	record := pipelineMaster.Process(20, &circuit)
	record.Print()
	runtime.GC()
}
