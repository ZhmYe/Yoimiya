package graph

import (
	"fmt"
	"sync"
)

// SITree Stage-Instruction Tree
type SITree struct {
	stages       []*Stage       // SITree节点
	instructions []int          // 只记录instruction id
	index        map[int]*Stage // 索引，iID -> Stage
	root         []*Stage       // 所有没有父stage的stage
}

func NewSITree() *SITree {
	t := new(SITree)
	t.stages = make([]*Stage, 0)
	t.instructions = make([]int, 0)
	t.index = make(map[int]*Stage)
	t.root = make([]*Stage, 0)
	return t
}
func (t *SITree) appendStage(stage *Stage) {
	stage.SetID(len(t.stages)) // 只有当真正append的时候才会添加id，确保id和index对应
	t.stages = append(t.stages, stage)
	t.batchUpdateIndex(stage.GetInstructions(), stage)
	//for _, iID := range stage.GetInstructions() {
	//	t.index[iID] = stage.GetID()
	//}
}
func (t *SITree) appendRoot(stage *Stage) {
	t.root = append(t.root, stage)
}

// Insert 插入一个instruction时会伴有它的所有前置一阶instruction(previousId)
// 首先将当前instruction放在一个新的stage中
// 然后首先判断当前instruction是否有多个父节点
// 如果是，为宽依赖，将新建的stage作为child连接到所有父stage之后
//	case1: 父stage的最后一个instruction就是父instruction，那么直接append
//	case2: 父stage的最后一个instruction不是父instruction，那么在此之前，已有instruction被认为是窄依赖添加到了
//	父instruction之后,此时需要进行分裂(split)，将父instruction之后的所有instruction放在一个新的stage中，然后
//	和当前stage一起作为child连接到父stage中
// 如果不是，那么判断父节点是否有多个子节点，同上述case1,case2，区别在如果是窄依赖直接combine

func (t *SITree) Insert(iID int, previousIds []int) {
	stage := NewStage(-1, iID) // id统一都默认初始化为-1，在append时处理
	// 如果没有父节点
	if len(previousIds) == 0 {
		t.appendStage(stage) // 直接append
		t.appendRoot(stage)  // 没有父节点则一定是root Stage
	} else if len(previousIds) == 1 {
		// 如果只有一个父节点
		// 暂时认为当前instruction和父instruction之间是窄依赖关系，合并stage
		previousId := previousIds[0]
		parentStage := t.GetStageByInstruction(previousId) // 得到父stage
		if t.checkSplit(parentStage, previousId) {
			// 需要分裂，那么最终为宽依赖
			t.Split(parentStage, previousId)
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
		} else if len(parentStage.GetSubStages()) != 0 {
			// 如果不需要分裂，但父stage有多个子stage，那么也是宽依赖
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
		} else {
			// 无需分裂，并且父stage当前没有子stage，那么暂时认为是窄依赖
			t.Combine(stage, parentStage)
		}
	} else {
		// 如果不止有一个父节点，一定是宽依赖
		// 首先把stage append
		t.appendStage(stage)
		hasBeenChild := make(map[int]bool) // 可能会出现多个previousId在同一个stage里面，那么无需后续重新添加child
		//parentStages := make(map[int]map[int]bool)
		//for _, previousId := range previousIds {
		//	_, exist := parentStages[t.GetStageByInstruction(previousId).GetID()]
		//	if exist {
		//		parentStages[t.GetStageByInstruction(previousId).GetID()][previousId] = true
		//	} else {
		//		parentStages[t.GetStageByInstruction(previousId).GetID()] = make(map[int]bool)
		//	}
		//}
		//for parentStageId, ids := range parentStages {
		//	parentStage := t.GetStageByIndex(parentStageId)
		//	if t.checkSplit(parentStage, ids) {
		//		t.Split(parentStage, ids)
		//	}
		//}
		for _, previousId := range previousIds {
			parentStage := t.GetStageByInstruction(previousId)
			var fission *Stage
			hasSplit := false
			// 判断是否需要split
			if t.checkSplit(parentStage, previousId) {
				// 需要分裂
				fission = t.Split(parentStage, previousId)
				hasSplit = true
			}
			_, flag := hasBeenChild[parentStage.GetID()]
			if flag {
				if hasSplit {
					// 分裂的两个stage都是当前stage的父stage
					fission.AddChild(stage)
					stage.AddParent(fission)
				}
				continue
			}
			hasBeenChild[parentStage.GetID()] = true
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)

		}
	}
	t.instructions = append(t.instructions, iID)
}

// checkSplit 判断是否需要分裂
// instructions的最后一个是否为previousId
func (t *SITree) checkSplit(stage *Stage, iID int) bool {
	return iID != stage.GetLastInstruction()
	//for id, _ := range iIDs {
	//	if id != stage.GetLastInstruction() {
	//		return true
	//	}
	//}
	//return false
}

// Combine 每次插入一个Instruction都先生成一个Stage，如果该Instruction可以因为窄依赖被加入到另一个Stage中，那么调用该函数
// element指要被合并的Stage,target指向哪个Stage合并，即最后的结果(id不变)
func (t *SITree) Combine(element *Stage, target *Stage) {
	target.BatchAddInstruction(element.GetInstructions(), false) // 这里是插入尾部
	t.batchUpdateIndex(element.GetInstructions(), target)
}

// Split 处理stage的分裂逻辑, cut代表从何处开始分裂
func (t *SITree) Split(stage *Stage, cut int) *Stage {
	beacon := -1 // cut的具体index
	for i := len(stage.GetInstructions()) - 1; i >= 0; i-- {
		if stage.GetInstructions()[i] == cut {
			beacon = i + 1
			break
		}
		//_, exist := cut[stage.GetInstructions()[i]]
		//if exist {
		//	beacon = i + 1
		//	break
		//}
	}
	//for i, id := range stage.GetInstructions() {
	//	_, exist := cut[id]
	//	if exist {
	//		beacon = i + 1 // 最后面的一个
	//		//break
	//	}
	//}
	if beacon == -1 {
		panic("Don't have such cut!!!")
	}
	fission := NewStage(-1, stage.GetInstructions()[beacon:]...)
	stage.CutInstruction(beacon) // 只保留前一部分
	fission.InheritChild(stage.GetSubStages())
	stage.RemoveAllChild()
	fission.AddParent(stage)
	stage.AddChild(fission)
	t.appendStage(fission) // 这里已经更新了后一部分instruction的映射
	return fission
}

// GetStageByIndex 获取对应位置的Stage
func (t *SITree) GetStageByIndex(index int) *Stage {
	return t.stages[index]
}

// GetStageByInstruction 获得Instruction对应的Stage
func (t *SITree) GetStageByInstruction(iId int) *Stage {
	_, exist := t.index[iId]
	if !exist {
		panic("Don't have such Instruction!!!")
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
func (t *SITree) Examine() {
	totalInstructionNumber := 0
	hasStore := make(map[int]bool)
	for index, stage := range t.stages {
		if index != stage.GetID() {
			fmt.Println("ID error")
		}
		for _, i := range stage.GetInstructions() {
			_, has := hasStore[i]
			if has {
				fmt.Println("error")
			} else {
				hasStore[i] = true
			}
		}
		totalInstructionNumber += len(stage.GetInstructions())
	}

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

func (t *SITree) HeuristicSplit() (*SITree, *SITree) {
	ret := make([]*Stage, 0)
	weightMap, fatherMap := computeSITStagesWeight(t)
	childList := make(map[int]bool)
	//for k, v := range weightMap {
	//	fmt.Println(k, v)
	//}
	for len(weightMap) != 0 {
		sampleNum := 10
		Pos := -1
		Score := -1.0
		if sampleNum > len(weightMap) {
			sampleNum = len(weightMap)
		}
		for k, v := range weightMap {
			if Score < v {
				Pos = k
				Score = v
			}
			sampleNum--
			if sampleNum == 0 {
				break
			}
		}
		targetStage := t.GetStageByInstruction(Pos)
		ret = append(ret, targetStage)
		childList[targetStage.GetLastInstruction()] = true
		wG := sync.WaitGroup{}
		for _, stage := range targetStage.GetSubStages() {
			childIdx := stage.GetLastInstruction()
			childList[childIdx] = true
			wG.Add(1)
			go func(childIdx int) {
				weightMapFixChild(t, weightMap, fatherMap, childList, make(map[int]bool), childIdx)
				wG.Done()
			}(childIdx)
		}
		wG.Wait()
		weightMapFixFather(weightMap, fatherMap, make(map[int]bool), Pos)
	}
	return generateNewSIT(t, fatherMap, ret, childList)
}
