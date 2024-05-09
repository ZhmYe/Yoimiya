package evaluate

import (
	_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	_ "net/http/pprof"                // 必须，引入 pprof 模块
	"runtime"
	"strconv"
	"testing"
)

// 并发测试
func TestMisalignedParalleling(t *testing.T) {
	misalignParallelingTest := func(nbTask int, cut int, circuit testCircuit, log bool) {
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU())
		record := instance.TestMisalignedParalleling(nbTask, cut)
		record.Sprintf(log, format(circuit.Name(), "misaligned_paralleling_"+"task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
	}
	nbTask := 10
	circuit := getCircuit(Mul)
	misalignParallelingTest(nbTask, 2, circuit, true)
	serialRunningTest(nbTask, circuit, true)
}

// split测试
func TestMemoryReduceByNSplit(t *testing.T) {
	NSplitTest := func(cut int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNSplit(cut)
		record.Sprintf(log, format(circuit.Name(), "n_split"))
	}
	NormalRunningTest := func(circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, format(circuit.Name(), "normal_running"))
	}
	circuit := getCircuit(Mul)
	NSplitTest(2, circuit, true)
	NormalRunningTest(circuit, true)
}

// todo 扩大约束数，查看内存数量减少变化，形成不同电路
// todo 扩大task数，查看misaligned效果
