package NormalPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"Yoimiya/plugin/Runner"
	"Yoimiya/plugin/Runner/Component"
	"runtime"
	"time"
)

type YoimiyaPipelineRunner struct {
	tasks   []*Runner.Task
	record  []plugin.PluginRecord
	circuit Circuit.TestCircuit
}

func NewYoimiyaPipelineRunner(circuit Circuit.TestCircuit) YoimiyaPipelineRunner {
	return YoimiyaPipelineRunner{circuit: circuit, tasks: make([]*Runner.Task, 0), record: make([]plugin.PluginRecord, 0)}
}
func (r *YoimiyaPipelineRunner) InjectTasks(nbTask int) {
	for len(r.tasks) < nbTask {
		r.tasks = append(r.tasks, Runner.NewTask(r.circuit, 1, len(r.tasks)))
	}
}
func (r *YoimiyaPipelineRunner) Prepare() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, []int) {
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
func (r *YoimiyaPipelineRunner) Process() {
	pk, vk, cs, inputID := r.Prepare()
	coordinator := Component.NewCoordinator(cs, pk, vk, inputID, 1, 1)
	coordinator.Inject(r.tasks)
	record := plugin.NewPluginRecord("Normal _Pipeline")
	go record.MemoryMonitor()
	//var wg sync.WaitGroup
	//wg.Add(len(r.tasks))
	startTime := time.Now()
	coordinator.Process(r.tasks)
	for {
		nbCommit := 0
		for _, task := range r.tasks {
			if task.Done() {
				nbCommit++
			}
		}
		if nbCommit == len(r.tasks) {
			break
		}
	}
	//for _, task := range r.tasks {
	//	tmpTask := task
	//	//tmpCs, tmpPK, tmpVK, tmpID := cs, pk, vk, inputID
	//	go func(task *Task) {
	//		task.SyncProcess(pk, cs, inputID, vk, &r.solveLock, &r.proveLock, &nbCommit)
	//	}(tmpTask)
	//
	//}
	//for {
	//	//for _, task := range r.tasks {
	//	//	task.SyncProcess(pk, cs, inputID, vk, &r.solveLock, &r.proveLock, &nbCommit)
	//	//}
	//	if nbCommit == len(r.tasks) {
	//		break
	//	}
	//}
	processTime := time.Since(startTime)
	record.SetTime("Total", processTime)
	record.Finish()
	r.record = append(r.record, record)
}

func (r *YoimiyaPipelineRunner) Record() {
	for _, record := range r.record {
		record.Print()
	}
}
