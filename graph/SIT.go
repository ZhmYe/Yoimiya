package graph

import (
	"S-gnark/Config"
	"fmt"
	"strconv"
)

type Layer int

// 用于标记每个stage处在切分后电路的位置，目前只支持二分电路
// TOP表示在上层电路作为中间变量，MIDDLE表示作为上层电路的输出、下层电路的输入，BOTTOM表示在下层电路中作为中间变量
const (
	TOP Layer = iota
	MIDDLE
	BOTTOM
	UNSET
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

// SetLayer 设置某个stage的layer
func (t *SITree) SetLayer(id int, layer Layer) {
	t.layers[id] = layer
}

// GetLayer 获取某个stage的layer
func (t *SITree) GetLayer(id int) Layer {
	if id >= len(t.layers) {
		panic("This stage Layer hasn't been set...")
	}
	return t.layers[id]
}

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
func (t *SITree) AssignLayer() {
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
		t.appendStage(stage)
		//if hasBottom {
		//	t.SetLayer(stage.GetID(), BOTTOM)
		//} else {
		//	t.SetLayer(stage.GetID(), MIDDLE)
		//}
	}
	t.instructions = append(t.instructions, iID)
}

// InsertWithLayer 加上Layer的逻辑
func (t *SITree) InsertWithLayer(iID int, previousIds []int) {
	stage := NewStage(-1, iID) // id统一都默认初始化为-1，在append时处理
	// 如果没有父节点
	if len(previousIds) == 0 {
		t.appendStage(stage) // 直接append
		t.appendRoot(stage)  // 没有父节点则一定是root Stage
		// RootStage的Layer标记为Top
		t.SetLayer(stage.GetID(), TOP)

	} else if len(previousIds) == 1 {
		// 如果只有一个父节点
		// 暂时认为当前instruction和父instruction之间是窄依赖关系，合并stage
		previousId := previousIds[0]
		parentStage := t.GetStageByInstruction(previousId) // 得到父stage
		if t.checkSplit(parentStage, previousId) {
			// 需要分裂，那么最终为宽依赖
			fission := t.Split(parentStage, previousId)
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
			// 判断父节点的layer
			switch t.GetLayer(parentStage.GetID()) {
			case TOP:
				// 父节点依旧保持为TOP
				t.SetLayer(fission.GetID(), TOP) // 将分身节点的layer也设置为TOP,这样无需处理下面的节点变更
				// 将当前stage设置为Middle

				// todo
				t.SetLayer(stage.GetID(), MIDDLE)
				//t.SetLayer(stage.GetID(), TOP) // 如果是单一父节点，不赋值MIDDLE
			case MIDDLE:
				// 如果父节点是Middle，那么将分裂体设置为Bottom，当前stage也设置为Bottom
				// 这里不会影响父节点是否可以作为Middle
				t.SetLayer(fission.GetID(), BOTTOM)
				t.SetLayer(stage.GetID(), BOTTOM)
			case BOTTOM:
				// 如果父节点是Bottom，将分裂体和当前stage设置为Bottom
				t.SetLayer(fission.GetID(), BOTTOM)
				t.SetLayer(stage.GetID(), BOTTOM)
			default:
				panic("Unset Layer Type...")
			}
		} else if len(parentStage.GetSubStages()) != 0 {
			// 如果不需要分裂，但父stage有多个子stage，那么也是宽依赖
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
			// 如果不需要分裂，父节点有stage
			switch t.GetLayer(parentStage.GetID()) {
			case TOP:
				// 如果父节点是TOP,那么可以将当前节点置为Middle
				// todo
				t.SetLayer(stage.GetID(), MIDDLE) // 只有一个父节点
				//t.SetLayer(stage.GetID(), TOP)
			case MIDDLE:
				// 如果父节点是Middle，此时不会影响父节点作为Middle，因此只需要把当前节点置为BOTTOM
				t.SetLayer(stage.GetID(), BOTTOM)
			case BOTTOM:
				// 如果父节点是BOTTOM，那么显然只需要把当前节点置为BOTTOM
				t.SetLayer(stage.GetID(), BOTTOM)
			default:
				panic("Unset Layer Type...")
			}
		} else {
			// 无需分裂，并且父stage当前没有子stage，那么暂时认为是窄依赖
			t.Combine(stage, parentStage)
			// 如果不需要分裂并且父节点没有stage，那么没有新的stage被添加，因此不需要更改layer
		}
	} else {
		// 如果不止有一个父节点，一定是宽依赖
		hasBeenChild := make(map[int]bool) // 可能会出现多个previousId在同一个stage里面，那么无需后续重新添加child
		// Layer逻辑：
		// 如果不止一个父节点，那么需要遍历所有父节点，此时父节点Layer内容可能变换
		// 尝试把当前节点置为Middle，但需要对所有父节点的Layer情况进行判断
		// 既然当前节点有多个父节点，那么要求其所有一阶父节点均为Middle或均不为Middle
		hasBottom := false
		hasTop := false
		hasMiddle := false
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
				// 判断父节点类型，对分裂体进行标注，暂时还不对Stage本身进行标注
				switch t.GetLayer(parentStage.GetID()) {
				case TOP:
					// 如果父节点为TOP，那么标注分裂体为TOP
					t.SetLayer(fission.GetID(), TOP)
					hasTop = true
				case BOTTOM:
					// 如果父节点为BOTTOM，那么标注分裂体为BOTTOM
					t.SetLayer(fission.GetID(), BOTTOM)
					hasBottom = true // 出现了Bottom的父节点
				case MIDDLE:
					// 如果父节点为Middle
					// 这里我们令父节点不再为Middle,这样避免处理判断其他父节点是否可以变成Middle的逻辑
					// 将父节点标注为TOP
					t.SetLayer(parentStage.GetID(), TOP)
					// 此时分裂体继承了父节点的所有子节点，并且分裂体有且仅有一个父节点，因此我们可以将分裂体置为Middle
					t.SetLayer(fission.GetID(), MIDDLE)
					//hasMiddle = true
					hasTop = true
				default:
					panic("Unset Layer Type...")
				}
			} else {
				// 如果不需要分裂
				switch t.GetLayer(parentStage.GetID()) {
				case TOP:
					// 如果父节点为TOP
					// 此时不需要做任何处理，需要等待后续处理当前stage的layer
					hasTop = true
				case BOTTOM:
					// 如果父节点为BOTTOM
					hasBottom = true
				case MIDDLE:
					hasMiddle = true
				default:
					panic("Unset Layer Type...")
				}
			}
			_, flag := hasBeenChild[parentStage.GetID()]
			if flag {
				if hasSplit {
					// 分裂的两个stage都是当前stage的父stage
					fission.AddChild(stage)
					stage.AddParent(fission)
					// 处理fission的layer是否需要更改
					switch t.GetLayer(fission.GetID()) {
					case TOP:
						hasTop = true
						// 如果fission是TOP，那么不影响，无需处理
					case BOTTOM:
						// 如果fission是BOTTOM，它也会作为当前节点的父节点
						hasBottom = true
					case MIDDLE:
						// 如果fission是MIDDLE，MIDDLE之间不能相互连接
						// 之前考虑过如果有多个一阶父节点父节点，那么这些一阶父节点父节点要么都为MIDDLE
						// 我们已经将其他非分裂体的父节点全部保证不为MIDDLE
						// 剩下的都是分裂体，且能进入这里的那些分裂体，其父节点也是当前stage的父节点
						// 此时我们可以通过这里的逻辑保证所有这些分裂体都是MIDDLE，因此可以将当前stage置为BOTTOM
						hasMiddle = true
					default:
						panic("Unset Layer Type...")
					}
				}
				continue
			}
			hasBeenChild[parentStage.GetID()] = true
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)

		}
		// 把stage append
		t.appendStage(stage)
		// 遍历完所有的父节点后，根据父节点的类型判断
		if hasTop && !hasMiddle && !hasBottom {
			// 如果只有top的父节点, 那么将当前节点置为middle
			// todo
			//if t.CheckParentSameDepth(stage.GetID()) {
			//	t.SetLayer(stage.GetID(), MIDDLE)
			//} else {
			//	t.SetLayer(stage.GetID(), TOP)
			//}
			t.SetLayer(stage.GetID(), MIDDLE)
		} else if !hasTop && hasMiddle && !hasBottom {
			// 如果只有middle的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else if !hasTop && !hasMiddle && hasBottom {
			// 如果只有bottom的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else if !hasTop {
			// 如果没有top的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else {
			// 如果有top的父节点，那么当前节点置为Bottom，并且递归修改为top的父节点
			t.SetLayer(stage.GetID(), BOTTOM)
			for _, pid := range stage.GetParentIDs() {
				pLayer := t.GetLayer(pid)
				if pLayer == TOP {
					t.switchTop(pid, MIDDLE)
				}
			}
		}
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

type ExamineResult int

const (
	PASS ExamineResult = iota
	HAS_LINK
	LAYER_UNSET
	SPLIT_ERROR
)

// Examine 之前已经保证SIT划分是正确的，这里就检验Layer
// todo MIDDLE之间是否一定不能相连？放宽条件到可以，因为middle是在上半电路中被计算完传下去的
func (t *SITree) Examine() ExamineResult {
	// 首先判断每个stage是否都被赋上了Layer
	if len(t.stages) != len(t.layers) {
		return LAYER_UNSET
	}
	// 记录所有的middle
	middleIds := make(map[int]bool, 0)
	for id, layer := range t.layers {
		if layer == MIDDLE {
			middleIds[id] = true
		}
	}
	//for id, _ := range middleIds {
	//	stage := t.GetStageByIndex(id)
	//	for _, pid := range stage.GetParentIDs() {
	//		if middleIds[pid] {
	//			return HAS_LINK
	//		}
	//	}
	//}
	for id, layer := range t.layers {
		if layer == TOP || layer == MIDDLE {
			// top和middle的父节点必须是top
			stage := t.GetStageByIndex(id)
			for _, pid := range stage.GetParentIDs() {
				pLayer := t.GetLayer(pid)
				if pLayer != TOP {
					fmt.Println(layer, pLayer)
					return SPLIT_ERROR
				}
			}
		}
		if layer == BOTTOM {
			// bottom的父节点必须是bottom和middle
			stage := t.GetStageByIndex(id)
			for _, pid := range stage.GetParentIDs() {
				pLayer := t.GetLayer(pid)
				if pLayer == TOP {
					fmt.Println(layer, pLayer, len(stage.GetParentIDs()))
					return SPLIT_ERROR
				}
			}
		}
	}
	return PASS

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
func (t *SITree) GetLayersInfo() (info [4]int) {
	for i, _ := range info {
		info[i] = 0
	}
	for _, layer := range t.GetLayers() {
		switch layer {
		case TOP:
			info[0]++
		case BOTTOM:
			info[2]++
		case MIDDLE:
			info[1]++
		case UNSET:
			info[3]++
		}
	}
	return info
}
func (t *SITree) GetLayers() []Layer {
	return t.layers
}
func (t *SITree) GetTopStage() []int {
	sIDs := make([]int, 0)
	for id, layer := range t.layers {
		if layer == TOP {
			sIDs = append(sIDs, id)
		}
	}
	return sIDs
}
func (t *SITree) GetMiddleStage() []int {
	sIDs := make([]int, 0)
	for id, layer := range t.layers {
		if layer == MIDDLE {
			sIDs = append(sIDs, id)
		}
	}
	return sIDs
}
func (t *SITree) GetBottomStage() []int {
	sIDs := make([]int, 0)
	for id, layer := range t.layers {
		if layer == BOTTOM {
			sIDs = append(sIDs, id)
		}
	}
	return sIDs
}
func (t *SITree) checkUnset() bool {
	for _, layer := range t.layers {
		if layer == UNSET {
			return true
		}
	}
	return false
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

//	func (t *SITree) GetAncestorStages(stageId int, result *map[int]bool) {
//		_, exist := (*result)[stageId]
//		if exist {
//			return
//		}
//		stage := t.GetStageByIndex(stageId)
//		for _, id := range stage.GetParentIDs() {
//			t.GetAncestorStages(id, result)
//			(*result)[id] = true
//		}
//	}
//
//	func (t *SITree) GetDescendantStages(stageId int, result *map[int]bool) {
//		_, exist := (*result)[stageId]
//		if exist {
//			return
//		}
//		stage := t.GetStageByIndex(stageId)
//		for _, id := range stage.GetChildIDs() {
//			t.GetDescendantStages(id, result)
//			(*result)[id] = true
//		}
//	}
func (t *SITree) CountLeafNode() (total int) {
	for _, stage := range t.GetStages() {
		if len(stage.GetSubStages()) == 0 {
			total++
		}
	}
	return
}
