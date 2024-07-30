package plugin

import (
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
)

func PrintConstraintSystemInfo(cs *cs_bn254.R1CS, name string) {
	fmt.Println("[", name, "]", " Compile Result: ")
	fmt.Println("	NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("	NbConstraints=", cs.GetNbConstraints())
	fmt.Println("	NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}

// BuildConstraintSystemFromIBR 构建子电路
func BuildConstraintSystemFromIBR(ibr constraint.IBR,
	commitmentInfo constraint.Commitments, coeffTable cs_bn254.CoeffTable,
	pli frontend.PackedLeafInfo, extra []constraint.ExtraValue, name string) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	//SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	// todo 这里的Commitment不能直接放到所有电路
	cs.CommitmentInfo = commitmentInfo
	cs.CoeffTable = coeffTable
	err := frontend.SetInputVariable(pli, ibr, cs, extra)
	//fmt.Println("		nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	//fmt.Println("		nbExtra=", len(extra))
	if err != nil {
		panic(err)
	}
	for _, item := range ibr.Items() {
		bID := cs.AddBlueprint(item.BluePrint)
		cs.AddInstructionInSpilt(bID, item.ConstraintOffset, item.CallData, item.IsForwardOutput())
		//if item.IsForwardOutput() {
		//	cs.AddForwardOutputInstruction(cs.GetNbInstructions() - 1) // iID = len(instruction) -1
		//}
	}
	PrintConstraintSystemInfo(cs, name)
	return cs, nil
}
