package graph

import (
	"testing"
)

func GenerateDAG() (*DAG, *DAG, []int) {
	forward := NewDAG()
	forward.Update(1, 2)
	forward.Update(1, 3)
	forward.Update(2, 4)
	forward.Update(2, 5)
	forward.Update(3, 5)
	forward.Update(3, 6)
	forward.Update(4, 7)

	backward := NewDAG()
	backward.Update(7, 4)
	backward.Update(5, 2)
	backward.Update(5, 3)
	backward.Update(6, 3)
	backward.Update(4, 2)
	backward.Update(2, 1)
	backward.Update(3, 1)

	lastLevel := []int{7, 5, 6}
	return forward, backward, lastLevel
}
func TestSplit(t *testing.T) {
	forward, backward, lastLevel := GenerateDAG()
	splitEngine := NewSplitEngine(forward, backward, lastLevel)
	splitEngine.Split()
	splitEngine.PrintStages()
}
