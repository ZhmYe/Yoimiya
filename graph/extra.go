package graph

import "math/rand"

// 这里是对SIT.go的补充，为了避免Merge

// GetMiddleOutputs 返回所有Middle的Stage里最后一个Instruction
// todo 前面的Instruction里的Wire会不会也是Output?
func (t *SITree) GetMiddleOutputs() map[int]int {
	result := make(map[int]int)
	middleStage := t.GetMiddleStage()
	for _, id := range middleStage {
		stage := t.GetStageByIndex(id)
		//for _, iID := range stage.GetInstructions() {
		//	result[iID] = true
		//}
		count := 0
		for _, subStage := range stage.GetChildIDs() {
			if t.GetLayer(subStage) == BOTTOM {
				count++
			}
		}
		result[stage.GetLastInstruction()] = count // 这里为bottom的子stage的数量说明该instruction会被使用几次
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
func (t *SITree) switchTop(stageID int, layer Layer) {
	t.SetLayer(stageID, layer)
	stage := t.GetStageByIndex(stageID)
	for _, subStage := range stage.GetChildIDs() {
		switch t.GetLayer(subStage) {
		case MIDDLE:
			// 如果子节点是MIDDLE，那么其子节点都是BOTTOM，将子节点修改为BOTTOM
			t.SetLayer(subStage, BOTTOM)
			// 其父节点是TOP，此时这些TOP必须修改为MIDDLE
			sStage := t.GetStageByIndex(subStage)
			for _, pID := range sStage.GetParentIDs() {
				// 这里不再重复执行stage本身
				if pID == stageID {
					continue
				}
				t.switchTop(pID, MIDDLE)
			}
		case BOTTOM:
			// 如果子节点是BOTTOM，那么不需要做其它操作
		case TOP:
			// 如果子节点是TOP，那么递归
			t.switchTop(subStage, BOTTOM)
		}
	}
}

func (t *SITree) RandomlySetMiddle(stageID int) {
	epsilon := 1.0
	if rand.Float64() < epsilon {
		t.SetLayer(stageID, MIDDLE)
	} else {
		t.SetLayer(stageID, TOP)
	}
}
