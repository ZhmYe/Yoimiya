package constraint

import (
	"fmt"
)

// BlueprintGenericR1C implements Blueprint and BlueprintR1C.
// Encodes
//
//	L * R == 0
type BlueprintGenericR1C struct{}

func (b *BlueprintGenericR1C) CalldataSize() int {
	// size of linear expressions are unknown.
	return -1
}
func (b *BlueprintGenericR1C) NbConstraints() int {
	return 1
}
func (b *BlueprintGenericR1C) NbOutputs(inst Instruction) int {
	return 0
}

func (b *BlueprintGenericR1C) CompressR1C(c *R1C, to *[]uint32) {
	// we store total nb inputs, len L, len R, len O, and then the "flatten" linear expressions
	nbInputs := 4 + 2*(len(c.L)+len(c.R)+len(c.O))
	(*to) = append((*to), uint32(nbInputs))
	(*to) = append((*to), uint32(len(c.L)), uint32(len(c.R)), uint32(len(c.O)))
	for _, t := range c.L {
		(*to) = append((*to), uint32(t.CoeffID()), uint32(t.WireID()))
	}
	for _, t := range c.R {
		(*to) = append((*to), uint32(t.CoeffID()), uint32(t.WireID()))
	}
	for _, t := range c.O {
		(*to) = append((*to), uint32(t.CoeffID()), uint32(t.WireID()))
	}
}

func (b *BlueprintGenericR1C) DecompressR1C(c *R1C, inst Instruction) {
	copySlice := func(slice *LinearExpression, expectedLen, idx int) {
		if cap(*slice) >= expectedLen {
			(*slice) = (*slice)[:expectedLen]
		} else {
			(*slice) = make(LinearExpression, expectedLen, expectedLen*2)
		}
		for k := 0; k < expectedLen; k++ {
			(*slice)[k].CID = inst.Calldata[idx]
			idx++
			(*slice)[k].VID = inst.Calldata[idx]
			idx++
		}
	}

	lenL := int(inst.Calldata[1])
	lenR := int(inst.Calldata[2])
	lenO := int(inst.Calldata[3])

	const offset = 4
	copySlice(&c.L, lenL, offset)
	copySlice(&c.R, lenR, offset+2*lenL)
	copySlice(&c.O, lenO, offset+2*(lenL+lenR))
}

/***
	Hints: ZhmYe
	here is how to instruct the SIT
	all type blueprint has NewUpdateInstructionTree
***/

// NewUpdateInstructionTree modify by ZhmYe
func (b *BlueprintGenericR1C) NewUpdateInstructionTree(inst Instruction, tree InstructionTree, iID int, cs *System, split bool, needAppend bool) {
	lenL := int(inst.Calldata[1])
	lenR := int(inst.Calldata[2])
	lenO := int(inst.Calldata[3])
	outputWires := make([]uint32, 0)
	//maxLevel := -1
	//outputWires := make(map[uint32]bool)
	previousIds := make([]int, 0)
	walkWires := func(n, idx int) {
		for k := 0; k < n; k++ {
			// 遍历每一个L、R、O中的WireID
			wireID := inst.Calldata[idx+1]
			idx += 2 // advance the offset (coeffID + wireID)
			// input or const
			// 这里修改了HasWire，将lbWireLevel去掉了，具体还要看core.go中AddInternalVariable
			//if !tree.HasWire(wireID) {
			//	continue
			//}
			if tree.IsInputOrConstant(wireID, split) {
				// todo 这里加上-1 * inputWire到previousIds
				// 但这样一来其他的算法会需要判断previousIds中是否有负数
				previousIds = append(previousIds, -int(wireID))

				//inputWires[wireID] = true
				//cs.UpdateUsedExtra(int(wireID))
				continue
			}

			// outputWires中存储所有level为LevelUnset的wireID
			// 原本下面通过判断level是否存在来判断，现在可以通过判断Wire2Instruction判断
			_, notOutput := cs.Wires2Instruction[wireID]
			if !notOutput {
				//if level := tree.GetWireLevel(wireID); level == LevelUnset {
				outputWires = append(outputWires, wireID)
				//outputWires[wireID] = true
			} else {
				// add by ZhmYe
				// 即使level没有超过最大level，只要有level就要遍历
				// 当前wireID已经在之前的Instruction中被记录，那么建立顺序关系
				// 前序Instruction

				// 这里改成获取instruction的level
				//
				previousInstructionID, exist := cs.Wires2Instruction[wireID]
				if !exist {
					fmt.Println(wireID)
					panic("error in hint")
				}
				previousIds = append(previousIds, previousInstructionID)
			}
		}
	}

	const offset = 4
	walkWires(lenL, offset)
	walkWires(lenR, offset+2*lenL)
	walkWires(lenO, offset+2*(lenL+lenR))

	// insert the new wires.
	//maxLevel++
	for _, wireID := range outputWires {
		// add by ZhmYe
		// 获得wire和Instruction之间的关系
		cs.AppendWire2Instruction(wireID, iID)
		if needAppend {
			cs.SetBias(wireID, cs.AddInternalVariable())
		}
		//if Config.Config.Split == Config.SPLIT_LEVELS {
		//	tree.InsertWire(wireID, maxLevel)
		//}
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

	//for wireID, _ := range inputWires {
	//	cs.UpdateUsedExtra(int(wireID))
	//}
	//return maxLevel
}

/***
	Hints: ZhmYe
	This is origin UpdateInstructionTree function in gnark
***/

func (b *BlueprintGenericR1C) UpdateInstructionTree(inst Instruction, tree InstructionTree) Level {
	// a R1C doesn't know which wires are input and which are outputs
	lenL := int(inst.Calldata[1])
	lenR := int(inst.Calldata[2])
	lenO := int(inst.Calldata[3])

	outputWires := make([]uint32, 0)
	maxLevel := LevelUnset
	walkWires := func(n, idx int) {
		for k := 0; k < n; k++ {
			wireID := inst.Calldata[idx+1]
			idx += 2 // advance the offset (coeffID + wireID)
			// input or const
			if !tree.HasWire(wireID) {
				continue
			}
			if level := tree.GetWireLevel(wireID); level == LevelUnset {
				outputWires = append(outputWires, wireID)
			} else if level > maxLevel {
				maxLevel = level
			}
		}
	}

	const offset = 4
	walkWires(lenL, offset)
	walkWires(lenR, offset+2*lenL)
	walkWires(lenO, offset+2*(lenL+lenR))

	// insert the new wires.
	maxLevel++
	for _, wireID := range outputWires {
		tree.InsertWire(wireID, maxLevel)
	}

	return maxLevel
}
