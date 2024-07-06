package MisalignedParalleling

import (
	"Yoimiya/Circuit/Circuit4MatrixMultiplication"
	"runtime"
	"testing"
)

func Test4Plugin(t *testing.T) {
	for _, n := range []int{112} {
		coordinator := NewCoordinator(1, 1)
		// 简单测试了不同核数串行跑20个fib的prove时间
		// 16 2min25s 145s
		// 20 1min28s 88s
		// 22 1min12s 72s
		// 23 1min5s
		//		nbTask = 20, 65s 37.45s 31s
		// 		nbTask = 128, 4min30s(270s)   3min40s(200s)    3m0.604507019s(180s)
		// 24 59s (37s )
		// 25 55s
		// 30 45s
		// 38 40s
		// 112 34s
		coordinator.SetNumCPUPerSlot(n)
		circuit := Circuit4MatrixMultiplication.NewInterfaceMatrixMultiplicationCircuit()
		go coordinator.MemoryMonitor()
		coordinator.init(&circuit, 1)
		coordinator.Process()
		runtime.GC()
		//fmt.Println(len(receipt))
	}
}
