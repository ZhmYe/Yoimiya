package NormalPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ProverInput struct {
	CommitmentInfo constraint.Groth16Commitments
	Solution       *cs.R1CSSolution
	NbPublic       int
	NbPrivate      int
	tID            int
	phase          int
	PublicWitness  witness.Witness
}
type SolverSolution struct {
	commitmentInfo constraint.Groth16Commitments
	solution       *cs.R1CSSolution
	nbPublic       int
	nbPrivate      int
}

// ProverEngine 负责生成proof以及setup得到pk,vk
type ProverEngine struct {
	pk           groth16.ProvingKey
	vk           groth16.VerifyingKey
	proveLimit   int
	pool         chan ProverInput
	numCPU       int // 最大核数
	records      []plugin.PluginRecord
	circuit      Circuit.TestCircuit
	fakeSolution SolverSolution
}

// Prepare Prover将circuit编译为ccs后，setup得到pk, vk
// todo Compiler+Splitor -> 得到ccs(pcs) + pk,vk，然后Solver和Prover读取
func (pe *ProverEngine) Prepare() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Config.Config.CancelSplit()
	record := plugin.NewPluginRecord("Prepare")
	master := plugin.NewMaster(1)
	assignment := pe.circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	witnessID := make([]int, 0)
	for len(witnessID) < pli.NbPublic()+pli.NbSecret()-1 {
		witnessID = append(witnessID, len(witnessID)+1)
	}
	ccs, compileTime := pe.circuit.Compile()
	plugin.PrintConstraintSystemInfo(ccs.(*cs.R1CS), pe.circuit.Name())
	runtime.GC()
	record.SetTime("Compile", compileTime)
	//fmt.Println(compileTime)
	//setupTime := time.Now()
	//pk, vk, err := groth16.Setup(cs)
	pk, vk, setupTime := master.SetUp(ccs)
	//if err != nil {
	//	panic(err)
	//}
	runtime.GC()
	////go r.record.MemoryMonitor()
	record.SetTime("SetUp", setupTime)
	record.Finish()
	pe.records = append(pe.records, record)
	//return pk, vk, cs, witnessID
	pe.pk, pe.vk = pk, vk
	//witnessID := make([]int, 0)
	//for len(witnessID) < pli.NbPublic()+pli.NbSecret()-1 {
	//	witnessID = append(witnessID, len(witnessID)+1)
	//}
	// todo 这里的solution可以不固定，通过网络发过来，测试时就用这个固定的solution
	pe.fakeSolution = fakeSolve(pe.circuit, ccs, witnessID, make([]constraint.ExtraValue, 0))
}

// ServerImpl Prover监听Solver的运行情况，一旦有新的solve任务完成，就将其添加到pool
func (pe *ProverEngine) ServerImpl() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Prover Server listening on localhost:8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			return
		}
		go func() {
			var message string
			if handleRequest(conn, &message) {
				fmt.Println("Add Solve Output to Pool...")
				tmp, _ := strconv.Atoi(message)
				pe.pool <- ProverInput{
					CommitmentInfo: pe.fakeSolution.commitmentInfo,
					Solution:       pe.fakeSolution.solution,
					NbPublic:       pe.fakeSolution.nbPublic,
					NbPrivate:      pe.fakeSolution.nbPrivate,
					tID:            tmp,
					phase:          0,
					PublicWitness:  nil,
				}
				if tmp == -1 {
					close(pe.pool)
				}
			}
		}()
	}
}

// ClientImpl prover一旦setup完成就向localhost:8081发送请求说明solver可以开始solve
func (pe *ProverEngine) ClientImpl() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	message := "Prover Set Up Finish!"
	conn.Write([]byte(message))
	//buffer := make([]byte, 1024)
	//n, err := conn.Read(buffer)
	//if err != nil {
	//	fmt.Println("Error reading:", err)
	//	return
	//}
	//fmt.Printf("Received: %s\n", string(buffer[:n]))
	fmt.Println("Waiting Solving...")
}
func (pe *ProverEngine) Start() {
	go pe.ServerImpl()
	pe.Prepare()                  // 得到pk, vk
	pe.ClientImpl()               // 向Solver说明set up 完成
	runtime.GOMAXPROCS(pe.numCPU) // 设置prove的最大核数
	startTime := time.Now()
	for input := range pe.pool {
		var wg sync.WaitGroup
		wg.Add(pe.proveLimit)
		for i := 0; i < pe.proveLimit; i++ {
			tmp := input
			go func(input ProverInput) {
				startTime := time.Now()
				groth16_bn254.GenerateZKP(input.CommitmentInfo, *input.Solution, pe.pk.(*groth16_bn254.ProvingKey), input.NbPublic, input.NbPrivate)
				//proof, err := groth16_bn254.GenerateZKP(input.CommitmentInfo, *input.Solution, pk.(*groth16_bn254.ProvingKey), input.NbPublic, input.NbPrivate)
				//if err != nil {
				//	panic(err)
				//}
				fmt.Printf("%d ProveTime: %s\n", input.tID, time.Since(startTime))
				//publicWitness, err := witness.Public()
				//if err != nil {
				//	panic(err)
				//}
				//*pe.output <- ProveOutput{
				//	proof: split.NewPackedProof(proof, vk, input.PublicWitness),
				//	tID:   input.tID,
				//	phase: input.phase,
				//}
				runtime.GC()
				wg.Done()
				//t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, input.PublicWitness))
			}(tmp)
		}
		wg.Wait()
	}
	fmt.Println(time.Since(startTime))
}
func NewProverEngine(circuit Circuit.TestCircuit) ProverEngine {
	return ProverEngine{
		//pk:         nil,
		//vk:         nil,
		proveLimit: 1,
		pool:       make(chan ProverInput, 100),
		numCPU:     runtime.NumCPU() - 2,
		records:    make([]plugin.PluginRecord, 0),
		circuit:    circuit,
	}
}
