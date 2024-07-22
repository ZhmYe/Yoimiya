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
	// p = 1,串行
	// (nbTask=20, 33s)
	// p = 2
	// (nbTask=20, 21s)
	// p = 3
	// (nbTask=20, 15s)
	record.Print()
	runtime.GC()
	//fmt.Println(len(receipt))
}
