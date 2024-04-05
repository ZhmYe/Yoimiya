package split

import (
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend"
	"S-gnark/graph"
	"fmt"
	"time"
)

/***

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// SplitAndProve 将传入的电路(constraintSystem)切分为多份，返回所有切出的子电路的proof
func SplitAndProve(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	extras := make([]constraint.ExtraValue, 0)
	startTime := time.Now()
	fmt.Println("=================Start Split=================")
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.Sit.AssignLayer()
		structureRoundLog(_r1cs, 0)
		top, bottom := _r1cs.Sit.CheckAndGetSubCircuitStageIDs()
		_r1cs.UpdateForwardOutput() // 这里从原电路中获得middle对应的wireIDs
		forwardOutput := _r1cs.GetForwardOutputs()
		//if err != nil {
		//	panic(err)
		//}
		var err error
		record := NewDataRecord(_r1cs)
		fmt.Print("	Top Circuit ")
		topCs, err := buildConstraintSystemFromIds(_r1cs.Sit.GetInstructionIdsFromStageIDs(top), record, assignment, forwardOutput, extras, true)
		if err != nil {
			panic(err)
		}

		proof := GetSplitProof(topCs, assignment, &extras, false)
		proofs = append(proofs, proof)
		//fmt.Println("bottom=", len(bottom))
		fmt.Print("	Bottom Circuit ")
		bottomCs, err := buildConstraintSystemFromIds(_r1cs.Sit.GetInstructionIdsFromStageIDs(bottom), record, assignment, forwardOutput, extras, false)
		if err != nil {
			panic(err)
		}
		proofs = append(proofs, GetSplitProof(bottomCs, assignment, &extras, false))
	default:
		panic("Only Support bn254 r1cs now...")
	}

	fmt.Println()
	fmt.Println("Total Time: ", time.Since(startTime))
	fmt.Println("=================Finish Split=================")
	return proofs, nil
}

// 将传入的cs转化为新的多个电路内部对应的sit，同时返回所有的instruction
func trySplit(cs *cs_bn254.R1CS) ([]*graph.SITree, error) {
	result := make([]*graph.SITree, 0)
	// todo 这里等待实现
	splitEngine := graph.NewSplitEngine(cs.Sit)
	//splitEngine.Split(1)
	top, bottom := splitEngine.Split(2)
	result = append(result, top)
	if bottom != nil {
		result = append(result, bottom)
	}
	return result, nil
}
func buildConstraintSystemFromSit(sit *graph.SITree, record *DataRecord) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的结果sit，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	cs := cs_bn254.NewR1CS(opt.Capacity)
	for i, stage := range sit.GetStages() {
		for j, iID := range stage.GetInstructions() {
			pi := record.GetPackedInstruction(iID)
			bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
			// todo 这里有很多Nb，如NbConstraints，暂时不确定前面是否需要加入
			// 这里需要重新得到WireID,不能沿用原来的WireID,因为后面的values是用WireID作为数组索引的
			//originInstruction := UnpackInstruction(pi, record) // 这里是原来的instruction，我们要尝试重新添加
			cs.AddInstruction(bID, unpack(pi, record))
			// 由于instruction变化，所以在这里需要重新映射stage内部的iID
			sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
		}
	}
	return cs, nil
}
func buildConstraintSystemFromIds(iIDs []int, record *DataRecord, assignment frontend.Circuit, forwardOutput []constraint.ExtraValue, extra []constraint.ExtraValue, isTop bool) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	if isTop {
		SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	}
	err := frontend.SetNbLeaf(assignment, cs, extra)
	if err != nil {
		return nil, err
	}
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	for _, iID := range iIDs {
		pi := record.GetPackedInstruction(iID)
		bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
		cs.AddInstructionInSpilt(bID, unpack(pi, record))
		//// 由于instruction变化，所以在这里需要重新映射stage内部的iID
		//sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
	}
	cs.CoeffTable = record.GetCoeffTable()
	if isTop {
		cs.CommitmentInfo = record.GetCommitmentInfo()
	}
	fmt.Println("Compile Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
	//fmt.Println(cs.Sit.GetStageNumber())
	return cs, nil
}
