package evaluate

import (
	"Yoimiya/Config"
	_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	_ "net/http/pprof"                // 必须，引入 pprof 模块
	"runtime"
	"strconv"
	"sync"
	"testing"
)

// 并发测试
func TestMisalignedParalleling(t *testing.T) {
	misalignParallelingTest := func(nbTask int, cut int, circuit testCircuit, log bool) {
		// 这里为了模拟的场景是，一个电路给一定数量的cpu
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU() / 3 * cut)
		record := instance.TestMisalignedParalleling(nbTask, cut)
		record.Sprintf(log, "MisalignedParallelingTest/"+circuit.Name()+strconv.Itoa(cut)+"_split", format(circuit.Name(), "misaligned_paralleling_"+"task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 3)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, "MisalignedParallelingTest/"+circuit.Name()+"/serial_running", format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
	}
	nbTask := 4
	circuit := getCircuit(Fib)
	cutList := []int{2, 3}
	for _, cut := range cutList {
		misalignParallelingTest(nbTask, cut, circuit, false)
	}
	serialRunningTest(nbTask, circuit, false)
}
func TestTmp(t *testing.T) {
	NormalRunningTest := func(circuit testCircuit, log bool) {
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_Test/"+circuit.Name()+"/normal_running", format(circuit.Name(), "normal_running"))
	}
	Config.Config.NbLoop = 3333
	runtime.GOMAXPROCS(108)
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			NormalRunningTest(getCircuit(Fib), false)
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GOMAXPROCS(108)
	Config.Config.NbLoop = 10000
	NormalRunningTest(getCircuit(Fib), false)
}

// split测试
func TestMemoryReduceByNSplit(t *testing.T) {
	NSplitTest := func(cut int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNSplit(cut)
		record.Sprintf(log, "N_Split_Test/"+circuit.Name()+"/"+strconv.Itoa(cut)+"_split", format(circuit.Name(), "n_split"))
	}
	NormalRunningTest := func(circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_Test/"+circuit.Name()+"/normal_running", format(circuit.Name(), "normal_running"))
	}
	circuit := getCircuit(Fib)
	// todo n=2,3,4,5
	nList := []int{2, 3, 4, 5}
	for _, n := range nList {
		NSplitTest(n, circuit, false)
	}
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
		runtime.GOMAXPROCS(runtime.NumCPU() / 3 * cut)
		record := instance.TestMisalignedParalleling(nbTask, cut)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "misaligned_paralleling_cut_"+strconv.Itoa(cut)+"_task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit testCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 3)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
	}
	nbTaskList := []int{2, 4, 8, 16, 32, 64, 128}
	circuitList := []CircuitOption{Fib, Mul}
	for _, c := range circuitList {
		circuit := getCircuit(c)
		for _, nbTask := range nbTaskList {
			nList := []int{2, 3}
			for _, cut := range nList {
				misalignParallelingTest(nbTask, cut, circuit, true)
			}
			serialRunningTest(nbTask, circuit, true)
		}
	}

}
