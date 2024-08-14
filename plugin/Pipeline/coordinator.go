package Pipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Request struct {
	tID   int
	extra []constraint.ExtraValue
}

// Coordinator
// 在原本的Parallel中，一个Slot会被一个完整的电路填充，一共有maxParallel个Slot同时运行
// Pipeline中，将一个Slot中的任务，通过split的方式达到Misaligned Parallel的效果
type Coordinator struct {
	tasks       []frontend.PackedLeafInfo // 每个任务用它assignment所对应的packedLeafInfo表示
	nbSplit     int                       // split数量
	maxParallel int                       // slot数量
	pcs         []PackedConstraintSystem  // split电路cs与其对应的pk,vk
}

func NewCoordinator(s int, p int) Coordinator {
	return Coordinator{tasks: make([]frontend.PackedLeafInfo, 0), nbSplit: s, maxParallel: p, pcs: make([]PackedConstraintSystem, 0)}
}
func (c *Coordinator) Process(nbTask int, circuit Circuit.TestCircuit) plugin.PluginRecord {
	Config.Config.SwitchToSplit()
	record := plugin.NewPluginRecord("Process")
	go record.MemoryMonitor()
	cs, compileTime := circuit.Compile()
	record.SetTime("Compile", compileTime)
	c.injectTasks(nbTask, circuit)
	pli := frontend.GetPackedLeafInfoFromAssignment(circuit.GetAssignment()) // 随机的assignment用来获取pli
	c.Split(cs, pli)

	var wg sync.WaitGroup
	wg.Add(c.maxParallel)
	nbTasksPerSlot := len(c.tasks) / c.maxParallel
	indexes := make([][]int, 0)
	for i := 0; i < c.maxParallel; i++ {
		index := make([]int, 0)
		for j := i * nbTasksPerSlot; j < (i+1)*nbTasksPerSlot; j++ {
			if j < len(c.tasks) {
				index = append(index, j)
			}
		}
		indexes = append(indexes, index)
	}
	startTime := time.Now()
	for i := 0; i < c.maxParallel; i++ {
		go c.runSlot(i, indexes[i], &wg)
	}

	wg.Wait()
	record.SetTime("Prove", time.Since(startTime))
	return record
}
func (c *Coordinator) injectTasks(nbTask int, circuit Circuit.TestCircuit) {
	for len(c.tasks) < nbTask {
		assignment := circuit.GetAssignment()
		pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
		c.tasks = append(c.tasks, pli)
	}
}
func (c *Coordinator) Split(cs constraint.ConstraintSystem, pli frontend.PackedLeafInfo) {
	var ibrs []constraint.IBR
	var commitment constraint.Commitments
	var coefftable cs_bn254.CoeffTable
	var extras []constraint.ExtraValue
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		//splitStartTime := time.Now()
		_r1cs.SplitEngine.AssignLayer(c.nbSplit)
		runtime.GC()
		ibrs = _r1cs.GetDataRecords()
		commitment = _r1cs.CommitmentInfo
		coefftable = _r1cs.CoeffTable
	default:
		panic("Only Support bn254 r1cs now...")
	}
	runtime.GC() //清理内存
	Config.Config.CancelSplit()
	for i, ibr := range ibrs {
		pcs, newExtras := BuildNewPackedConstraintSystem(pli, ibr, extras)
		extras = append(extras, newExtras...)
		pcs.SetCommitment(commitment)
		pcs.SetCoeff(coefftable)
		printConstraintSystemInfo(pcs.CS(), "Sub Circuit "+strconv.Itoa(i))
		pcs.SetUp()
		c.pcs = append(c.pcs, pcs)
		runtime.GC()
	}
}
func printConstraintSystemInfo(cs constraint.ConstraintSystem, name string) {
	fmt.Println("	[Compile Result]: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), "NbSecret=", cs.GetNbSecretVariables(), "NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbConstraint=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}

func (c *Coordinator) runSlot(sID int, tIDs []int, slotWg *sync.WaitGroup) {
	// 针对每个split都开一个channel
	//fmt.Println(tIDs)
	channels := make([]chan Request, len(c.pcs))
	for i := range channels {
		channels[i] = make(chan Request, len(tIDs)) // 初始化每个channel
	}

	//firstRequest := Request{
	//	tID:   0,
	//	extra: make([]constraint.ExtraValue, 0),
	//}
	//channels[0] <- firstRequest // 发送第一个请求

	var wg sync.WaitGroup
	wg.Add(len(c.pcs))

	runner := func(rID int, channel <-chan Request) {
		//defer wg.Done()
		inputID := c.pcs[rID].Witness()
		cs := c.pcs[rID].CS()
		//printConstraintSystemInfo(cs, "")
		pk := c.pcs[rID].pk
		total := 0
		for request := range channel {

			//fmt.Println(sID, rID, tIDs[request.tID])
			tID, extra := request.tID, request.extra
			witness, err := frontend.GenerateSplitWitnessFromPli(c.tasks[tIDs[tID]], inputID, extra, ecc.BN254.ScalarField())
			tmp := time.Now()
			if err != nil {
				panic(err)
			}
			_, err = groth16.Prove(cs, pk, witness)
			if err != nil {
				panic(err)
			}
			fmt.Printf("run slot: %v \n", time.Since(tmp))
			GetExtra := func(system constraint.ConstraintSystem) []constraint.ExtraValue {
				switch _r1cs := system.(type) {
				case *cs_bn254.R1CS:
					return _r1cs.GetForwardOutputs()
				default:
					panic("Only Support bn254 r1cs now...")
				}
			}
			newExtra := GetExtra(cs)

			// 最后一个channel不需要做额外的操作，其它channel需要将当前任务传给下一个channel
			total++
			if rID != len(c.pcs)-1 {
				nextRequest := Request{
					tID:   tID,
					extra: append(extra, newExtra...),
				}
				channels[rID+1] <- nextRequest
				// 记录处理的请求数量，并在处理完所有请求后关闭下一个channel，因为不会再有输入
				if total == len(tIDs) {
					close(channels[rID+1])
				}
			}
			if total == len(tIDs) {
				wg.Done()
			}

		}
	}

	//for rID := 0; rID < len(c.pcs); rID++ {
	go runner(0, channels[0])
	//}
	go func() {
		defer close(channels[0])
		for i := 0; i < len(tIDs); i++ {
			request := Request{
				tID:   i,
				extra: make([]constraint.ExtraValue, 0),
			}
			channels[0] <- request
		}
	}()
	wg.Wait()
	slotWg.Done()
}
