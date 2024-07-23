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
	"sync"
	"time"
)

type Groth16PipelineRunner struct {
	tasks     []*Task
	record    []plugin.PluginRecord
	circuit   Circuit.TestCircuit
	proveLock sync.Mutex
}

func NewGroth16PipelineRunner(circuit Circuit.TestCircuit) Groth16PipelineRunner {
	return Groth16PipelineRunner{tasks: make([]*Task, 0), circuit: circuit, record: make([]plugin.PluginRecord, 0)}
}
func (r *Groth16PipelineRunner) InjectTasks(nbTask int) {
	for len(r.tasks) < nbTask {
		r.tasks = append(r.tasks, NewTask(r.circuit, 1, len(r.tasks)))
	}
}
func (r *Groth16PipelineRunner) Prepare() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, []int) {
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
func (r *Groth16PipelineRunner) Process() {
	pk, vk, cs, inputID := r.Prepare()
	record := plugin.NewPluginRecord("Normal Pipeline")
	go record.MemoryMonitor()
	//var wg sync.WaitGroup
	//wg.Add(len(r.tasks))
	startTime := time.Now()
	var nbCommit int
	nbCommit = 0
	for {
		for _, task := range r.tasks {
			task.SyncProcess(pk, cs, inputID, vk, &r.proveLock, &nbCommit)
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
func (r *Groth16PipelineRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
}
