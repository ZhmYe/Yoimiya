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
func Test4SolvePerformance(t *testing.T) {
	Experiment_Solve_Performance(Circuit.FibSquare, false)
}
