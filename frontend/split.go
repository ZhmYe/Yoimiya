package frontend

import (
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/graph"
	"fmt"
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
}

func NewDataRecord(cs *cs_bn254.R1CS) *DataRecord {
	record := DataRecord{
		CallData:     make(map[int]uint32),
		Instructions: make(map[int]constraint.PackedInstruction),
		Blueprints:   make(map[constraint.BlueprintID]constraint.Blueprint),
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

// Split 将传入的电路(constraintSystem)切分为多份，返回所有切出的子电路
func Split(cs constraint.ConstraintSystem) ([]constraint.ConstraintSystem, error) {
	fmt.Println("Enter Split Function...")
	split := make([]constraint.ConstraintSystem, 0)
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		sits, err := trySplit(_r1cs)
		if err != nil {
			panic(err)
		}
		// todo 这里等待实现
		record := NewDataRecord(_r1cs)
		// 这里是否可以直接将_r1cs置为nil
		for _, sit := range sits {
			// todo 这部分可能提到外面， Split返回[][]int
			subCs, err := buildConstraintSystemFromSit(sit, record)
			if err != nil {
				panic(err)
			}
			split = append(split, subCs)
		}
		return split, nil
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// 将传入的cs转化为新的多个电路内部对应的sit，同时返回所有的instruction
func trySplit(cs *cs_bn254.R1CS) ([]*graph.SITree, error) {
	fmt.Println("Enter trySplit Function...")
	result := make([]*graph.SITree, 0)
	// todo 这里等待实现
	top, bottom := cs.Sit.HeuristicSplit()
	result = append(result, top)
	if bottom != nil {
		result = append(result, bottom)
	}
	fmt.Println(len(result))
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
func buildConstraintSystemFromSit(sit *graph.SITree, record *DataRecord) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的结果sit，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	fmt.Println("Enter build Function...")
	opt := defaultCompileConfig()
	cs := cs_bn254.NewR1CS(opt.Capacity)
	for i, stage := range sit.GetStages() {
		for j, iID := range stage.GetInstructions() {
			pi := record.GetPackedInstruction(iID)
			bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
			// todo 这里有很多Nb，如NbConstraints，暂时不确定前面是否需要加入
			// todo，这里会有第二次构建SIT，可以删掉？或者前面给定的结果不需要是SIT?
			cs.AddInstruction(bID, unpack(pi, record))
			// 由于instruction变化，所以在这里需要重新映射stage内部的iID
			sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
		}
	}
	return cs, nil
}
