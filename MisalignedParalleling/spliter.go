package MisalignedParalleling

import (
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"fmt"
)

// 这里用来对某一个电路，进行切分，划分为n部分，返回n个cs
// 下半电路需要下半电路的extra，先不用extra的具体的值，在后面的函数里再进行赋值，先得到个数

type CsSpliter struct {
	cut int // 将电路划分为n部分，目前暂时默认是2
}

// Split 将传入的原生电路cs分割为cut个电路，通过PackedConstraintSystem封装
// 这里暂时就写Cut=2的逻辑 todo
func (b *CsSpliter) Split(cs constraint.ConstraintSystem, assignment frontend.Circuit) []PackedConstraintSystem {
	result := make([]PackedConstraintSystem, 0)
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SplitEngine.AssignLayer()
		split.StructureRoundLog(_r1cs, 0)
		top, bottom := _r1cs.SplitEngine.GetSubCircuitInstructionIDs()
		_r1cs.UpdateForwardOutput()                // 这里从原电路中获得middle对应的wireIDs
		forwardOutput := _r1cs.GetForwardOutputs() // 这里就是PackedConstraintSystem上半电路需要Set的forwardOutput
		var err error
		record := split.NewDataRecord(_r1cs)
		fmt.Print("	Top Circuit ")
		// 上半电路不需要extra
		topCs := buildPackedConstraintSystemFromIds(top, record, assignment, []constraint.ExtraValue{})
		if err != nil {
			panic(err)
		}
		// n != 2，那么只要在第一个电路调用该函数就可以了
		topCs.SetCommitment(record.GetCommitmentInfo())
		// 每次切分得到的上层电路都需要调用
		topCs.SetForwardOutput(forwardOutput)
		result = append(result, topCs)
		//fmt.Println("bottom=", len(bottom))
		fmt.Print("	Bottom Circuit ")
		// 下半电路的extra就是forwardOutput，这里先不考虑n刀，默认n=2
		bottomCs := buildPackedConstraintSystemFromIds(bottom, record, assignment, forwardOutput)
		if err != nil {
			panic(err)
		}
		result = append(result, bottomCs)
	default:
		panic("Only Support bn254 r1cs now...")
	}
	return result
}
func buildPackedConstraintSystemFromIds(iIDs []int, record *split.DataRecord,
	assignment frontend.Circuit, extra []constraint.ExtraValue) PackedConstraintSystem {
	cs, err := buildConstraintSystemFromIds(iIDs, record, assignment, extra)
	if err != nil {
		panic(err)
	}
	return *NewPackedConstraintSystem(&cs)
}
func buildConstraintSystemFromIds(iIDs []int, record *split.DataRecord,
	assignment frontend.Circuit, extra []constraint.ExtraValue) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	// 这部分放到外面去
	//if isTop {
	//	SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	//	cs.CommitmentInfo = record.GetCommitmentInfo()
	//}
	// 这里主要是添加Public和private变量，extra需要有具体的wireID
	pli := frontend.GetNbLeaf(assignment)
	err := frontend.SetNbLeaf(pli, cs, extra)
	if err != nil {
		return nil, err
	}
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	for _, iID := range iIDs {
		pi := record.GetPackedInstruction(iID)
		bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
		cs.AddInstructionInSpilt(bID, split.Unpack(pi, record))
	}
	cs.CoeffTable = record.GetCoeffTable()
	fmt.Println("Compile Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
	//fmt.Println(cs.Sit.GetStageNumber())
	return cs, nil
}
func NewSpliter(cut int) CsSpliter {
	return CsSpliter{cut: cut}
}
