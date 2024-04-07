package Sit

import "fmt"

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
