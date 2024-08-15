package NormalPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/backend/witness"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"Yoimiya/plugin/Runner"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type SolverInput struct {
	tID     int
	phase   int
	witness witness.Witness
}
type SolverEngine struct {
	ccs        *cs.R1CS
	witnessID  []int
	solveLimit int
	pool       chan SolverInput
	numCPU     int // 最大核数
	records    []plugin.PluginRecord
	flag       bool
	circuit    Circuit.TestCircuit // 电路
	count      int                 // 任务计数
	tasks      []*Runner.Task      // 任务
	solveLock  chan int
	nbTask     int
}

// Prepare Solver只需要将circuit编译为ccs，得到witness ID即可
func (se *SolverEngine) Prepare() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Config.Config.CancelSplit()
	record := plugin.NewPluginRecord("Prepare")
	//master := plugin.NewMaster(1)
	assignment := se.circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	witnessID := make([]int, 0)
	for len(witnessID) < pli.NbPublic()+pli.NbSecret()-1 {
		witnessID = append(witnessID, len(witnessID)+1)
	}
	ccs, compileTime := se.circuit.Compile()
	plugin.PrintConstraintSystemInfo(ccs.(*cs.R1CS), se.circuit.Name())
	runtime.GC()
	record.SetTime("Compile", compileTime)
	//fmt.Println(compileTime)
	//setupTime := time.Now()
	//pk, vk, err := groth16.Setup(cs)
	//pk, vk, setupTime := master.SetUp(cs)
	//if err != nil {
	//	panic(err)
	//}
	runtime.GC()
	////go r.record.MemoryMonitor()
	//record.SetTime("SetUp", setupTime)
	record.Finish()
	se.records = append(se.records, record)
	se.ccs, se.witnessID = ccs.(*cs.R1CS), witnessID
}

// ServerImpl Solve监听Prover的运行情况，确认prover的setup已经运行完成，此时solver开始进行solve任务
func (se *SolverEngine) ServerImpl() {
	listener, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Solver Server listening on localhost:8081")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			return
		}
		//go func() {
		var message string
		if handleRequest(conn, &message) {
			fmt.Println("Prover Set Up Has Finish...")
			se.flag = true
			break
		}
		//}()
	}
}

// ClientImpl solver每完成一个任务的solve，就向prover发送请求使得其可以开始prove
func (se *SolverEngine) ClientImpl(tID int, phase int) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	message := strconv.Itoa(tID*1 + phase)
	conn.Write([]byte(message))
	//buffer := make([]byte, 1024)
	//n, err := conn.Read(buffer)
	//if err != nil {
	//	fmt.Println("Error reading:", err)
	//	return
	//}
	//fmt.Printf("Received: %s\n", string(buffer[:n]))
}
func (se *SolverEngine) InjectTasks() {
	for len(se.tasks) < se.nbTask {
		se.tasks = append(se.tasks, Runner.NewTask(se.circuit, 1, len(se.tasks)))
	}
}
func (se *SolverEngine) Start() {
	go se.ServerImpl()
	se.Prepare() // compile
	se.InjectTasks()
	// 等待prover setup完成
	for {
		if se.flag {
			break
		}
	}
	for _, task := range se.tasks {
		request := task.Params()
		witness, err := frontend.GenerateSplitWitnessFromPli(request.Pli, se.witnessID, request.Extra, ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		se.pool <- SolverInput{
			tID:     request.ID,
			phase:   request.Phase,
			witness: witness,
		}
	}
	close(se.pool)
	runtime.GOMAXPROCS(se.numCPU) // 设置solve的最大核数，默认为1 * 2 = 2个超线程
	startTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(se.tasks))
	for input := range se.pool {
		//var wg sync.WaitGroup
		//wg.Add(se.solveLimit)
		//for i := 0; i < se.solveLimit; i++ {
		tmp := input
		go func(input SolverInput) {
			//prover := plugin.NewProver(se.pk)
			//commitmentsInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.Solve(se.ccs, input.witness, pk.(*groth16_bn254.ProvingKey))
			// 这里就单纯的solve一下，简单起见prove那边就用统一的solution
			se.solveLock <- 1
			startTime := time.Now()
			groth16_bn254.SimpleSolve(se.ccs, input.witness)
			<-se.solveLock
			wg.Done()
			//if err != nil {
			//	panic(err)
			//}
			fmt.Printf("%d solveTime: %s\n", input.tID, time.Since(startTime))
			//publicW, err := input.witness.Public()
			//if err != nil {
			//	panic(err)
			//}
			se.ClientImpl(input.tID, input.phase)
			//wg.Done()
		}(tmp)
		//}
		//wg.Wait()
		se.count++
		//fmt.Println(input.tID, input.phase)
		//time.Sleep(time.Second)
	}
	wg.Wait()
	se.ClientImpl(-1*len(se.tasks), 0)
	fmt.Println(time.Since(startTime))
	time.Sleep(time.Minute)

}
func NewSolverEngine(circuit Circuit.TestCircuit, nbTask int, solveLimit int, nbCpu int) SolverEngine {
	//circuit := evaluate.GetCircuit(opt)
	return SolverEngine{
		ccs:        nil,
		witnessID:  make([]int, 0),
		solveLimit: solveLimit,
		pool:       make(chan SolverInput, 100),
		numCPU:     nbCpu,
		records:    make([]plugin.PluginRecord, 0),
		flag:       false,
		circuit:    circuit,
		count:      0,
		solveLock:  make(chan int, solveLimit),
		nbTask:     nbTask,
	}
}
