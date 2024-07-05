package plugin

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
)

// PackedConstraintSystem 这里封装cs,并记录一些内部数据
type PackedConstraintSystem struct {
	cs      constraint.ConstraintSystem // 切分出的电路
	witness []int
}

func (p *PackedConstraintSystem) Witness() []int {
	return p.witness
}
func (p *PackedConstraintSystem) SetUp() (groth16.ProvingKey, groth16.VerifyingKey, error) {
	return groth16.Setup(p.cs)
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
func (p *PackedConstraintSystem) CS() constraint.ConstraintSystem { return p.cs }

// // Process 根据前面的电路计算结果extra，运行当前电路
//
//	func (p *PackedConstraintSystem) Process(extra []constraint.ExtraValue, assignment frontend.Circuit) split.PackedProof {
//		//startTime := time.Now()
//		pk, vk := frontend.SetUpSplit(p.cs)
//		//fmt.Println("Set Up Time: ", time.Since(startTime))
//		fullWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField())
//		publicWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
//		proof, err := groth16.Prove(p.cs, pk.(groth16.ProvingKey), fullWitness)
//		if err != nil {
//			panic(err)
//		}
//		//fmt.Println("Process Time: ", time.Since(startTime))
//		//var nextExtra []fr.Element
//		//nextForwardOutput, nextExtra := GetExtra(split)
//
//		// todo  这一段放在Task里生成,PackedConstraintSystem提供GetForwardOutput接口，返回计算后的结果
//
//		//if !isCluster {
//		//	newExtra := GetExtra(split)
//		//	//*extra = make([]any, 0)
//		//	for _, e := range newExtra {
//		//		*extra = append(*extra, e)
//		//	}
//		//	//for i, e := range *extra {
//		//	//	count, isUsed := usedExtra[e.GetWireID()]
//		//	//	if isUsed {
//		//	//		//e.Consume(count)
//		//	//		(*extra)[i].Consume(count)
//		//	//	}
//		//	//}
//		//}
//		return split.NewPackedProof(proof, vk, publicWitness)
//	}

func NewPackedConstraintSystem(cs constraint.ConstraintSystem) PackedConstraintSystem {
	witness := make([]int, cs.GetNbPublicVariables()+cs.GetNbSecretVariables()-1)
	for i := 0; i < len(witness); i++ {
		witness[i] = i + 1
	}
	return PackedConstraintSystem{cs: cs, witness: witness}
}

// BuildNewPackedConstraintSystem todo
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
		cs.AddInstructionInSpilt(bID, item.CallData, item.IsForwardOutput())
		//if item.IsForwardOutput() {
		//	cs.AddForwardOutputInstruction(cs.GetNbInstructions() - 1) // iID = len(instruction) -1
		//}
	}

	//newExtra := cs.GetForwardOutputs()
	//extra = append(extra, cs.GetForwardOutputs()...)
	return PackedConstraintSystem{cs: cs, witness: ibr.GetWitness()}, cs.GetForwardOutputs()
}