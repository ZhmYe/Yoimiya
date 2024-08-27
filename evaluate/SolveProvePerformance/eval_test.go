package SolveProvePerformance

import (
	"Yoimiya/Circuit"
	"testing"
)

func Test4SolveProveTimePerformance(t *testing.T) {
	Experiment_Solve_Prove_Time_Performance(Circuit.FibSquare, true)
}
func Test4SolveProveCPUPerformance(t *testing.T) {
	Experiment_Solve_Prove_CPU_Performance(Circuit.FibSquare, true)
}
func Test4SolveProveMemoryPerformance(t *testing.T) {
	Experiment_Solve_Prove_Memory_Performance(Circuit.FibSquare, false)
}
