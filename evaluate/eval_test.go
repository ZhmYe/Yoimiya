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

// split测试
func TestMemoryReduceByNSplit(t *testing.T) {
	Experiment_N_Split_Memory_Reduce(Matrix, false)
	//Experiment_N_Split_Memory_Reduce(Fib, false)
	//Experiment_N_Split_Memory_Reduce(Mul, true)
}

// todo 扩大约束数，查看内存数量减少变化，形成不同电路
func TestMemoryReduceInDifferentNbLoop(t *testing.T) {
	Experiment_N_Split_Memory_Reduce_With_NbLoop(Fib, 2, true)
	Experiment_N_Split_Memory_Reduce_With_NbLoop(Mul, 2, true)
}

// todo 扩大task数，查看misaligned效果
func TestMisalignedParallelingInDifferentNbTasks(t *testing.T) {
	Experiment_MisAligned_Paralleling_With_NbTask(Fib, true)
	Experiment_MisAligned_Paralleling_With_NbTask(Mul, true)

}
