package PackedLevel

import (
	"Yoimiya/graph"
)

// PackedLevel 用于记录某一instruction自身所在的LEVEL，以及它的子节点最大深度
type PackedLevel struct {
	Levels [][]int // 同core.go中的LEVELs
	//deepest []int   // 这里用于某一个cs，cs里的instruction id是连续的
	deepest []Sequence
	// todo 这里可以把deepest和layer结合一下
	layer []graph.Layer
	index map[int]int // 对应instruction -> level，如果直接遍历太慢了
	mark  []Sequence
}

const TOP = graph.TOP
const MIDDLE = graph.MIDDLE
const BOTTOM = graph.BOTTOM
const UNSET = graph.UNSET

type Sequence struct {
	depth int
	bias  int
}

func isGreater(s1 Sequence, s2 Sequence) bool {
	if s1.depth > s2.depth {
		return true
	} else if s1.depth == s2.depth {
		if s1.bias > s2.bias {
			return true
		}
	}
	return false
}
func NewPackedLevel() *PackedLevel {
	l := new(PackedLevel)
	l.Levels = make([][]int, 0)
	//l.deepest = make([]int, 0)
	l.deepest = make([]Sequence, 0)
	l.layer = make([]graph.Layer, 0)
	l.index = make(map[int]int)
	l.mark = make([]Sequence, 0)
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
	l.deepest = append(l.deepest, Sequence{
		depth: maxLevel,
		bias:  len(l.Levels[maxLevel]) - 1,
	})
	//l.deepest = append(l.deepest, maxLevel) // 记录instruction当前能抵达的最大深度，即它自己当前的深度
	l.layer = append(l.layer, graph.UNSET)
	// 更新父节点的deepest
	sequence := l.deepest[iID]
	for _, id := range previousIDs {
		current := l.deepest[id]
		if isGreater(sequence, current) {
			l.deepest[id] = sequence
		}
		//if current < maxLevel {
		//	l.deepest[id] = maxLevel
		//}
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

//func (l *PackedLevel) GetSubCircuitInstructionIDs() ([]int, []int) {
//	top, bottom := make([]int, 0), make([]int, 0)
//	for iID, layer := range l.layer {
//		if layer == BOTTOM {
//			bottom = append(bottom, iID)
//		} else {
//			top = append(top, iID)
//		}
//	}
//	return top, bottom
//}

func (l *PackedLevel) GetSubCircuitInstructionIDs() [][]int {
	subiIDs := make([][]int, 0)
	idx := 0
	iIDs := make([]int, 0)
	for i, level := range l.Levels {
		for j, iID := range level {
			sequence := Sequence{
				depth: i,
				bias:  j,
			}
			if isGreater(sequence, l.mark[idx]) {
				subiIDs = append(subiIDs, iIDs)
				idx++
				iIDs = make([]int, 0)
			}
			iIDs = append(iIDs, iID)
		}
	}
	if len(iIDs) != 0 {
		subiIDs = append(subiIDs, iIDs)
	}
	return subiIDs
	//top, bottom := make([]int, 0), make([]int, 0)
	//for iID, layer := range l.layer {
	//	if layer == BOTTOM {
	//		bottom = append(bottom, iID)
	//	} else {
	//		top = append(top, iID)
	//	}
	//}
	//return top, bottom
}
func (l *PackedLevel) AssignLayer(cut int) {
	//totalStageNumber := t.GetStageNumber()
	totalInstructionNumber := len(l.layer)
	//total := 0
	//splitDepth := -1
	// 遍历所有level
	//for i := 0; i < len(l.Levels); i++ {
	//	level := l.Levels[i]
	//	if 2*(total+len(level)) > totalInstructionNumber {
	//		splitDepth = i - 1
	//		break
	//	}
	//	total += len(level)
	//}
	// todo 这里的逻辑 优化到cut=n
	count := 0
	threshold := totalInstructionNumber / cut // 每次split的阈值
	//mark := make([]Sequence, 0)
	flag := false
	for i, level := range l.Levels {
		if flag {
			break
		}
		for j, _ := range level {
			if flag {
				break
			}
			count++ // 累加计数
			// 如果达到阈值，将累计的这些作为一个新的电路
			if count == threshold {
				count = 0 // 重新开始计数
				l.mark = append(l.mark, Sequence{
					depth: i,
					bias:  j,
				})
				// 最后一份直接计入最后的depth和bias即可
				if len(l.mark)+1 == cut {
					l.mark = append(l.mark, Sequence{
						depth: len(l.Levels),
						bias:  len(l.Levels[len(l.Levels)-1]),
					})
					flag = true
				}
			}
		}
	}
	//if count != 0 {
	//	l.mark = append(l.mark, Sequence{
	//		depth: len(l.Levels) - 1,
	//		bias:  len(l.Levels[len(l.Levels)-1]),
	//	})
	//}
	// 这里mark就记录了每一个子电路到哪一个instruction停下
	// 再次遍历所有的instruction
	checkMiddle := func(level int, bias int, depth Sequence, mark []Sequence) bool {
		sequence := -1
		for c, item := range mark {
			l, b := item.depth, item.bias
			if level < l {
				// 如果当前instruction所在的level比l小，那么就属于第c个电路
				sequence = c
				break
			}
			if level == l {
				// 在同一层但bias较小
				if bias <= b {
					sequence = c
					break
				}
			}
		}
		if sequence == -1 {
			sequence = len(mark)
		}
		// 如果达到的最大深度比自己所在的电路大，则为middle
		if isGreater(depth, mark[sequence]) {
			return true
		}
		return false
	}
	for i, level := range l.Levels {
		for j, id := range level {
			depth := l.deepest[id]
			if checkMiddle(i, j, depth, l.mark) {
				l.layer[id] = MIDDLE
			} else {
				l.layer[id] = TOP
			}
		}
	}

	//for i, level := range l.Levels {
	//	for _, id := range level {
	//		depth := l.deepest[id]
	//		// 如果本身所处层数大于splitDepth，则为bottom
	//		if i > splitDepth {
	//			l.layer[id] = BOTTOM
	//		} else {
	//			// 如果当前就在splitDepth或者子节点达到的最大位置在splitDepth之后，则为middle
	//			if depth > splitDepth || i == splitDepth {
	//				l.layer[id] = MIDDLE
	//			} else {
	//				l.layer[id] = TOP
	//			}
	//		}
	//	}
	//}

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
func (l *PackedLevel) IsMiddle(iID int) bool {
	return l.layer[iID] == MIDDLE
}
