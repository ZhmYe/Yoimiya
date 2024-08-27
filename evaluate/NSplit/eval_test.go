package NSplit

import (
	"Yoimiya/Circuit"
	"testing"
)

func Test4NormalRunningPerformance(t *testing.T) {
	Experiment_Normal_Performance(Circuit.Fib, true)
}
func Test4NSplitPerformance(t *testing.T) {
	Experiment_N_Split_Performance(Circuit.Fib, true)
}
func Test4NormalRunningSizePerformance(t *testing.T) {
	//Experiment_Graph_Size_Normal_Performance_Fib(Circuit.FibSquare, true)
	Experiment_Graph_Size_Normal_Performace(Circuit.Matrix, true)
}
func Test4NSplitSizePerformance(t *testing.T) {
	//Experiment_Graph_Size_Performance_Fib(Circuit.FibSquare, true)
	Experiment_Graph_Size_Performance_Split(Circuit.Matrix, true)
}
