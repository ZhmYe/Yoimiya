package split

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
)

// 这里存放各种split算法公用的一些函数和结构

// DataRecord 记录Instruction, CallData, BluePrint
// 为了切割完可以将原来的cs直接扔掉
// 这三部分数据需要被切割出的电路分别继承一部分，使用map比较方便
type DataRecord struct {
	Instructions   map[int]constraint.PackedInstruction
	CallData       map[int]uint32
	Blueprints     map[constraint.BlueprintID]constraint.Blueprint
	CoeffTable     cs_bn254.CoeffTable
	CommitmentInfo constraint.Commitments
}

func NewDataRecord(cs *cs_bn254.R1CS) *DataRecord {
	record := DataRecord{
		CallData:       make(map[int]uint32),
		Instructions:   make(map[int]constraint.PackedInstruction),
		Blueprints:     make(map[constraint.BlueprintID]constraint.Blueprint),
		CoeffTable:     cs_bn254.NewCoeffTable(0),
		CommitmentInfo: cs.CommitmentInfo,
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
	record.CommitmentInfo = cs.CommitmentInfo
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
func (r *DataRecord) GetCommitmentInfo() constraint.Commitments {
	return r.CommitmentInfo
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

// SplitMode
// NoSplit: 用于测试不切分电路下的内存使用情况
// Alone: 每切一次，生成一个电路并给出proof
// Cluster: 将切分出的多个电路合并为一个电路，给定最终电路的数量进行合并
type SplitMode int

const (
	Cluster SplitMode = iota
	Alone
	NoSplit
)

// Param 用于传递参数
type Param struct {
	mode  SplitMode
	cut   int
	debug bool
}

func NewParam(needSplit bool, isCluster bool, cut int, debug bool) Param {
	if !needSplit {
		return Param{mode: NoSplit, cut: -1, debug: debug}
	}
	if isCluster {
		return Param{mode: Cluster, cut: cut, debug: debug}
	}
	return Param{mode: Alone, cut: -1, debug: debug}
}
