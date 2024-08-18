package Serial

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"Yoimiya/plugin"
	"Yoimiya/plugin/Runner"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"time"
)

// todo
// 这里serial也加上split

type YoimiyaSerialRunner struct {
	tasks   []*Runner.Task
	record  []plugin.PluginRecord
	circuit Circuit.TestCircuit
}

func NewYoimiyaSerialRunner(circuit Circuit.TestCircuit) YoimiyaSerialRunner {
	return YoimiyaSerialRunner{circuit: circuit, tasks: make([]*Runner.Task, 0), record: make([]plugin.PluginRecord, 0)}
}
func (r *YoimiyaSerialRunner) InjectTasks(nbTask int) {
	for len(r.tasks) < nbTask {
		r.tasks = append(r.tasks, Runner.NewTask(r.circuit, 1, len(r.tasks)))
	}
}
func (r *YoimiyaSerialRunner) Prepare() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, []int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Config.Config.CancelSplit()
	record := plugin.NewPluginRecord("Prepare")
	master := plugin.NewMaster(1)
	assignment := r.circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	witnessID := make([]int, 0)
	for len(witnessID) < pli.NbPublic()+pli.NbSecret()-1 {
		witnessID = append(witnessID, len(witnessID)+1)
	}
	cs, compileTime := r.circuit.Compile()
	plugin.PrintConstraintSystemInfo(cs.(*cs_bn254.R1CS), r.circuit.Name())
	runtime.GC()
	record.SetTime("Compile", compileTime)
	//fmt.Println(compileTime)
	//setupTime := time.Now()
	//pk, vk, err := groth16.Setup(cs)
	pk, vk, setupTime := master.SetUp(cs)
	//if err != nil {
	//	panic(err)
	//}
	runtime.GC()
	////go r.record.MemoryMonitor()
	record.SetTime("SetUp", setupTime)
	record.Finish()
	r.record = append(r.record, record)
	return pk, vk, cs, witnessID
}
func (r *YoimiyaSerialRunner) Process() {
	pk, vk, ccs, inputID := r.Prepare()
	record := plugin.NewPluginRecord("Process")
	go record.MemoryMonitor()
	startTime := time.Now()
	processTask := func(t *Runner.Task) bool {
		if t.Done() {
			return false
		}
		runtime.GOMAXPROCS(runtime.NumCPU())
		//fmt.Println(inputID)
		//t.execLock.Lock()
		request := t.Params()
		witness, err := frontend.GenerateSplitWitnessFromPli(request.Pli, inputID, request.Extra, ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		prover := plugin.NewProver(pk)
		startTime := time.Now()
		proof, err := prover.SolveAndProve(ccs.(*cs_bn254.R1CS), witness)
		//t.execLock.Unlock()
		fmt.Println(time.Since(startTime))
		//proof, err := groth16.Prove(ccs, pk, witness)
		if err != nil {
			panic(err)
		}
		publicWitness, err := witness.Public()
		if err != nil {
			panic(err)
		}
		//t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, publicWitness))
		response := Runner.Response{ID: request.ID, Proof: split.NewPackedProof(proof, vk, publicWitness), Extra: make([]constraint.ExtraValue, 0)}
		return t.HandleResponse(response)
		//return t.Next()
	}
	for {
		nbCommit := 0
		for _, task := range r.tasks {
			if !processTask(task) {
				nbCommit++
			}
			////task.Process(pk, cs, inputID, vk)
			//if !task.Next() {
			//	nbCommit++
			//}
			runtime.GC()
		}
		if nbCommit == len(r.tasks) {
			break
		}
	}
	processTime := time.Since(startTime)
	record.SetTime("Total", processTime)
	record.Finish()
	r.record = append(r.record, record)
}
func (r *YoimiyaSerialRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
}
