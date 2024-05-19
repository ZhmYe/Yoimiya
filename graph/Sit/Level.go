package Sit

// ComputeDepth 计算某个stage的深度，由其所有父节点的深度最大值+1
func (t *SITree) ComputeDepth(stage *Stage) {
	if stage.GetCount() == 0 {
		// 是rootStage，深度初始化为0
		t.depth = append(t.depth, 0)
		//t.depth[stage.GetID()] = 1
	} else {
		// 遍历其所有父节点
		maxD := -1
		for _, parent := range stage.GetParentIDs() {
			parentDepth := t.GetDepth(parent)
			if parentDepth > maxD {
				maxD = parentDepth
			}
		}
		t.depth = append(t.depth, maxD+1)
		// 更新最大深度
		if maxD+1 > t.maxDepth {
			t.maxDepth = maxD + 1
		}
	}
}
func (t *SITree) GetDepth(stageID int) int {
	return t.depth[stageID]
}

// GenerateLEVEL 生成LEVEL
func (t *SITree) GenerateLEVEL() [][]int {
	LEVEL := make([][]int, t.maxDepth+1)
	for i := 0; i < t.maxDepth; i++ {
		LEVEL[i] = make([]int, 0)
	}
	for stageID, depth := range t.depth {
		LEVEL[depth] = append(LEVEL[depth], stageID)
	}
	return LEVEL
}
func (t *SITree) AssignLayer(cut int) {
	LEVEL := t.GenerateLEVEL()
	totalStageNumber := t.GetStageNumber()
	total := 0
	splitDepth := -1
	// 遍历所有level
	for i := 0; i < t.maxDepth; i++ {
		level := LEVEL[i]
		if 2*(total+len(level)) > totalStageNumber {
			splitDepth = i
			break
		}
		total += len(level)
	}
	// todo 这里可以把最后一层再简单划分一下?
	// 在得到了划分的LEVEL位置后，遍历所有的stage，判断TOP/MIDDLE/BOTTOM
	for _, stage := range t.stages {
		depth := t.GetDepth(stage.GetID()) // 得到stage的level
		// 如果深度比划分的LEVEL位置大，那么是BOTTOM
		if depth > splitDepth {
			t.SetLayer(stage.GetID(), BOTTOM)
		} else {
			// 如果深度比划分的LEVEL小
			// 判断stage的所有子节点，是否有超过splitLevel的，如果有，则为MIDDLE
			flag := false
			for _, id := range stage.GetChildIDs() {
				subDepth := t.GetDepth(id)
				if subDepth > splitDepth {
					flag = true
					break
				}
			}
			if flag {
				t.SetLayer(stage.GetID(), MIDDLE)
			} else {
				t.SetLayer(stage.GetID(), TOP)
			}
		}
	}
}
