package SplitPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"runtime"
	"sync"
	"time"
)

type Groth16SplitPipelineRunner struct {
	tasks     []*Task
	record    []plugin.PluginRecord
	circuit   Circuit.TestCircuit
	proveLock sync.Mutex
	split     int
}

func NewGroth16SplitPipelineRunner(circuit Circuit.TestCircuit, s int) Groth16SplitPipelineRunner {
	return Groth16SplitPipelineRunner{
		tasks:   make([]*Task, 0),
		record:  make([]plugin.PluginRecord, 0),
		circuit: circuit,
		split:   s,
		//proveLock: sync.Mutex{},
	}
}

func (r *Groth16SplitPipelineRunner) Prepare() *PipelineConstraintSystem {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Config.Config.SwitchToSplit()
	record := plugin.NewPluginRecord("Prepare")
	//master := plugin.NewMaster(1)
	//assignment := r.circuit.GetAssignment()
	//pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	cs, compileTime := r.circuit.Compile()
	plugin.PrintConstraintSystemInfo(cs.(*cs_bn254.R1CS), r.circuit.Name())
	runtime.GC()
	record.SetTime("Compile", compileTime)
	ibrs, commitments, coefftable, pli, splitTime := r.Split(cs)
	record.SetTime("Split", splitTime)
	r.record = append(r.record, record)
	pcs, pipelineCsRecord := NewPipelineConstraintSystem(pli, ibrs, commitments, coefftable)
	r.record = append(r.record, pipelineCsRecord)
	return pcs
}

func (r *Groth16SplitPipelineRunner) Split(cs constraint.ConstraintSystem) ([]constraint.IBR,
	constraint.Commitments, cs_bn254.CoeffTable, frontend.PackedLeafInfo, time.Duration) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	master := plugin.NewMaster(r.split)
	assignment := r.circuit.GetAssignment()
	return master.Split(cs, assignment)

}

func (r *Groth16SplitPipelineRunner) Process() {
	// 先有prepare，得到可迭代的PipelineConstraintSystem里
	pcs := r.Prepare()
	record := plugin.NewPluginRecord("Split Pipeline")
	go record.MemoryMonitor()
	//var wg sync.WaitGroup
	//wg.Add(len(r.tasks))
	startTime := time.Now()
	var nbCommit int
	nbCommit = 0
	//for {
	//	if nbCommit == len(r.tasks) {
	//		break
	//	}
	for pcs.Next(&record) {
		cs, pk, vk, inputID := pcs.Params()
		for _, task := range r.tasks {
			task.SyncProcess(pk, cs, inputID, vk, &r.proveLock, &nbCommit)
		}
		runtime.GC()
	}
	//}
	for {
		if nbCommit == len(r.tasks) {
			break
		}
	}
	processTime := time.Since(startTime)
	record.SetTime("Total", processTime)
	record.Finish()
	r.record = append(r.record, record)

}
func (r *Groth16SplitPipelineRunner) InjectTasks(nbTask int) {
	for len(r.tasks) < nbTask {
		r.tasks = append(r.tasks, NewTask(r.circuit, 1, len(r.tasks)))
	}
}
func (r *Groth16SplitPipelineRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
}
