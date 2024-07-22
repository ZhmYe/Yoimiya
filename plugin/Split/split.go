package Split

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"Yoimiya/plugin"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"strconv"
	"time"
)

type Groth16SplitRunner struct {
	record []plugin.PluginRecord
	//prover   Prover
	//verifier Verifier
	split int
}

func NewGroth16SplitRunner(s int) Groth16SplitRunner {
	return Groth16SplitRunner{
		record: make([]plugin.PluginRecord, 0), // 每个split记录一次
		//prover:   Prover{},
		//verifier: Verifier{},
		split: s,
	}
}

// Prepare 编译电路，并通过master将cs split为若干个部分
func (r *Groth16SplitRunner) Prepare(circuit Circuit.TestCircuit) ([]constraint.IBR, constraint.Commitments, cs_bn254.CoeffTable, frontend.PackedLeafInfo) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	originCircuitRecord := plugin.NewPluginRecord("Prepare")
	assignment := circuit.GetAssignment()
	Config.Config.SwitchToSplit()
	cs, compileTime := circuit.Compile()
	runtime.GC()
	originCircuitRecord.SetTime("Compile", compileTime)
	master := NewMaster(r.split)
	ibrs, commitment, coefftable, pli, splitTime := master.Split(cs, assignment)
	originCircuitRecord.SetTime("Split", splitTime)
	runtime.GC() //清理内存
	Config.Config.CancelSplit()
	r.record = append(r.record, originCircuitRecord)
	return ibrs, commitment, coefftable, pli
}

// Process 完成Prepare过程后，将各个split build为cs，然后通过master setup得到pk, vk
// 对每个cs进行prove得到proof
func (r *Groth16SplitRunner) Process(circuit Circuit.TestCircuit) ([]groth16.Proof, error) {
	//go r.record.MemoryMonitor()
	proofs := make([]groth16.Proof, 0)
	extras := make([]constraint.ExtraValue, 0)
	ibrs, commitment, coefftable, pli := r.Prepare(circuit)
	runtime.GC() //清理内存
	Config.Config.CancelSplit()
	for i, ibr := range ibrs {
		r.ProcessImpl(i, ibr, pli, commitment, coefftable, &extras)
		runtime.GC()
	}
	//r.record.Finish()
	//r.record.Print()
	return proofs, nil
}

func (r *Groth16SplitRunner) ProcessImpl(i int, ibr constraint.IBR, pli frontend.PackedLeafInfo, commitment constraint.Commitments, coefftable cs_bn254.CoeffTable, extras *[]constraint.ExtraValue) groth16.Proof {
	runtime.GOMAXPROCS(runtime.NumCPU())
	master := NewMaster(1)
	record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
	buildStartTime := time.Now()
	//fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
	fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
	//fmt.Println("	Witness Filled...")
	//fmt.Println("	Build Sub Circuit ", i, " From IBR...")
	SubCs, err := buildConstraintSystemFromIBR(ibr,
		commitment, coefftable, pli, *extras, "Sub Circuit"+strconv.Itoa(i))
	//fmt.Println("	Sub Circuit ", i, " Building Finish...")
	if err != nil {
		panic(err)
	}
	record.SetTime("Build", time.Since(buildStartTime))
	//startTime := time.Now()
	pk, _, setupTime := master.SetUp(SubCs)
	record.SetTime("SetUp", setupTime)
	runtime.GC()
	//publicWitness, _ := fullWitness.Public()
	go record.MemoryMonitor()
	prover := NewProver(pk)
	//verifier := NewVerifier(vk, publicWitness)
	runtime.GOMAXPROCS(16)
	switch _r1cs := SubCs.(type) {
	case *cs_bn254.R1CS:
		proveStartTime := time.Now()
		proof, err := prover.SolveAndProve(_r1cs, fullWitness)
		if err != nil {
			panic(err)
		}
		record.SetTime("Prove", time.Since(proveStartTime))
		newExtra := split.GetExtra(SubCs)
		//*extra = make([]any, 0)
		for _, e := range newExtra {
			newE := constraint.NewExtraValue(e.GetWireID())
			newE.SetValue(e.GetValue())
			*extras = append(*extras, newE)
		}
		runtime.GC() //清理内存
		//fmt.Println("Sub Circuit ", i, " Processing Finished...")
		record.Finish()
		r.record = append(r.record, record)
		return proof
	default:
		panic("Only Support BN254")
	}
	//r.prover.SolveAndProve(SubCs, fullWitness)
	//proof := split.ProveSplitWithWitness(SubCs,
	//	fullWitness,
	//	extras, false)
	// =================
	//proofs = append(proofs, ProveSplitWithWitness(SubCs,
	//	fullWitness,
	//	&extras, false))
	//proofs = append(proofs, GetSplitProof(SubCs, assignment, &extras, false))
}

// buildConstraintSystemFromIBR 构建子电路
func buildConstraintSystemFromIBR(ibr constraint.IBR,
	commitmentInfo constraint.Commitments, coeffTable cs_bn254.CoeffTable,
	pli frontend.PackedLeafInfo, extra []constraint.ExtraValue, name string) (constraint.ConstraintSystem, error) {
	// todo 核心逻辑
	// 这里根据切割返回出来的有序instruction ids，得到新的电路cs
	// record中记录了CallData、Blueprint、Instruction的map
	// CallData、Instruction应该是一一对应的关系，map取出后可删除
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	//SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	// todo 这里的Commitment不能直接放到所有电路
	cs.CommitmentInfo = commitmentInfo
	cs.CoeffTable = coeffTable
	err := frontend.SetInputVariable(pli, ibr, cs, extra)
	//fmt.Println("		nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	//fmt.Println("		nbExtra=", len(extra))
	if err != nil {
		panic(err)
	}
	for _, item := range ibr.Items() {
		bID := cs.AddBlueprint(item.BluePrint)
		cs.AddInstructionInSpilt(bID, item.ConstraintOffset, item.CallData, item.IsForwardOutput())
		//if item.IsForwardOutput() {
		//	cs.AddForwardOutputInstruction(cs.GetNbInstructions() - 1) // iID = len(instruction) -1
		//}
	}
	printConstraintSystemInfo(cs, name)
	return cs, nil
}
func printConstraintSystemInfo(cs *cs_bn254.R1CS, name string) {
	fmt.Println("[", name, "]", " Compile Result: ")
	fmt.Println("	NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("	NbConstraints=", cs.GetNbConstraints())
	fmt.Println("	NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}
func (r *Groth16SplitRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
	//r.record.Print()
}
