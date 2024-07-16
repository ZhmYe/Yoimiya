package Split

import (
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"fmt"
	"runtime"
	"time"
)

// Master 用于对给定的cs进行setup或者split+setup
type Master struct {
	split  int
	record plugin.PluginRecord
}

func NewMaster(s int) Master {
	return Master{split: s, record: plugin.NewPluginRecord()}
}

func (m *Master) Split(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]constraint.ConstraintSystem, error) {
	runtime.GC() //清理内存
	//forwardOutput := make([]constraint.ExtraValue, 0)
	var ibrs []constraint.IBR
	// todo record的改写
	//var record *DataRecord
	//var topIBR, bottomIBR constraint.IBR
	var commitment constraint.Commitments
	var coefftable cs_bn254.CoeffTable
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	runtime.GC() //清理内存
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		splitStartTime := time.Now()
		_r1cs.SplitEngine.AssignLayer(m.split)
		m.record.SetTime("Split", time.Since(splitStartTime))
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
