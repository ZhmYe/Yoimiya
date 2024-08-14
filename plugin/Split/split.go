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
	master := plugin.NewMaster(r.split)
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
	master := plugin.NewMaster(1)
	record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
	buildStartTime := time.Now()
	//fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
	fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
	//fmt.Println("	Witness Filled...")
	//fmt.Println("	Build Sub Circuit ", i, " From IBR...")
	SubCs, err := plugin.BuildConstraintSystemFromIBR(ibr,
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
	time.Sleep(time.Duration(10) * time.Second)
	go record.MemoryMonitor()
	prover := plugin.NewProver(pk)
	//verifier := NewVerifier(vk, publicWitness)
	//runtime.GOMAXPROCS(16)
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

func (r *Groth16SplitRunner) Record() []plugin.PluginRecord {
	for _, record := range r.record {
		record.Print()
	}
	return r.record
	//r.record.Print()
}
