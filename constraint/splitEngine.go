package constraint

import (
	"S-gnark/Config"
	"S-gnark/graph/PackedLevel"
	"S-gnark/graph/Sit"
)

type SplitEngine interface {
	GetLayersInfo() [4]int
	//CheckAndGetSubCircuitStageIDs() ([]int, []int)
	Insert(iID int, previousIDs []int)
	AssignLayer()
	//GetInstructionIdsFromStageIDs(ids []int) []int
	GetSubCircuitInstructionIDs() ([]int, []int)
	GetStageNumber() int
	GetMiddleOutputs() map[int]bool
	GetAllInstructions() []int
}

func InitSplitEngine() SplitEngine {
	var s SplitEngine
	switch Config.Config.Split {
	case Config.SPLIT_STAGES:
		s = Sit.NewSITree()
		//s = sit
	case Config.SPLIT_LEVELS:
		s = PackedLevel.NewPackedLevel()
	}
	return s
}
