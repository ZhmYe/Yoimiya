package constraint

import (
	"Yoimiya/Config"
	LRO_Tree "Yoimiya/graph/LRO-Tree"
	"Yoimiya/graph/PackedLevel"
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
	GenerateSplitWitness() [][]int
}

func InitSplitEngine() SplitEngine {
	var s SplitEngine
	if Config.Config.IsSplit() {
		switch Config.Config.Split {
		case Config.SPLIT_STAGES:
			s = Sit.NewSITree()
		//s = sit
		case Config.SPLIT_LEVELS:
			s = PackedLevel.NewPackedLevel()
		//case Config.SPLIT_LEVELS:
		//	s = LayeredGraph.NewLayeredGraph()
		case Config.SPLIT_LRO:
			s = LRO_Tree.NewLroTree()
		}
	} else {
		s = PackedLevel.NewPackedLevel()
	}
	return s
}
