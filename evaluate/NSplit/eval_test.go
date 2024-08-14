package NSplit

import (
	"Yoimiya/Circuit"
	"testing"
)

func Test4NormalRunningPerformance(t *testing.T) {
	Experiment_Normal_Performance(Circuit.FibSquare, true)
}
func Test4NSplitPerformance(t *testing.T) {
	Experiment_N_Split_Performance(Circuit.FibSquare, true)
}