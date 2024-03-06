package frontend

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/graph"
	"S-gnark/logger"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

/***

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// DataRecord 记录Instruction, CallData, BluePrint
// 为了切割完可以将原来的cs直接扔掉
// 这三部分数据需要被切割出的电路分别继承一部分，使用map比较方便
type DataRecord struct {
	Instructions map[int]constraint.PackedInstruction
	CallData     map[int]uint32
	Blueprints   map[constraint.BlueprintID]constraint.Blueprint
	CoeffTable   cs_bn254.CoeffTable
}

func NewDataRecord(cs *cs_bn254.R1CS) *DataRecord {
	record := DataRecord{
		CallData:     make(map[int]uint32),
		Instructions: make(map[int]constraint.PackedInstruction),
		Blueprints:   make(map[constraint.BlueprintID]constraint.Blueprint),
		CoeffTable:   cs_bn254.NewCoeffTable(0),
	}
	for i, data := range cs.CallData {
		record.CallData[i] = data
	}
	for i, instruction := range cs.Instructions {
		record.Instructions[i] = instruction
	}
	for i, blueprint := range cs.Blueprints {
		record.Blueprints[constraint.BlueprintID(i)] = blueprint
	}
	record.CoeffTable = cs.CoeffTable
	return &record
}

// todo get完是否可以直接删除

func (r *DataRecord) GetPackedInstruction(iId int) constraint.PackedInstruction {
	return r.Instructions[iId]
}
func (r *DataRecord) GetBluePrint(index constraint.BlueprintID) constraint.Blueprint {
	return r.Blueprints[index]
}
func (r *DataRecord) GetCallDatas(start int, end int) []uint32 {
	callData := make([]uint32, 0)
	for i := start; i <= end+1; i++ {
		callData = append(callData, r.CallData[i])
	}
	return callData
}
func (r *DataRecord) GetCallData(index int) uint32 {
	return r.CallData[index]
}
func (r *DataRecord) GetCoeffTable() cs_bn254.CoeffTable {
	return r.CoeffTable
}

// PackedProof 包含prove得到proof，以及验证需要的publicWitness、verifyingKey
type PackedProof struct {
	proof         groth16.Proof
	vk            groth16.VerifyingKey
	publicWitness witness.Witness
}

func NewPackedProof(proof groth16.Proof, vk groth16.VerifyingKey, w witness.Witness) PackedProof {
	return PackedProof{proof: proof, vk: vk, publicWitness: w}
}
func (p *PackedProof) GetProof() groth16.Proof {
	return p.proof
}
func (p *PackedProof) GetVerifyingKey() groth16.VerifyingKey {
	return p.vk
}
func (p *PackedProof) GetPublicWitness() witness.Witness {
	return p.publicWitness
}

// Split 将传入的电路(constraintSystem)切分为多份，返回所有切出的子电路
// todo 这里可以加入prove的逻辑
// todo 如果加入prove的逻辑，可以修改为返回proof以及verify需要的内容
func Split(cs constraint.ConstraintSystem, assignment Circuit) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	toSplitCs := cs
	flag := true
	hasSplit := false
	extra := make([]any, 0)
	forwardOutput := make([]int, 0)
	for {
		if !flag {
			fmt.Println("Recursive Split Finish...")
			break
		}
		switch _r1cs := toSplitCs.(type) {
		case *cs_bn254.R1CS:

			switch _r1cs.Sit.Examine() {
			case graph.PASS:
				log := logger.Logger()
				log.Debug().Str("SIT LAYER EXAMINE", "PASS").Msg("YZM DEBUG")
				fmt.Println("Examine PASS...")
			case graph.HAS_LINK:
				panic("Sit Layer Error: HAS_LINK...")
			case graph.LAYER_UNSET:
				panic("Sit Layer Error: LAYER_UNSET...")

			}
			fmt.Println(_r1cs.Sit.GetLayersInfo(), _r1cs.Sit.GetStageNumber(), _r1cs.Sit.GetTotalInstructionNumber())
			//sits, err := trySplit(_r1cs)
			top, bottom := _r1cs.Sit.CheckAndGetSubCircuitStageIDs()
			//if err != nil {
			//	panic(err)
			//}
			// todo 这里等待实现
			record := NewDataRecord(_r1cs)
			subCs, err := buildConstraintSystemFromIds(top, record, hasSplit, assignment)
			if err != nil {
				panic(err)
			}

			// 这里加入prove的逻辑，这样top可以丢弃 todo
			// todo 这里需要加入得到extra也就是MIDDLE的值的逻辑
			// todo any如何转化？
			proof := SplitAndProve(subCs, assignment, &extra, &forwardOutput)
			proofs = append(proofs, proof)
			if len(bottom) == 0 {
				flag = false
			} else {
				fmt.Println("bottom=", len(bottom))
				hasSplit = true
				toSplitCs, err = buildConstraintSystemFromIds(bottom, record, hasSplit, assignment)
				if err != nil {
					panic(err)
				}
			}
			// 这里是否可以直接将_r1cs置为nil
			//for _, sit := range sits {
			//	// todo 这部分可能提到外面， Split返回[][]int
			//	subCs, err := buildConstraintSystemFromSit(sit, record)
			//	if err != nil {
			//		panic(err)
			//	}
			//	split = append(split, subCs)
			//}
		default:
			panic("Only Support bn254 r1cs now...")
		}
	}
	return proofs, nil
}
func GetExtra(system constraint.ConstraintSystem) ([]int, []fr.Element) {
	switch _r1cs := system.(type) {
	case *cs_bn254.R1CS:
		forwardOutput := _r1cs.GetForwardOutputs() // 获得更新后的forwardOutput，即middle output wireID
		extra := _r1cs.GetExtra()
		return forwardOutput, extra
	default:
		panic("Only Support bn254 r1cs now...")
	}
}
func SetForwardOutput(split constraint.ConstraintSystem, forwardOutput []int) {
	switch _r1cs := split.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SetForwardOutput(forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// SplitAndProve 传入Split后的电路，进行prove，记录Prove时间和内存使用
// todo 内存使用记录
func SplitAndProve(split constraint.ConstraintSystem, assignment Circuit, extra *[]any, forwardOutput *[]int) PackedProof {
	//err := SetNbLeaf(assignment, &split)
	//if err != nil {
	//panic(err)
	//}
	pk, vk := SetUpSplit(split)
	fullWitness, _ := generateWitness(assignment, *extra, ecc.BN254.ScalarField())
	publicWitness, _ := generateWitness(assignment, *extra, ecc.BN254.ScalarField(), PublicOnly())
	SetForwardOutput(split, *forwardOutput)
	proof, _ := groth16.Prove(split, pk.(groth16.ProvingKey), fullWitness)
	var nextExtra []fr.Element
	*forwardOutput, nextExtra = GetExtra(split)
	*extra = make([]any, 0)
	for _, e := range nextExtra {
		*extra = append(*extra, e)
	}
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
	opt := defaultCompileConfig()
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
func buildConstraintSystemFromIds(iIDs []int, record *DataRecord, hasSplit bool, assignment Circuit) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := defaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	SetNbLeaf(assignment, cs)
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	for _, iID := range iIDs {
		pi := record.GetPackedInstruction(iID)
		bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
		// todo 这里的r1cs.Coefficients缺失
		if hasSplit {
			cs.AddInstructionInSpilt(bID, unpack(pi, record), true)
		} else {
			cs.AddInstructionInSpilt(bID, unpack(pi, record), false)
		}
		//// 由于instruction变化，所以在这里需要重新映射stage内部的iID
		//sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
	}
	cs.CoeffTable = record.GetCoeffTable()
	fmt.Println(cs.GetNbPublicVariables(), cs.GetNbSecretVariables(), cs.GetNbInternalVariables())
	fmt.Println(cs.GetNbCoefficients())
	return cs, nil
}
