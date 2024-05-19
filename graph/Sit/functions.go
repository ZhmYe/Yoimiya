package Sit

import (
	"fmt"
	"strconv"
)

// GetStageByIndex 获取对应位置的Stage
func (t *SITree) GetStageByIndex(index int) *Stage {
	return t.stages[index]
}

// GetStageByInstruction 获得Instruction对应的Stage
func (t *SITree) GetStageByInstruction(iId int) *Stage {
	_, exist := t.index[iId]
	if !exist {
		fmt.Println(len(t.index))
		panic("Don't have such Instruction!!!" + strconv.Itoa(iId))
	}
	return t.index[iId]
}

// updateIndex 更新索引
func (t *SITree) updateIndex(iId int, stage *Stage) {
	t.index[iId] = stage
}
func (t *SITree) batchUpdateIndex(ids []int, stage *Stage) {
	for _, id := range ids {
		t.updateIndex(id, stage)
	}
}
func (t *SITree) GetRootStages() []*Stage {
	//for _, stage := range t.root {
	//	fmt.Println(stage.GetParentIDs(), stage.GetChildIDs(), stage.GetInstructions())
	//}
	return t.root
}
func (t *SITree) GetStageNumber() int {
	return len(t.stages)
}
func (t *SITree) GetTotalInstructionNumber() int {
	return len(t.instructions)
}

func (t *SITree) GetStages() []*Stage {
	return t.stages
}
func (t *SITree) GetEdges() int {
	total := 0
	for _, stage := range t.stages {
		total += len(stage.GetSubStages())
	}
	return total
}

func (t *SITree) ModifyiID(stageIndex int, instructionIndex int, iID int) {
	t.stages[stageIndex].Instructions[instructionIndex] = iID
}

func (t *SITree) GetParents(stageId int) []int {
	return t.GetStageByIndex(stageId).GetParentIDs()
}
func (t *SITree) GetParentsMap(stageId int) map[int]bool {
	parents := t.GetParents(stageId)
	result := make(map[int]bool)
	for _, id := range parents {
		result[id] = true
	}
	return result
}
func (t *SITree) HasParent(stageId int) bool {
	return len(t.GetParents(stageId)) != 0
}
func (t *SITree) GetSubCircuitInstructionIDs() [][]int {
	top, bottom := t.CheckAndGetSubCircuitStageIDs()
	return [][]int{t.GetInstructionIdsFromStageIDs(top), t.GetInstructionIdsFromStageIDs(bottom)}
}

// CheckAndGetSubCircuitStageIDs 获得各种类型的Layer对应的stage id列表
// 在SIT创建的时候，已经保证sit.stages中的结果是拓扑意义上有序的 todo ?
// 因此，现在按序遍历时也已保证有序
func (t *SITree) CheckAndGetSubCircuitStageIDs() ([]int, []int) {
	top := make([]int, 0) // middle的也放到top里
	//middle := make([]int, 0)
	bottom := make([]int, 0)
	for sID, layer := range t.layers {
		if layer == TOP {
			top = append(top, sID)
		} else if layer == MIDDLE {
			top = append(top, sID)
		} else if layer == BOTTOM {
			bottom = append(bottom, sID)
		} else {
			panic("stage " + strconv.Itoa(sID) + " hasn't set layer!!!")
		}
	}
	return top, bottom
}

// GetInstructionsFromStages 根据给定的有序Stage列表，返回有序的Instruction列表
func (t *SITree) GetInstructionsFromStages(sIDs ...int) []int {
	result := make([]int, 0)
	for _, sID := range sIDs {
		stage := t.GetStageByIndex(sID)
		result = append(result, stage.GetInstructions()...)
	}
	return result
}

func (t *SITree) CountLeafNode() (total int) {
	for _, stage := range t.GetStages() {
		if len(stage.GetSubStages()) == 0 {
			total++
		}
	}
	return
}

// GetMiddleOutputs 返回所有Middle的Stage里最后一个Instruction
// todo 前面的Instruction里的Wire会不会也是Output?
func (t *SITree) GetMiddleOutputs() map[int]bool {
	result := make(map[int]bool)
	middleStage := t.GetMiddleStage()
	for _, id := range middleStage {
		stage := t.GetStageByIndex(id)
		for _, iID := range stage.GetInstructions() {
			result[iID] = true
		}
		//count := 0
		//for _, subStage := range stage.GetChildIDs() {
		//	if t.GetLayer(subStage) == BOTTOM {
		//		count++
		//	}
		//}
		//result[stage.GetLastInstruction()] = count // 这里为bottom的子stage的数量说明该instruction会被使用几次
	}
	return result
}

// GetInstructionIdsFromStageIDs 得到的stageIDs是严格有序的，stage里的instruction以数组形式排列有序
// todo 目前stageIDs不是严格有序的，但instructions有严格有序的
func (t *SITree) GetInstructionIdsFromStageIDs(stageIDs []int) []int {
	iIDs := make([]int, 0)
	temp := make(map[int]bool)
	for _, stageID := range stageIDs {
		stage := t.GetStageByIndex(stageID)
		for _, id := range stage.GetInstructions() {
			temp[id] = true
		}
		//iIDs = append(iIDs, stage.GetInstructions()...)
	}
	for _, id := range t.instructions {
		_, exist := temp[id]
		if exist {
			iIDs = append(iIDs, id)
		}
	}
	return iIDs
}
func (t *SITree) GetAllInstructions() []int {
	result := make([]int, 0)
	for _, stage := range t.stages {
		result = append(result, stage.Instructions...)
	}
	return result
}
func (t *SITree) IsMiddle(iID int) bool {
	return t.layers[iID] == MIDDLE
}
