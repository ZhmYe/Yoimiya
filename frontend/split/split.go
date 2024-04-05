package split

import (
	"S-gnark/backend/groth16"
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
)

// Split 将传入的电路(constraintSystem)切分为多份,统一接口
// 分为两种执行方式：合并(Cluster, 切分后根据传入的cut数合并)、分散(Alone, 每切出一个小电路就给出证明)
func Split(cs constraint.ConstraintSystem, assignment frontend.Circuit, param Param) ([]PackedProof, error) {
	switch param.mode {
	case Cluster:
		//return SplitAndCluster(cs, assignment, param.cut)
		return SplitAndCluster(cs, assignment, param.cut)
	case Alone:
		return SplitAndProve(cs, assignment)
	case NoSplit:
		return SimpleProve(cs, assignment)
	default:
		return SplitAndProve(cs, assignment)
	}
}

func structureRoundLog(_r1cs *cs_bn254.R1CS, round int) {
	fmt.Println("Round ", round)
	fmt.Println("	Circuit Structure: ")
	fmt.Println("		Layers Number: ", _r1cs.Sit.GetLayersInfo())
	fmt.Println("		Stage Number:", _r1cs.Sit.GetStageNumber())
	fmt.Println("		Instruction Number: ", _r1cs.Sit.GetTotalInstructionNumber())
	fmt.Println("		NbPublic=", _r1cs.GetNbPublicVariables(), " NbSecret=", _r1cs.GetNbSecretVariables(), " NbInternal=", _r1cs.GetNbInternalVariables())
	fmt.Println("		Wire Number: ", _r1cs.GetNbPublicVariables()+_r1cs.GetNbSecretVariables()+_r1cs.GetNbInternalVariables())
}

func GetExtra(system constraint.ConstraintSystem) ([]constraint.ExtraValue, map[int]int) {
	switch _r1cs := system.(type) {
	case *cs_bn254.R1CS:
		//forwardOutput := _r1cs.GetForwardOutputs() // 获得更新后的forwardOutput，即middle output wireID
		//extra := _r1cs.GetForwardOutputs()
		usedExtra := _r1cs.GetUsedExtra()
		return _r1cs.GetForwardOutputs(), usedExtra
	default:
		panic("Only Support bn254 r1cs now...")
	}
}
func SetForwardOutput(split constraint.ConstraintSystem, forwardOutput []constraint.ExtraValue) {
	switch _r1cs := split.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SetForwardOutput(forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// GetSplitProof 传入Split后的电路，进行prove，记录Prove时间和内存使用
// todo 内存使用记录
func GetSplitProof(split constraint.ConstraintSystem,
	assignment frontend.Circuit, extra *[]constraint.ExtraValue, isCluster bool) PackedProof {
	//err := SetNbLeaf(assignment, &split)
	//if err != nil {
	//panic(err)
	//}
	pk, vk := frontend.SetUpSplit(split)
	fullWitness, _ := frontend.GenerateWitness(assignment, *extra, ecc.BN254.ScalarField())
	publicWitness, _ := frontend.GenerateWitness(assignment, *extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	proof, _ := groth16.Prove(split, pk.(groth16.ProvingKey), fullWitness)
	//var nextExtra []fr.Element
	//nextForwardOutput, nextExtra := GetExtra(split)
	if !isCluster {
		newExtra, usedExtra := GetExtra(split)
		//*extra = make([]any, 0)
		for _, e := range newExtra {
			*extra = append(*extra, e)
		}
		for i, e := range *extra {
			count, isUsed := usedExtra[e.GetWireID()]
			if isUsed {
				//e.Consume(count)
				(*extra)[i].Consume(count)
			}
		}
	}
	//*extra = newExtra
	//for _, f := range nextForwardOutput {
	//	*forwardOutput = append(*forwardOutput, f)
	//}
	//err := groth16.Verify(proof, vk, publicWitness)
	//if err != nil {
	//	panic(err)
	//} else {
	//	fmt.Println("PASS")
	//}
	return NewPackedProof(proof, vk, publicWitness)
}

// unpack 用于提取PackedInstruction中的信息,返回callData
func unpack(pi constraint.PackedInstruction, record *DataRecord) []uint32 {
	blueprint := record.GetBluePrint(pi.BlueprintID)
	cSize := blueprint.CalldataSize()
	if cSize < 0 {
		// by convention, we store nbInputs < 0 for non-static input length.
		cSize = int(record.GetCallData(int(pi.StartCallData)))
	}
	return record.GetCallDatas(int(pi.StartCallData), int(pi.StartCallData+uint64(cSize)))
}

// UnpackInstruction 将PackedInstruction还原为Instruction
func UnpackInstruction(pi constraint.PackedInstruction, record *DataRecord) constraint.Instruction {
	blueprint := record.GetBluePrint(pi.BlueprintID)
	cSize := blueprint.CalldataSize()
	if cSize < 0 {
		// by convention, we store nbInputs < 0 for non-static input length.
		cSize = int(record.GetCallData(int(pi.StartCallData)))
	}

	return constraint.Instruction{
		ConstraintOffset: pi.ConstraintOffset, // todo 这里不一定是原来的
		WireOffset:       pi.WireOffset,       // todo 这里不一定是原来的
		Calldata:         record.GetCallDatas(int(pi.StartCallData), int(pi.StartCallData+uint64(cSize))),
	}
}
