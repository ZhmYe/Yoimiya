package split

import (
	"Yoimiya/Record"
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"time"
)

// Split 将传入的电路(constraintSystem)切分为多份,统一接口
// 分为两种执行方式：合并(Cluster, 切分后根据传入的cut数合并)、分散(Alone, 每切出一个小电路就给出证明)
func Split(cs constraint.ConstraintSystem, assignment frontend.Circuit, cut int) ([]PackedProof, error) {
	if cut <= 1 {
		return SimpleProve(cs, assignment)
	}
	return SplitAndProve(cs, assignment, cut)
	//switch param.mode {
	//case Cluster:
	//	//return SplitAndCluster(cs, assignment, param.cut)
	//	return SplitAndCluster(cs, assignment, param.cut)
	//case Alone:
	//	return SplitAndProve(cs, assignment, param.cut)
	//case NoSplit:
	//	return SimpleProve(cs, assignment)
	//default:
	//	return SplitAndProve(cs, assignment, param.cut)
	//}
}

func StructureRoundLog(_r1cs *cs_bn254.R1CS) {
	//fmt.Println("Round ", round)
	fmt.Println("Circuit Structure: ")
	fmt.Println("	Layers Number: ", _r1cs.SplitEngine.GetLayersInfo())
	fmt.Println("	Stage/LEVEL Number:", _r1cs.SplitEngine.GetStageNumber())
	fmt.Println("	Constraint Number: ", _r1cs.NbConstraints)
	fmt.Println("	NbPublic=", _r1cs.GetNbPublicVariables(), " NbSecret=", _r1cs.GetNbSecretVariables(), " NbInternal=", _r1cs.GetNbInternalVariables())
	fmt.Println("	Wire Number: ", _r1cs.GetNbPublicVariables()+_r1cs.GetNbSecretVariables()+_r1cs.GetNbInternalVariables())
}

func GetExtra(system constraint.ConstraintSystem) []constraint.ExtraValue {
	switch _r1cs := system.(type) {
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
func SetForwardOutput(split constraint.ConstraintSystem, forwardOutput []constraint.ExtraValue) {
	switch _r1cs := split.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SetForwardOutput(forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// ProveSplitWithWitness 传入Split后的电路，进行prove，记录Prove时间和内存使用
func ProveSplitWithWitness(split constraint.ConstraintSystem,
	fullWitness witness.Witness, extra *[]constraint.ExtraValue, isCluster bool) PackedProof {
	//err := SetNbLeaf(assignment, &split)
	//if err != nil {
	//panic(err)
	//}
	startTime := time.Now()
	pk, vk := frontend.SetUpSplit(split)
	runtime.GC()
	publicWitness, _ := fullWitness.Public()
	Record.GlobalRecord.SetSetUpTime(time.Since(startTime))
	startTime = time.Now()
	proof, err := groth16.Prove(split, pk.(groth16.ProvingKey), fullWitness)
	fmt.Println("Prove Time", time.Since(startTime))
	Record.GlobalRecord.SetSolveTime(time.Since(startTime))
	if err != nil {
		panic(err)
	}
	//var nextExtra []fr.Element
	//nextForwardOutput, nextExtra := GetExtra(split)
	if !isCluster {
		newExtra := GetExtra(split)
		//*extra = make([]any, 0)
		for _, e := range newExtra {
			newE := constraint.NewExtraValue(e.GetWireID())
			newE.SetValue(e.GetValue())
			*extra = append(*extra, newE)
		}
	}
	return NewPackedProof(proof, vk, publicWitness)
}

// GetSplitProof 传入Split后的电路，进行prove，记录Prove时间和内存使用
// todo 内存使用记录
func GetSplitProof(split constraint.ConstraintSystem,
	assignment frontend.Circuit, extra *[]constraint.ExtraValue, isCluster bool) PackedProof {
	//err := SetNbLeaf(assignment, &split)
	//if err != nil {
	//panic(err)
	//}
	startTime := time.Now()
	pk, vk := frontend.SetUpSplit(split)
	//fullWitness, _ := frontend.GenerateSplitWitnessFromPli()
	fullWitness, _ := frontend.GenerateWitness(assignment, *extra, ecc.BN254.ScalarField())
	publicWitness, _ := frontend.GenerateWitness(assignment, *extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	Record.GlobalRecord.SetSetUpTime(time.Since(startTime))
	startTime = time.Now()
	proof, err := groth16.Prove(split, pk.(groth16.ProvingKey), fullWitness)
	Record.GlobalRecord.SetSolveTime(time.Since(startTime))
	if err != nil {
		panic(err)
	}
	//var nextExtra []fr.Element
	//nextForwardOutput, nextExtra := GetExtra(split)
	if !isCluster {
		newExtra := GetExtra(split)
		//*extra = make([]any, 0)
		for _, e := range newExtra {
			newE := constraint.NewExtraValue(e.GetWireID())
			newE.SetValue(e.GetValue())
			*extra = append(*extra, newE)
		}
	}
	return NewPackedProof(proof, vk, publicWitness)
}

// Unpack 用于提取PackedInstruction中的信息,返回callData
func Unpack(pi constraint.PackedInstruction, record *DataRecord) []uint32 {
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
