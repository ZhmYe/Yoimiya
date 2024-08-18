package SplitPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"fmt"
	"runtime"
	"time"
)

type SpliterEngine struct {
	s      int                 // sub circuit的个数
	method Config.SPLIT_METHOD // 划分算法，这里暂时先不区分，在Config里区分 todo
}

// Split Spliter负责将circuit按照划分方式前置的编译方式编译为DAG后，将DAG split为若干个子图，然后分别组装为ccs
// 返回值: split的结果..., compile time, split time
func (s *SpliterEngine) Split(circuit Circuit.TestCircuit) ([]constraint.IBR,
	constraint.Commitments, cs.CoeffTable, frontend.PackedLeafInfo, time.Duration, time.Duration) {
	// Compile to ccs with DAG
	runtime.GOMAXPROCS(runtime.NumCPU())
	Config.Config.SwitchToSplit()
	ccs, compileTime := circuit.Compile()
	plugin.PrintConstraintSystemInfo(ccs.(*cs.R1CS), circuit.Name())
	runtime.GC()
	//record.SetTime("Compile", compileTime)
	ibrs, commitments, coefftable, pli, splitTime := s.split(ccs, circuit.GetAssignment())
	return ibrs, commitments, coefftable, pli, compileTime, splitTime

}
func (s *SpliterEngine) split(ccs constraint.ConstraintSystem,
	assignment frontend.Circuit) ([]constraint.IBR, constraint.Commitments, cs.CoeffTable, frontend.PackedLeafInfo, time.Duration) {
	runtime.GC() //清理内存
	//forwardOutput := make([]constraint.ExtraValue, 0)
	var ibrs []constraint.IBR
	// todo record的改写
	//var record *DataRecord
	//var topIBR, bottomIBR constraint.IBR
	var commitment constraint.Commitments
	var coefftable cs.CoeffTable
	var splitTime time.Duration
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	runtime.GC() //清理内存
	switch _r1cs := ccs.(type) {
	case *cs.R1CS:
		splitStartTime := time.Now()
		_r1cs.SplitEngine.AssignLayer(s.s)
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
func StructureRoundLog(_r1cs *cs.R1CS) {
	//fmt.Println("Round ", round)
	fmt.Println("Circuit Structure: ")
	fmt.Println("	Layers Number: ", _r1cs.SplitEngine.GetLayersInfo())
	fmt.Println("	Stage/LEVEL Number:", _r1cs.SplitEngine.GetStageNumber())
	fmt.Println("	Constraint Number: ", _r1cs.NbConstraints)
	fmt.Println("	NbPublic=", _r1cs.GetNbPublicVariables(), " NbSecret=", _r1cs.GetNbSecretVariables(), " NbInternal=", _r1cs.GetNbInternalVariables())
	fmt.Println("	Wire Number: ", _r1cs.GetNbPublicVariables()+_r1cs.GetNbSecretVariables()+_r1cs.GetNbInternalVariables())
}
func NewSplitor(s int) SpliterEngine {
	return SpliterEngine{
		s:      s,
		method: Config.SPLIT_LRO,
	}
}
