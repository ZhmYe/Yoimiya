package graph

import (
	"fmt"
	"testing"
)

func GenerateSIT(option int) *SITree {
	t := NewSITree()
	switch option {
	case 1:
		t.Insert(1, []int{})
		t.Insert(2, []int{1})
		t.Insert(3, []int{1})
		t.Insert(4, []int{2})
		t.Insert(5, []int{2, 3})
		t.Insert(6, []int{3})
		t.Insert(7, []int{4})
	case 2:
		t.Insert(1, []int{})
		t.Insert(2, []int{1})
		t.Insert(3, []int{1})
		t.Insert(4, []int{2, 3})
		t.Insert(5, []int{4})
		t.Insert(6, []int{4})
		t.Insert(7, []int{5, 6})
		t.Insert(8, []int{7})
		t.Insert(9, []int{7})
		t.Insert(10, []int{8, 9})
		t.Insert(11, []int{10})
		t.Insert(12, []int{10})
		t.Insert(13, []int{11, 12})
		t.Insert(14, []int{13})
		t.Insert(15, []int{13})
	case 3:
		t.Insert(1, []int{})
		t.Insert(2, []int{1})
		t.Insert(3, []int{1})
		t.Insert(4, []int{2})
		t.Insert(5, []int{2})
		t.Insert(6, []int{4, 5})
		t.Insert(7, []int{6})
		t.Insert(10, []int{3})
		t.Insert(8, []int{6, 10})
		t.Insert(9, []int{7, 8})
	case 4:
		t.Insert(1, []int{})
		t.Insert(2, []int{1})
		t.Insert(3, []int{1})
		t.Insert(4, []int{2})
		t.Insert(5, []int{2, 3})
		t.Insert(6, []int{4, 5})
		t.Insert(7, []int{6})
		t.Insert(8, []int{6})
		t.Insert(9, []int{7, 8})

	default:
		t.Insert(1, []int{})
	}
	return t
}

func TestMemoryReduceSplit(t *testing.T) {
	options := []int{1, 2, 3, 4}
	for _, opt := range options {
		t.Log("Test SITree", opt)
		sit := GenerateSIT(opt)
		formerSIT, latterSIT := sit.HeuristicSplit()
		fmt.Println("Former SITree")
		fmt.Println("Root Stages:")
		for _, stage := range formerSIT.GetRootStages() {
			fmt.Println(stage.GetInstructions())
		}
		fmt.Println("All Stages:")
		for _, stage := range formerSIT.GetStages() {
			fmt.Println(stage.GetInstructions())
		}
		fmt.Println("Latter SITree")
		fmt.Println("Root Stages:")
		for _, stage := range latterSIT.GetRootStages() {
			fmt.Println(stage.GetInstructions())
		}
		fmt.Println("All Stages:")
		for _, stage := range latterSIT.GetStages() {
			fmt.Println(stage.GetInstructions())
		}
	}
}
