package Parallel

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"runtime"
	"testing"
)

func Test4Parallel(t *testing.T) {
	parallelMaster := NewParallelMaster(3)
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	record := parallelMaster.Process(20, &circuit)
	record.Print()
	runtime.GC()
	//fmt.Println(len(receipt))
}
