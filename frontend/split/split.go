package split

import (
	"S-gnark/backend/groth16"
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend"
	"S-gnark/graph"
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
	return NewPackedProof(proof, vk, publicWitness)
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
	fmt.Println("Compile Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
	//fmt.Println(cs.Sit.GetStageNumber())
	return cs, nil
}
