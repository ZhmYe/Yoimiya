package constraint

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
	here is how to get levels.
	todo modify
	todo English comments
***/

// NewUpdateInstructionTree modify by ZhmYe
func (b *BlueprintGenericR1C) NewUpdateInstructionTree(inst Instruction, tree InstructionTree, iID int, cs *System) Level {
	/***
		Hints: ZhmYe
		callData:
			len: 4 + 2 * len(L) + 2 * len(R) + 2 * len(O)
			4: len, len(L), len(R), len(O)
			2* len(L)/len(R)/len(O): 2 * (CoeffID(), WireID())

		Wires and levels detailed in constraints/instruction_tree.go
	***/
	// a R1C doesn't know which wires are input and which are outputs
	//fmt.Println(iID)
	lenL := int(inst.Calldata[1])
	lenR := int(inst.Calldata[2])
	lenO := int(inst.Calldata[3])
	outputWires := make([]uint32, 0)
	maxLevel := LevelUnset
	cs.initDegree(iID)
	walkWires := func(n, idx int) {
		for k := 0; k < n; k++ {
			// 遍历每一个L、R、O中的WireID
			wireID := inst.Calldata[idx+1]
			idx += 2 // advance the offset (coeffID + wireID)
			// input or const
			if !tree.HasWire(wireID) {
				continue
			}
			// outputWires中存储所有level为LevelUnset的wireID
			if level := tree.GetWireLevel(wireID); level == LevelUnset {
				outputWires = append(outputWires, wireID)
			} else if level > maxLevel {
				// 如果当前wireID所在level不是LevelUnset(已经被记录了)，那么更新最大level, 便于知道outputWires应该插入到哪里
				// 那些还没有加入到level中的wires需要等这些已经加入的output wires计算完成后才能计算，因此level需要更大
				maxLevel = level
				// add by ZhmYe
				// 当前wireID已经在之前的Instruction中被记录，那么建立顺序关系
				// 前序Instruction
				for _, previousInstructionID := range cs.Wires2Instruction[wireID] {
					cs.InstructionForwardDAG.Update(previousInstructionID, iID)
					cs.InstructionBackwardDAG.Update(iID, previousInstructionID)
					cs.UpdateDegree(false, previousInstructionID) // 更新degree,这里用于更新Backward的degree
				}
			} else {
				// add by ZhmYe
				// 即使level没有超过最大level，只要有level就要遍历
				for _, previousInstructionID := range cs.Wires2Instruction[wireID] {
					cs.InstructionForwardDAG.Update(previousInstructionID, iID)
					cs.InstructionBackwardDAG.Update(iID, previousInstructionID)
					cs.UpdateDegree(false, previousInstructionID) // 更新degree,这里用于更新Backward的degree
				}
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
		// add by ZhmYe
		// 获得wire和Instruction之间的关系
		cs.AppendWire2Instruction(wireID, iID)
		tree.InsertWire(wireID, maxLevel)
	}

	return maxLevel
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
