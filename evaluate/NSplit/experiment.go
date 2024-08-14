package NSplit

import (
	"Yoimiya/Circuit"
	"Yoimiya/evaluate"
	"Yoimiya/plugin/Split"
	"runtime"
	"strconv"
	"time"
)

// Experiment_N_Split_Performance
// N-Split Experiment
// 测试N-split在不同电路上的效果

// Experiment_Normal_Performance 测试在不split的情况下，生成zkp所需的内存、时间
func Experiment_Normal_Performance(option Circuit.CircuitOption, log bool) {
	runner := Split.NewGroth16NormalRunner()
	circuit := evaluate.GetCircuit(option)
	_, err := runner.Process(circuit)
	if err != nil {
		panic(err)
	}
	records := runner.Record()
	if log {
		evaluate.PluginRecordLog(records, "NSplitPerformance"+"/"+"record_log_"+evaluate.Format(circuit.Name(), "normal_performance"))

	}
	//evaluate.PluginRecordLog(records, "NSplitPerformance"+"/"+"record_log_"+evaluate.Format(circuit.Name(), "normal_performance"))
}

// Experiment_N_Split_Performance 测试split=n的情况下，生成zkp所需的内存、时间
func Experiment_N_Split_Performance(option Circuit.CircuitOption, log bool) {
	splitList := []int{2, 3, 4, 5}
	for _, s := range splitList {
		runner := Split.NewGroth16SplitRunner(s)
		circuit := evaluate.GetCircuit(option)
		//circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		_, err := runner.Process(circuit)
		if err != nil {
			panic(err)
		}
		records := runner.Record()
		if log {
			evaluate.PluginRecordLog(records, "NSplitPerformance"+"/"+"record_log_"+evaluate.Format(circuit.Name(), strconv.Itoa(s)+"_split_performance"))
		}
		runtime.GC()
		time.Sleep(time.Second * time.Duration(10))
	}
}
