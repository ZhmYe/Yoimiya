package split

import (
	"Yoimiya/Record"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
	"time"
)

/***

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// SplitAndProve 这里把sit和level用interface统一写成了splitEngine，在内部区分
func SplitAndProve(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	extras := make([]constraint.ExtraValue, 0)
	startTime := time.Now()
	fmt.Println("=================Start Split=================")
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SplitEngine.AssignLayer()
		StructureRoundLog(_r1cs, 0)
		top, bottom := _r1cs.SplitEngine.GetSubCircuitInstructionIDs()
		//_r1cs.SplitEngine.GetSubCircuitInstructionIDs()
		_r1cs.UpdateForwardOutput() // 这里从原电路中获得middle对应的wireIDs
		forwardOutput := _r1cs.GetForwardOutputs()
		Record.GlobalRecord.SetSplitTime(time.Since(startTime)) // 设置SplitTime
		buildStartTime := time.Now()
		var err error
		record := NewDataRecord(_r1cs)
		fmt.Print("	Top Circuit ")
		topCs, err := buildConstraintSystemFromIds(top, record, assignment, forwardOutput, extras, true)
		if err != nil {
			panic(err)
		}
		Record.GlobalRecord.SetBuildTime(time.Since(buildStartTime))
		//testTime := time.Now()
		//for {
		//	if time.Since(testTime) >= time.Duration(20)*time.Second {
		//		break
		//	}
		//}
		proof := GetSplitProof(topCs, assignment, &extras, false)
		proofs = append(proofs, proof)
		fmt.Print("	Bottom Circuit ")
		buildStartTime = time.Now()
		bottomCs, err := buildConstraintSystemFromIds(bottom, record, assignment, forwardOutput, extras, false)
		if err != nil {
			panic(err)
		}
		Record.GlobalRecord.SetBuildTime(time.Since(buildStartTime))
		//GetSplitProof(bottomCs, assignment, &extras, false)
		proofs = append(proofs, GetSplitProof(bottomCs, assignment, &extras, false))
	default:
		panic("Only Support bn254 r1cs now...")
	}

	fmt.Println()
	fmt.Println("Total Time: ", time.Since(startTime))
	fmt.Println("=================Finish Split=================")
	return proofs, nil
}
func buildConstraintSystemFromIds(iIDs []int, record *DataRecord, assignment frontend.Circuit,
	forwardOutput []constraint.ExtraValue, extra []constraint.ExtraValue, isTop bool) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	if isTop {
		SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
		cs.CommitmentInfo = record.GetCommitmentInfo()
	}
	err := frontend.SetNbLeaf(assignment, cs, extra)
	if err != nil {
		return nil, err
	}
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	for _, iID := range iIDs {
		pi := record.GetPackedInstruction(iID)
		bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
		cs.AddInstructionInSpilt(bID, Unpack(pi, record))
		//// 由于instruction变化，所以在这里需要重新映射stage内部的iID
		//sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
	}
	cs.CoeffTable = record.GetCoeffTable()
	fmt.Println("Compile Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())

	//fmt.Println(cs.Sit.GetStageNumber())
	return cs, nil
}
