package plugin

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
	"runtime"
	"time"
)

// Master 用于对给定的cs进行setup或者split+setup
type Master struct {
	split int
	//record plugin.PluginRecord
}

func NewMaster(s int) Master {
	return Master{split: s}
}

func (m *Master) Split(cs constraint.ConstraintSystem,
	assignment frontend.Circuit) ([]constraint.IBR, constraint.Commitments, cs_bn254.CoeffTable, frontend.PackedLeafInfo, time.Duration) {
	runtime.GC() //清理内存
	//forwardOutput := make([]constraint.ExtraValue, 0)
	var ibrs []constraint.IBR
	// todo record的改写
	//var record *DataRecord
	//var topIBR, bottomIBR constraint.IBR
	var commitment constraint.Commitments
	var coefftable cs_bn254.CoeffTable
	var splitTime time.Duration
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	runtime.GC() //清理内存
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		splitStartTime := time.Now()
		_r1cs.SplitEngine.AssignLayer(m.split)
		splitTime = time.Since(splitStartTime)
		runtime.GC()
		StructureRoundLog(_r1cs)
		//top, bottom = _r1cs.SplitEngine.GetSubCircuitInstructionIDs()
		//record = NewDataRecord(_r1cs)
		//topIBR = _r1cs.GetDataRecords(top)
		//bottomIBR = _r1cs.GetDataRecords(bottom)
		//topIBR, bottomIBR = _r1cs.GetDataRecords(top, bottom) // instruction-blueprint
		ibrs = _r1cs.GetDataRecords()
		commitment = _r1cs.CommitmentInfo
		coefftable = _r1cs.CoeffTable
		//_r1cs.UpdateForwardOutput() // 这里从原电路中获得middle对应的wireIDs
		//forwardOutput = _r1cs.GetForwardOutputs()
	default:
		panic("Only Support bn254 r1cs now...")
	}
	return ibrs, commitment, coefftable, pli, splitTime
}
func (m *Master) SetUp(cs constraint.ConstraintSystem) (groth16.ProvingKey, groth16.VerifyingKey, time.Duration) {
	startTime := time.Now()
	pk, vk, err := groth16.Setup(cs)
	if err != nil {
		panic(err)
	}
	return pk, vk, time.Since(startTime)
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
