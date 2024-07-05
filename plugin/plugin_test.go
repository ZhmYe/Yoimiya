package plugin

import (
	"Yoimiya/Circuit/Circuit4Fib"
	"fmt"
	"testing"
)

func Test4Plugin(t *testing.T) {
	coordinator := NewCoordinator(3, 1)
	// 简单测试了不同核数串行跑20个fib的prove时间
	// 16 2min25s 145s
	// 20 1min28s 88s
	// 22 1min12s 72s
	// 23 1min5s  65s (39s 31s)
	// 24 59s (37s )
	// 25 55s
	// 30 45s
	// 38 40s
	// 112 34s
	coordinator.SetNumCPUPerSlot(23)
	circuit := Circuit4Fib.NewLoopFibonacciCircuit()
	go coordinator.MemoryMonitor()
	coordinator.init(&circuit, 20)
	receipt := coordinator.Process()
	fmt.Println(len(receipt))
}
