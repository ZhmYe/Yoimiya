package Sit

import (
	"Yoimiya/Config"
)

// SITree Stage-Instruction Tree
type SITree struct {
	stages       []*Stage       // SITree节点
	instructions []int          // 只记录instruction id
	index        map[int]*Stage // 索引，iID -> Stage
	root         []*Stage       // 所有没有父stage的stage
	depth        []int          // 深度
	maxDepth     int            // 最大深度
	layers       []Layer        // 这里和stages的下标相对应
}

func NewSITree() *SITree {
	t := new(SITree)
	t.stages = make([]*Stage, 0)
	t.instructions = make([]int, 0)
	t.index = make(map[int]*Stage)
	t.root = make([]*Stage, 0)
	t.depth = make([]int, 0) // 这里记录每个stage的depth
	t.layers = make([]Layer, 0)
	return t
}

// appendStage 正式将stage添加到stage列表中
func (t *SITree) appendStage(stage *Stage) {
	stage.SetID(len(t.stages)) // 只有当真正append的时候才会添加id，确保id和index对应
	t.stages = append(t.stages, stage)
	t.batchUpdateIndex(stage.GetInstructions(), stage)
	t.layers = append(t.layers, UNSET)
	t.ComputeDepth(stage)
	//for _, iID := range stage.GetInstructions() {
	//	t.index[iID] = stage.GetID()
	//}
}

// appendRoot 某个stage没有父节点，是整个sit的根节点（可能不止一个）
func (t *SITree) appendRoot(stage *Stage) {
	t.root = append(t.root, stage)
}

// Insert 插入一个instruction时会伴有它的所有前置一阶instruction(previousId)
// 首先将当前instruction放在一个新的stage中
// 然后首先判断当前instruction是否有多个父节点
// 如果是，为宽依赖，将新建的stage作为child连接到所有父stage之后
//
//	case1: 父stage的最后一个instruction就是父instruction，那么直接append
//	case2: 父stage的最后一个instruction不是父instruction，那么在此之前，已有instruction被认为是窄依赖添加到了
//	父instruction之后,此时需要进行分裂(split)，将父instruction之后的所有instruction放在一个新的stage中，然后
//	和当前stage一起作为child连接到父stage中
//
// 如果不是，那么判断父节点是否有多个子节点，同上述case1,case2，区别在如果是窄依赖直接combine
// todo 这里的注释需要更新
func (t *SITree) Insert(iID int, previousIds []int) {
	if Config.Config.IsCluster() {
		t.InsertWithLayer(iID, previousIds)
	} else {
		t.InsertWithoutLayer(iID, previousIds)
	}
}
func (t *SITree) InsertWithoutLayer(iID int, previousIds []int) {
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
		hasBeenChild := make(map[int]bool) // 可能会出现多个previousId在同一个stage里面，那么无需后续重新添加child
		// 遍历所有父节点
		for _, previousId := range previousIds {
			// 首先针对所有父节点
			parentStage := t.GetStageByInstruction(previousId)
			var fission *Stage
			hasSplit := false
			// 判断是否需要split
			if t.checkSplit(parentStage, previousId) {
				// 需要分裂
				fission = t.Split(parentStage, previousId)
				hasSplit = true
			} else {
				// 如果不需要分裂
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
		// 把stage append
		// 这里一定要在最后添加，因为computeDepth需要所有父节点包括fission
		t.appendStage(stage)
		//if hasBottom {
		//	t.SetLayer(stage.GetID(), BOTTOM)
		//} else {
		//	t.SetLayer(stage.GetID(), MIDDLE)
		//}
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
	}
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
