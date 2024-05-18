package PackedLevel

import (
	"Yoimiya/graph"
)

// PackedLevel 用于记录某一instruction自身所在的LEVEL，以及它的子节点最大深度
type PackedLevel struct {
	Levels  [][]int // 同core.go中的LEVELs
	deepest []int   // 这里用于某一个cs，cs里的instruction id是连续的
	// todo 这里可以把deepest和layer结合一下
	layer []graph.Layer
	index map[int]int // 对应instruction -> level，如果直接遍历太慢了
}

const TOP = graph.TOP
const MIDDLE = graph.MIDDLE
const BOTTOM = graph.BOTTOM
const UNSET = graph.UNSET

func NewPackedLevel() *PackedLevel {
	l := new(PackedLevel)
	l.Levels = make([][]int, 0)
	l.deepest = make([]int, 0)
	l.layer = make([]graph.Layer, 0)
	l.index = make(map[int]int)
	return l
}
func (l *PackedLevel) Insert(iID int, previousIDs []int) {
	// 如果level已经比当前维护的levels多了，新建一层，并将iID更新
	//previousIdsMap := make(map[int]bool)
	maxLevel := -1
	//for _, id := range previousIDs {
	//	previousIdsMap[id] = true
	//}
	// todo 这里用遍历一次来换取instruction -> level的内存
	// todo 实践表明太慢了
	for _, id := range previousIDs {
		level, exist := l.index[id]
		if !exist {
			panic("no such instruction!!!")
		}
		if level > maxLevel {
			maxLevel = level
		}
	}
	//for i, level := range l.Levels {
	//	for _, id := range level {
	//		_, exist := previousIdsMap[id]
	//		if exist {
	//			if i > maxLevel {
	//				maxLevel = i
	//			}
	//		}
	//	}
	//}
	maxLevel++
	if maxLevel >= len(l.Levels) {
		l.Levels = append(l.Levels, []int{iID})
	} else {
		// 将iID添加到level
		l.Levels[maxLevel] = append(l.Levels[maxLevel], iID)
	}
	l.index[iID] = maxLevel
	// 这里可以保证iID就是下标
	l.deepest = append(l.deepest, maxLevel) // 记录instruction当前能抵达的最大深度，即它自己当前的深度
	l.layer = append(l.layer, graph.UNSET)
	// 更新父节点的deepest
	for _, id := range previousIDs {
		current := l.deepest[id]
		if current < maxLevel {
			l.deepest[id] = maxLevel
		}
	}
}

// GetStageNumber 这里简单的返回level的个数
func (l *PackedLevel) GetStageNumber() int {
	return len(l.Levels)
}
func (l *PackedLevel) GetLayersInfo() (info [4]int) {
	for i, _ := range info {
		info[i] = 0
	}
	for _, layer := range l.layer {
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

func (l *PackedLevel) GetSubCircuitInstructionIDs() ([]int, []int) {
	top, bottom := make([]int, 0), make([]int, 0)
	for iID, layer := range l.layer {
		if layer == BOTTOM {
			bottom = append(bottom, iID)
		} else {
			top = append(top, iID)
		}
	}
	return top, bottom
}

func (l *PackedLevel) AssignLayer() {
	//totalStageNumber := t.GetStageNumber()
	totalInstructionNumber := len(l.layer)
	total := 0
	splitDepth := -1
	// 遍历所有level
	for i := 0; i < len(l.Levels); i++ {
		level := l.Levels[i]
		if 2*(total+len(level)) > totalInstructionNumber {
			splitDepth = i - 1
			break
		}
		total += len(level)
	}
	// todo 这里的逻辑 优化到cut=n
	for i, level := range l.Levels {
		for _, id := range level {
			depth := l.deepest[id]
			// 如果本身所处层数大于splitDepth，则为bottom
			if i > splitDepth {
				l.layer[id] = BOTTOM
			} else {
				// 如果当前就在splitDepth或者子节点达到的最大位置在splitDepth之后，则为middle
				if depth > splitDepth || i == splitDepth {
					l.layer[id] = MIDDLE
				} else {
					l.layer[id] = TOP
				}
			}
		}
	}

}

func (l *PackedLevel) GetMiddleOutputs() map[int]bool {
	result := make(map[int]bool)
	for iID, layer := range l.layer {
		if layer == MIDDLE {
			result[iID] = true
		}
	}
	return result
}
func (l *PackedLevel) GetAllInstructions() []int {
	result := make([]int, 0)
	for iID, _ := range l.layer {
		result = append(result, iID)
	}
	return result
}
