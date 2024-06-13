package constraint

import (
	"Yoimiya/Config"
	"Yoimiya/graph/LayeredGraph"
	"Yoimiya/graph/Sit"
)

type SplitEngine interface {
	GetLayersInfo() [4]int
	//CheckAndGetSubCircuitStageIDs() ([]int, []int)

	Insert(iID int, previousIDs []int)
	AssignLayer(cut int)
	//GetInstructionIdsFromStageIDs(ids []int) []int

	GetSubCircuitInstructionIDs() [][]int
	GetStageNumber() int
	GetMiddleOutputs() map[int]bool
	GetAllInstructions() []int
	IsMiddle(iID int) bool
}

func InitSplitEngine() SplitEngine {
	var s SplitEngine
	switch Config.Config.Split {
	case Config.SPLIT_STAGES:
		s = Sit.NewSITree()
	//s = sit
	//case Config.SPLIT_LEVELS:
	//	s = PackedLevel.NewPackedLevel()
	case Config.SPLIT_LEVELS:
		s = LayeredGraph.NewLayeredGraph()

	}
	return s
}
