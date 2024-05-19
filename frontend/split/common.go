package split

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
)

// 这里存放各种split算法公用的一些函数和结构

// DataRecord 记录Instruction, CallData, BluePrint
// 为了切割完可以将原来的cs直接扔掉
type DataRecord struct {
	Instructions   []constraint.PackedInstruction
	CallData       []uint32
	Blueprints     []constraint.Blueprint
	CoeffTable     cs_bn254.CoeffTable
	CommitmentInfo constraint.Commitments
}

func NewDataRecord(cs *cs_bn254.R1CS) *DataRecord {
	record := DataRecord{
		CallData:       make([]uint32, 0),
		Instructions:   make([]constraint.PackedInstruction, 0),
		Blueprints:     make([]constraint.Blueprint, 0),
		CoeffTable:     cs_bn254.NewCoeffTable(0),
		CommitmentInfo: cs.CommitmentInfo,
	}
	for _, data := range cs.CallData {
		record.CallData = append(record.CallData, data)
	}
	for _, instruction := range cs.Instructions {
		record.Instructions = append(record.Instructions, instruction)
	}
	for _, blueprint := range cs.Blueprints {
		record.Blueprints = append(record.Blueprints, blueprint)
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
	for i := start; i < end; i++ {
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
	return Param{mode: Alone, cut: cut, debug: debug}
}
