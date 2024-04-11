package Sit

import (
	"Yoimiya/graph"
	"math/rand"
)

type Layer = graph.Layer

const UNSET = graph.UNSET
const TOP = graph.TOP
const MIDDLE = graph.MIDDLE
const BOTTOM = graph.BOTTOM

// SetLayer 设置某个stage的layer
func (t *SITree) SetLayer(id int, layer graph.Layer) {
	t.layers[id] = layer
}
func (t *SITree) checkUnset() bool {
	for _, layer := range t.layers {
		if layer == graph.UNSET {
			return true
		}
	}
	return false
}

// GetLayer 获取某个stage的layer
func (t *SITree) GetLayer(id int) Layer {
	if id >= len(t.layers) {
		panic("This stage Layer hasn't been set...")
	}
	return t.layers[id]
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
