package _Pipeline

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
)

type PackedConstraintSystem struct {
	cs      constraint.ConstraintSystem
	pk      groth16.ProvingKey
	vk      groth16.VerifyingKey
	witness []int // 记录该电路哪些input是witness
}

func BuildNewPackedConstraintSystem(pli frontend.PackedLeafInfo, ibr constraint.IBR, extra []constraint.ExtraValue) (PackedConstraintSystem, []constraint.ExtraValue) {
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	//SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	// todo 这里的Commitment不能直接放到所有电路
	//cs.CommitmentInfo = commitmentInfo
	//cs.CoeffTable = coeffTable
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

	//newExtra := cs.GetForwardOutputs()
	//extra = append(extra, cs.GetForwardOutputs()...)
	return PackedConstraintSystem{cs: cs, witness: ibr.GetWitness()}, cs.GetForwardOutputs()
}
func (p *PackedConstraintSystem) CS() constraint.ConstraintSystem {
	return p.cs
}
func (p *PackedConstraintSystem) Witness() []int {
	return p.witness
}
func (p *PackedConstraintSystem) SetUp() {
	pk, vk, err := groth16.Setup(p.cs)
	if err != nil {
		panic(err)
	}
	p.pk, p.vk = pk, vk
}

// SetForwardOutput cs应该传递给下面的cs的内容
func (p *PackedConstraintSystem) SetForwardOutput(forwardOutput []constraint.ExtraValue) {
	switch _r1cs := p.cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SetForwardOutput(forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// SetCommitment 设置commitment，这里暂定就是第一个电路需要设置commitment
func (p *PackedConstraintSystem) SetCommitment(commitment constraint.Commitments) {
	switch _r1cs := p.cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.CommitmentInfo = commitment
	default:
		panic("Only Support bn254 r1cs now...")
	}
	//p.cs.CommitmentInfo = commitment
}
func (p *PackedConstraintSystem) SetCoeff(coeff cs_bn254.CoeffTable) {
	switch _r1cs := p.cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.CoeffTable = coeff
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// GetForwardOutput 该接口由Task调用，不断更新extra，然后将extra传入process接口
func (p *PackedConstraintSystem) GetForwardOutput() []constraint.ExtraValue {
	switch _r1cs := p.cs.(type) {
	case *cs_bn254.R1CS:
		//forwardOutput := _r1cs.GetForwardOutputs() // 获得更新后的forwardOutput，即middle output wireID
		//extra := _r1cs.GetForwardOutputs()
		//usedExtra := _r1cs.GetUsedExtra()
		//return _r1cs.GetForwardOutputs(), usedExtra
		return _r1cs.GetForwardOutputs()
	default:
		panic("Only Support bn254 r1cs now...")
	}
}
