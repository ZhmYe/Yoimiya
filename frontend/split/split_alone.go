package split

import (
	"Yoimiya/Config"
	"Yoimiya/Record"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"time"
)

/***

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// SplitAndProve 这里把sit和level用interface统一写成了splitEngine，在内部区分
func SplitAndProve(cs constraint.ConstraintSystem, assignment frontend.Circuit, cut int) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	extras := make([]constraint.ExtraValue, 0)
	//forwardOutput := make([]constraint.ExtraValue, 0)
	startTime := time.Now()
	fmt.Println("=================Start Split=================")
	var ibrs []constraint.IBR
	// todo record的改写
	//var record *DataRecord
	//var topIBR, bottomIBR constraint.IBR
	var commitment constraint.Commitments
	var coefftable cs_bn254.CoeffTable
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		_r1cs.SplitEngine.AssignLayer(cut)
		StructureRoundLog(_r1cs, 0)
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
	runtime.GC() //清理内存
	Config.Config.CancelSplit()
	for i, ibr := range ibrs {
		fmt.Print("	Sub Circuit ", i, " ")
		buildStartTime := time.Now()
		SubCs, err := buildConstraintSystemFromIBR(ibr,
			commitment, coefftable, pli, extras)
		if err != nil {
			panic(err)
		}
		Record.GlobalRecord.SetBuildTime(time.Since(buildStartTime))
		//GetSplitProof(SubCs, assignment, &extras, false)
		fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), extras, ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		proofs = append(proofs, ProveSplitWithWitness(SubCs,
			fullWitness,
			&extras, false))
		//proofs = append(proofs, GetSplitProof(SubCs, assignment, &extras, false))
		runtime.GC() //清理内存
	}
	//fmt.Print("	Top Circuit ")
	//buildStartTime := time.Now()
	//topCs, err := buildTopConstraintSystemFromIBR(topIBR,
	//	commitment, coefftable, pli)
	//if err != nil {
	//	panic(err)
	//}
	//GetSplitProof(topCs, assignment, &extras, false)
	//runtime.GC() //清理内存
	////testTime := time.Now()
	////for {
	////	if time.Since(testTime) >= time.Duration(20)*time.Second {
	////		break
	////	}
	////}
	////if err != nil {
	////	panic(err)
	////}
	////proofs = append(proofs, proof)
	//fmt.Print("	Bottom Circuit ")
	//buildStartTime = time.Now()
	//bottomCs, err := buildBottomConstraintSystemFromIBR(bottomIBR, coefftable,
	//	pli, extras)
	//Record.GlobalRecord.SetBuildTime(time.Since(buildStartTime))
	//GetSplitProof(bottomCs, assignment, &extras, false)
	//proofs = append(proofs, GetSplitProof(bottomCs, assignment, &extras, false)
	fmt.Println()
	fmt.Println("Total Time: ", time.Since(startTime))
	fmt.Println("=================Finish Split=================")
	return proofs, nil
}

// buildConstraintSystemFromIBR 构建子电路
func buildConstraintSystemFromIBR(ibr constraint.IBR,
	commitmentInfo constraint.Commitments, coeffTable cs_bn254.CoeffTable,
	pli frontend.PackedLeafInfo, extra []constraint.ExtraValue) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	//SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	cs.CommitmentInfo = commitmentInfo
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	cs.CoeffTable = coeffTable
	//fmt.Println("Length of Extra: ", len(extra))
	err := frontend.SetInputVariable(pli, ibr, cs, extra)
	if err != nil {
		panic(err)
	}
	for _, item := range ibr.Items() {
		bID := cs.AddBlueprint(item.BluePrint)
		cs.AddInstructionInSpilt(bID, item.CallData)
		if item.IsForwardOutput() {
			cs.AddForwardOutput(cs.GetNbInstructions() - 1) // iID = len(instruction) -1
		}
	}
	printConstraintSystemInfo(cs)
	return cs, nil
}

func printConstraintSystemInfo(cs *cs_bn254.R1CS) {
	fmt.Println("Compile Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}
