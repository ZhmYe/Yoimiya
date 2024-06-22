package constraint

import (
	"Yoimiya/constraint/solver"
	"Yoimiya/debug"
	"fmt"
	"math"
)

type BlueprintGenericHint struct{}

func (b *BlueprintGenericHint) DecompressHint(h *HintMapping, inst Instruction) {
	// ignore first call data == nbInputs
	h.HintID = solver.HintID(inst.Calldata[1])
	lenInputs := int(inst.Calldata[2])
	if cap(h.Inputs) >= lenInputs {
		h.Inputs = h.Inputs[:lenInputs]
	} else {
		h.Inputs = make([]LinearExpression, lenInputs)
	}

	j := 3
	for i := 0; i < lenInputs; i++ {
		n := int(inst.Calldata[j]) // len of linear expr
		j++
		if cap(h.Inputs[i]) >= n {
			h.Inputs[i] = h.Inputs[i][:0]
		} else {
			h.Inputs[i] = make(LinearExpression, 0, n)
		}
		for k := 0; k < n; k++ {
			h.Inputs[i] = append(h.Inputs[i], Term{CID: inst.Calldata[j], VID: inst.Calldata[j+1]})
			j += 2
		}
	}
	h.OutputRange.Start = inst.Calldata[j]
	h.OutputRange.End = inst.Calldata[j+1]
}

func (b *BlueprintGenericHint) CompressHint(h HintMapping, to *[]uint32) {
	nbInputs := 1 // storing nb inputs
	nbInputs++    // hintID
	nbInputs++    // len(h.Inputs)
	for i := 0; i < len(h.Inputs); i++ {
		nbInputs++ // len of h.Inputs[i]
		nbInputs += len(h.Inputs[i]) * 2
	}

	nbInputs += 2 // output range start / end

	(*to) = append((*to), uint32(nbInputs))
	(*to) = append((*to), uint32(h.HintID))
	(*to) = append((*to), uint32(len(h.Inputs)))

	for _, l := range h.Inputs {
		(*to) = append((*to), uint32(len(l)))
		for _, t := range l {
			(*to) = append((*to), uint32(t.CoeffID()), uint32(t.WireID()))
		}
	}

	(*to) = append((*to), h.OutputRange.Start)
	(*to) = append((*to), h.OutputRange.End)
}

func (b *BlueprintGenericHint) CalldataSize() int {
	return -1
}
func (b *BlueprintGenericHint) NbConstraints() int {
	return 0
}

func (b *BlueprintGenericHint) NbOutputs(inst Instruction) int {
	return 0
}

func (b *BlueprintGenericHint) UpdateInstructionTree(inst Instruction, tree InstructionTree) Level {
	// BlueprintGenericHint knows the input and output to the instruction
	maxLevel := LevelUnset

	// iterate over the inputs and find the max level
	lenInputs := int(inst.Calldata[2])
	j := 3
	for i := 0; i < lenInputs; i++ {
		n := int(inst.Calldata[j]) // len of linear expr
		j++

		for k := 0; k < n; k++ {
			wireID := inst.Calldata[j+1]
			j += 2
			if !tree.HasWire(wireID) {
				continue
			}
			if level := tree.GetWireLevel(wireID); level > maxLevel {
				maxLevel = level
			}
			if debug.Debug && tree.GetWireLevel(wireID) == LevelUnset {
				panic("wire we depend on is not in the instruction tree")
			}
		}
	}

	// iterate over the outputs and insert them at maxLevel + 1
	outputLevel := maxLevel + 1
	for k := inst.Calldata[j]; k < inst.Calldata[j+1]; k++ {
		tree.InsertWire(k, outputLevel)
	}
	return outputLevel
}

// NewUpdateInstructionTree add by ZhmYe
func (b *BlueprintGenericHint) NewUpdateInstructionTree(inst Instruction, tree InstructionTree, iID int, cs *System, split bool, needAppend bool) {
	// BlueprintGenericHint knows the input and output to the instruction
	//maxLevel := -1
	// iterate over the inputs and find the max level
	lenInputs := int(inst.Calldata[2])
	j := 3
	//cs.initDegree(iID)
	previousIds := make([]int, 0)
	//inputWires := make(map[uint32]bool)
	for i := 0; i < lenInputs; i++ {
		n := int(inst.Calldata[j]) // len of linear expr
		j++

		for k := 0; k < n; k++ {
			wireID := inst.Calldata[j+1]
			j += 2
			if tree.IsInputOrConstant(wireID, split) {
				// todo 这里加上-1 * inputWire到previousIds
				// 但这样一来其他的算法会需要判断previousIds中是否有负数
				if wireID != math.MaxUint32 {
					previousIds = append(previousIds, -int(wireID))
				}
				//cs.UpdateUsedExtra(int(wireID))
				//inputWires[wireID] = true
				continue
			}
			// add by ZhmYe
			// 前序Instruction
			previousInstructionID, exist := cs.Wires2Instruction[wireID]
			if !exist {
				fmt.Println(wireID)
				panic("error in hint")
			}
			//fmt.Println(wireID, len(cs.Wires2Instruction), previousInstructionID, iID)
			previousIds = append(previousIds, previousInstructionID)
		}
	}

	// iterate over the outputs and insert them at maxLevel + 1
	for k := inst.Calldata[j]; k < inst.Calldata[j+1]; k++ {
		cs.AppendWire2Instruction(k, iID)
		if needAppend {
			cs.SetBias(k, cs.AddInternalVariable())
		}
		//tree.InsertWire(k, outputLevel)
	}
	cs.SplitEngine.Insert(iID, previousIds)
	//switch Config.Config.Split {
	//// 如果是stage，则开始维护sit
	//case Config.SPLIT_STAGES:
	//	cs.Sit.Insert(iID, previousIds)
	//// 如果是level，则计算得到Levels
	//case Config.SPLIT_LEVELS:
	//	previousIdsMap := make(map[int]bool)
	//	for _, id := range previousIds {
	//		previousIdsMap[id] = true
	//	}
	//	// todo 这里用遍历一次来换取instruction -> level的内存
	//	for i, level := range cs.Levels {
	//		for _, id := range level {
	//			_, exist := previousIdsMap[id]
	//			if exist {
	//				if i > maxLevel {
	//					maxLevel = i
	//				}
	//			}
	//		}
	//	}
	//	maxLevel++
	//	// we can't skip levels, so appending is fine.
	//	if maxLevel >= len(cs.Levels) {
	//		cs.Levels = append(cs.Levels, []int{iID})
	//	} else {
	//		cs.Levels[maxLevel] = append(cs.Levels[maxLevel], iID)
	//	}
	//	cs.deepest = append(cs.deepest, maxLevel) // 记录instruction当前能抵达的最大深度，即它自己当前的深度
	//	// 更新父节点的deepest
	//	for _, id := range previousIds {
	//		cs.deepest[id] = maxLevel
	//	}
	//default:
	//	panic("error SPLIT_METHOD")
	//
	//}
}
