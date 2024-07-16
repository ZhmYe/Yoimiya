package evaluate

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"runtime"
	"strconv"
)

// Experiment_N_Split_Memory_Reduce
// 实验一: 测试N-Split和不进行split的电路所需的内存占用对比
// n=2,3,4,5
// Figure
// 1. N-split 内存
// 2. N-split 时间结构: Compile + Split + Build + SetUp + Solve
// todo 3. N-split Verify Time
// Table
// 1. Split后的约束、变量
func Experiment_N_Split_Memory_Reduce(option CircuitOption, log bool) {
	NSplitTest := func(cut int, circuit Circuit.TestCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNSplit(cut)
		record.Sprintf(log, "N_Split_Test/"+circuit.Name()+"/"+strconv.Itoa(cut)+"_split", format(circuit.Name(), "n_split"))
	}
	NormalRunningTest := func(circuit Circuit.TestCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_Test/"+circuit.Name()+"/normal_running", format(circuit.Name(), "normal_running"))
	}
	//CircuitList := []CircuitOption{Fib, Mul}
	//for _, option := range CircuitList {
	circuit := getCircuit(option)
	// todo n=2,3,4,5
	nList := []int{2, 3, 4, 5}
	for _, n := range nList {
		NSplitTest(n, circuit, log)
	}
	NormalRunningTest(circuit, log)
	//}
}

// Experiment_N_Split_Memory_Reduce_With_NbLoop 实验二: 测试在某个特定的n下，对电路的约束数量进行改变(修改Loop数量)，对比内存的变化
// n = 2
// Figure
// 1. 不同约束的电路内存对比
// todo 2. 不同约束的电路Verify Time
func Experiment_N_Split_Memory_Reduce_With_NbLoop(option CircuitOption, n int, log bool) {
	NSplitInDifferentNbLoopTest := func(nbLoop int, cut int, circuit Circuit.TestCircuit, log bool) {
		Config.Config.NbLoop = nbLoop
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU())
		record := instance.TestNSplit(cut)
		record.Sprintf(log, "N_Split_nbLoop_Test/"+circuit.Name()+"/nbLoop_"+strconv.Itoa(nbLoop), format(circuit.Name(), "n_split_"+"loop_"+strconv.Itoa(nbLoop)))
	}
	NormalRunningTest := func(nbLoop int, circuit Circuit.TestCircuit, log bool) {
		Config.Config.NbLoop = nbLoop
		runtime.GOMAXPROCS(runtime.NumCPU())
		instance := Instance{circuit: circuit}
		record := instance.TestNormal()
		record.Sprintf(log, "N_Split_nbLoop_Test/"+circuit.Name()+"/nbLoop_"+strconv.Itoa(nbLoop), format(circuit.Name(), "normal_running_"+"loop_"+strconv.Itoa(nbLoop)))
	}
	nbLoopList := []int{1000, 10000, 100000, 500000, 1000000, 5000000, 10000000, 50000000}
	//CircuitList := []CircuitOption{Fib, Mul}
	//for _, option := range CircuitList {
	circuit := getCircuit(option)
	for _, nbLoop := range nbLoopList {
		NSplitInDifferentNbLoopTest(nbLoop, n, circuit, log)
		NormalRunningTest(nbLoop, circuit, log)
	}
	//}
}

// Experiment_MisAligned_Paralleling_With_NbTask 实验三: 测试N-split的流水线并行和串行运行的时间和内存
// n=2,3
// Figure
// 1. 不同任务数下错位并行和串行所需要的时间
// 2. 不同任务数下错位并行和串行所需要的内存
func Experiment_MisAligned_Paralleling_With_NbTask(option CircuitOption, log bool) {
	misalignParallelingTest := func(nbTask int, cut int, circuit Circuit.TestCircuit, log bool) {
		instance := Instance{circuit: circuit}
		runtime.GOMAXPROCS(runtime.NumCPU() / 3 * cut)
		record := instance.TestMisalignedParalleling(nbTask, cut)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "misaligned_paralleling_cut_"+strconv.Itoa(cut)+"_task_"+strconv.Itoa(nbTask)))
	}
	serialRunningTest := func(nbTask int, circuit Circuit.TestCircuit, log bool) {
		runtime.GOMAXPROCS(runtime.NumCPU() / 3)
		instance := Instance{circuit: circuit}
		record := instance.TestSerialRunning(nbTask)
		record.Sprintf(log, "MisalignedParalleling_nbTask_Test/"+circuit.Name()+"/nbTask_"+strconv.Itoa(nbTask), format(circuit.Name(), "serial_running_"+"task_"+strconv.Itoa(nbTask)))
	}
	nbTaskList := []int{2, 4, 8, 16, 32, 64, 128}
	//circuitList := []CircuitOption{Fib, Mul}
	//for _, c := range circuitList {
	circuit := getCircuit(option)
	for _, nbTask := range nbTaskList {
		nList := []int{2, 3}
		for _, cut := range nList {
			misalignParallelingTest(nbTask, cut, circuit, true)
		}
		serialRunningTest(nbTask, circuit, true)
	}
	//}
}
