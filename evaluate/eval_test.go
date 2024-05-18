package evaluate

import (
	"Yoimiya/Config"
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
		record.Sprintf(log, "MisalignedParallelingTest/"+circuit.Name(), format(circuit.Name(), "misaligned_paralleling_"+"task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, "MisalignedParallelingTest/"+circuit.Name(), format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
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
		record.Sprintf(log, "N_Split_Test/"+circuit.Name(), format(circuit.Name(), "n_split"))
	}
	NormalRunningTest := func(circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_Test/"+circuit.Name(), format(circuit.Name(), "normal_running"))
	}
	circuit := getCircuit(Fib)
	NSplitTest(2, circuit, true)
	NormalRunningTest(circuit, true)
}

// todo 扩大约束数，查看内存数量减少变化，形成不同电路
func TestMemoryReduceInDifferentNbLoop(t *testing.T) {
	NSplitInDifferentNbLoopTest := func(nbLoop int, cut int, circuit testCircuit, log bool) {
		Config.Config.NbLoop = nbLoop
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU())
		record := instance.TestNSplit(cut)
		record.Sprintf(log, "N_Split_nbLoop_Test/"+circuit.Name()+"/nbLoop_"+strconv.Itoa(nbLoop), format(circuit.Name(), "n_split_"+"loop_"+strconv.Itoa(nbLoop)))
	}
	NormalRunningTest := func(nbLoop int, circuit testCircuit, log bool) {
		Config.Config.NbLoop = nbLoop
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_nbLoop_Test/"+circuit.Name()+"/nbLoop_"+strconv.Itoa(nbLoop), format(circuit.Name(), "normal_running_"+"loop_"+strconv.Itoa(nbLoop)))
	}
	nbLoopList := []int{1000, 10000, 100000, 500000, 1000000, 5000000, 10000000, 50000000}
	circuit := getCircuit(Mul)
	for _, nbLoop := range nbLoopList {
		NSplitInDifferentNbLoopTest(nbLoop, 2, circuit, true)
		NormalRunningTest(nbLoop, circuit, true)
	}
}

// todo 扩大task数，查看misaligned效果
func TestMisalignedParallelingInDifferentNbTasks(t *testing.T) {
	misalignParallelingTest := func(nbTask int, cut int, circuit testCircuit, log bool) {
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU())
		record := instance.TestMisalignedParalleling(nbTask, cut)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "misaligned_paralleling_"+"task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
	}
	nbTaskList := []int{2, 4, 8, 16, 32, 64, 128}

	circuit := getCircuit(Mul)
	for _, nbTask := range nbTaskList {
		misalignParallelingTest(nbTask, 2, circuit, true)
		serialRunningTest(nbTask, circuit, true)
	}
}
