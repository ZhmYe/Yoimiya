package SplitPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"runtime"
	"time"
)

type Groth16SerialRunner struct {
	tasks   []*Task
	record  []plugin.PluginRecord
	circuit Circuit.TestCircuit
}

func NewGroth16SerialRunner(circuit Circuit.TestCircuit) Groth16SerialRunner {
	return Groth16SerialRunner{circuit: circuit, tasks: make([]*Task, 0), record: make([]plugin.PluginRecord, 0)}
}
func (r *Groth16SerialRunner) InjectTasks(nbTask int) {
	for len(r.tasks) < nbTask {
		r.tasks = append(r.tasks, NewTask(r.circuit, 1, len(r.tasks)))
	}
}
func (r *Groth16SerialRunner) Prepare() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, []int) {
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
func (r *Groth16SerialRunner) Process() {
	pk, vk, cs, inputID := r.Prepare()
	record := plugin.NewPluginRecord("Process")
	go record.MemoryMonitor()
	startTime := time.Now()
	for {
		nbCommit := 0
		for _, task := range r.tasks {
			task.Process(pk, cs, inputID, vk)
			if !task.Next() {
				nbCommit++
			}
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
func (r *Groth16SerialRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
}
