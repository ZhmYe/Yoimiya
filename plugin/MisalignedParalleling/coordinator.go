package MisalignedParalleling

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

// coordinator
// 有若干个task，不断的注入到coordinator中
// 每个task的电路一致，但assignment不同
// coordinator首先将电路compile完成得到各个split，然后分配到不同的slot中，并给出slot之间的运行关系
// 每当一个新的task进入，根据目前的slot情况分配task

type Coordinator struct {
	slot          []*Slot
	nbSplit       int // 将原电路分为多少个split, 如果nbSplit=1，那么无需split
	nbParallel    int // 最大并发粒度，也就是让多少个split并行，默认和nbSplit一样
	tasks         []*Task
	finish        bool
	memoryAlloc   uint64
	NumCPUPerSlot int
	//m           []uint64
}

func NewCoordinator(s int, p int) *Coordinator {
	if s <= 1 {
		s, p = 1, 1
	}
	return &Coordinator{
		nbParallel:    p,
		nbSplit:       s,
		slot:          make([]*Slot, 0),
		tasks:         make([]*Task, 0),
		finish:        false,
		memoryAlloc:   uint64(0),
		NumCPUPerSlot: runtime.NumCPU(),
		//m:           make([]uint64, 0),
	}
}
func (c *Coordinator) SetNumCPUPerSlot(n int) {
	c.NumCPUPerSlot = n
}
func (c *Coordinator) NeedSplit() bool {
	return c.nbSplit != 1
}
func (c *Coordinator) init(circuit Circuit.TestCircuit, nbTask int) {
	if c.NeedSplit() {
		Config.Config.SwitchToSplit()
	} else {
		Config.Config.CancelSplit()
	}
	cs, compileTime := circuit.Compile()
	fmt.Println("Coordinator Init Log")
	//fmt.Println("	[Circuit Name]: ", circuit.Name())
	fmt.Println("	[Split]: ", c.nbSplit)
	fmt.Println("	[Max Parallel]: ", c.nbParallel)
	fmt.Println("	[Compile Time]: ", compileTime)
	printConstraintSystemInfo(cs, circuit.Name())
	if !c.NeedSplit() {
		c.AddSlot(NewPackedConstraintSystem(cs))
	} else {
		// todo split逻辑
		var ibrs []constraint.IBR
		// todo record的改写
		var commitment constraint.Commitments
		var coefftable cs_bn254.CoeffTable
		var extras []constraint.ExtraValue
		pli := frontend.GetPackedLeafInfoFromAssignment(circuit.GetAssignment()) // 随机的assignment用来获取pli
		switch _r1cs := cs.(type) {
		case *cs_bn254.R1CS:
			splitStartTime := time.Now()
			_r1cs.SplitEngine.AssignLayer(c.nbSplit)
			fmt.Println("	[Split Time]: ", time.Since(splitStartTime))
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
			c.AddSlot(pcs)
		}
	}
	for _, slot := range c.slot {
		slot.Setup()
	}
	fmt.Println("	[Slot Number]: ", len(c.slot))

	// 生成task
	for len(c.tasks) < nbTask {
		task := NewTask(len(c.tasks), circuit.GetAssignment(), c.nbSplit)
		task.SetCoordinator(c)
		c.tasks = append(c.tasks, task)
	}
	runtime.GC()
}
func (c *Coordinator) AddSlot(pcs PackedConstraintSystem) {
	//pk, vk, err := pcs.SetUp()
	//if err != nil {
	//	panic(err)
	//}
	c.slot = append(c.slot,
		&Slot{
			//pk:     pk,
			//vk:     vk,
			pcs:    pcs,
			id:     len(c.slot),
			buffer: NewBuffer(len(c.slot), len(c.tasks)),
		})
	runtime.GC()
}
func (c *Coordinator) MemoryMonitor() {
	startTime := time.Now()
	// 这里可以看到整体内存趋势
	//memorySeq := make([]uint64, 0)
	for {
		if c.finish {
			//fmt.Println(memorySeq)
			break
		}
		if time.Since(startTime) >= time.Duration(10)*time.Millisecond {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			nowAlloc := m.Alloc
			if nowAlloc > c.memoryAlloc {
				c.memoryAlloc = nowAlloc
				//fmt.Println(nowAlloc)
			}
			//c.m = append(c.m, nowAlloc)
			startTime = time.Now()
		}
	}
}
func formatMemoryAlloc(m uint64) float64 {
	return float64(m) / 1024 / 1024 / 1024
}

// Process todo
// 处理一系列并发任务的逻辑
func (c *Coordinator) Process() []TaskReceipt {
	startTime := time.Now()
	runtime.GOMAXPROCS(c.NumCPUPerSlot * c.nbSplit)
	go c.MemoryMonitor()
	for _, task := range c.tasks {
		go task.Request()
	}
	for _, slot := range c.slot {
		go func(slot *Slot) {
			for {
				//if !slot.IsEmpty() {
				receipt := slot.Process()
				tID := receipt.tID
				//fmt.Println("Send Response To", tID, "Phase = ", slot.id)
				c.tasks[tID].HandleResponse(receipt)
				//}
				//if len(slot.buffer.items) == len(c.tasks) {
				//	receipts := slot.Process()
				//	for _, receipt := range receipts {
				//		tmp := receipt
				//		go func(receipt SlotReceipt) {
				//			tID := receipt.tID
				//			fmt.Println("Send Response To", tID, "Phase = ", slot.id)
				//			c.tasks[tID].HandleResponse(receipt)
				//		}(tmp)
				//	}
				//	break
				//}
			}
		}(slot)
	}
	taskReceipt := make([]TaskReceipt, len(c.tasks))
	for {
		count := 0
		for i, task := range c.tasks {
			if task.finish {
				count++
				taskReceipt[i] = task.receipt
			}
		}
		if count == len(c.tasks) {
			c.finish = true
			fmt.Println(time.Since(startTime))
			break
		}
	}
	fmt.Println(formatMemoryAlloc(c.memoryAlloc), "GB")
	//fmt.Println(c.m)
	return taskReceipt
}
func (c *Coordinator) HandleRequest(tID int, pli frontend.PackedLeafInfo, extra []constraint.ExtraValue, phase int) bool {
	//fmt.Println("Receive Request From ", tID, ", Phase = ", phase)
	slot := c.slot[phase]
	slot.HandleRequest(tID, pli, extra)
	return true
}
func printConstraintSystemInfo(cs constraint.ConstraintSystem, name string) {
	fmt.Println("	[Compile Result]: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), "NbSecret=", cs.GetNbSecretVariables(), "NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbConstraint=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}
