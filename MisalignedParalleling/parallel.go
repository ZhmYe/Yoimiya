package MisalignedParalleling

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// 这里写n个电路错位并行的逻辑
// 当上一个电路的上半结束后，下一个电路的上半可以填充；同理，当上一个电路的下半结束后，如果下一个电路的下半可以填充
// 这里的并行包括setup和prove
// todo setup能否进一步并行，setup的时间占了运行的大部分
// todo 这里有效果的前提是，上下半分别setup+prove的时间小于原电路setup+prove的时间
// 在本地12核上
// SetUp时间
// 变量数    时间
// 2399207   2m29.1111371s 149s
// 1876269   1m9.4599315s  69s
// 2884420   3m20.4227137s 200s

type MParallelingMaster struct {
	Tasks  []*Task   // 所有任务
	slots  []*Slot   // 把每个cs看成是一个slot
	buffer []*Buffer // 每个Buffer针对每个slot，每个slot内可以添加不同task的process任务
}

func (m *MParallelingMaster) Initialize(nbTasks int, cut int, csGenerator func() constraint.ConstraintSystem, assignmentGenerator func() frontend.Circuit) {
	//assignmentGenerator1 := func() frontend.Circuit {
	//	assignment, _ := Circuit4VerifyCircuit.GetVerifyCircuitAssignment(Circuit4VerifyCircuit.GetVerifyCircuitParam())
	//	return assignment
	//}
	//csGenerator1 := func() constraint.ConstraintSystem {
	//	_, outerCircuit := Circuit4VerifyCircuit.GetVerifyCircuitAssignment(Circuit4VerifyCircuit.GetVerifyCircuitParam())
	//	return Circuit4VerifyCircuit.GetVerifyCircuitCs(outerCircuit)
	//}
	generator := NewGenerator(nbTasks)
	spliter := NewSpliter(cut)
	// 切分电路并生成不同的slot
	packedCss := spliter.Split(csGenerator(), assignmentGenerator())
	for i, packedCs := range packedCss {
		slot := NewSlot(i, packedCs)
		m.slots = append(m.slots, slot)
		m.buffer = append(m.buffer, NewBuffer(i))
	}
	m.Tasks = generator.generate(assignmentGenerator) // 生成指定数量的task,每个task包含属于自己的assignment即输入
	firstSlot := m.slots[0]
	for _, task := range m.Tasks {
		firstSlot.Push(task.id)
	}
	m.slots[0] = firstSlot
}

func (m *MParallelingMaster) Start() {
	// todo 这里的逻辑
	// 启动monitor用于debug
	go func() {
		m.monitor()
	}()
	// 启动slot对应的coordinator
	var wg4Slot sync.WaitGroup
	wg4Slot.Add(len(m.slots))
	for _, slot := range m.slots {
		tmp := slot.id
		go func(id int, wg *sync.WaitGroup) {
			m.coordinator(id, wg)
		}(tmp, &wg4Slot)
	}
	wg4Slot.Wait()
	flag := true
	for _, task := range m.Tasks {
		//task.Verify()
		for i, packedProof := range task.proofs {
			proof := packedProof.GetProof()
			verifyKey := packedProof.GetVerifyingKey()
			publicWitness := packedProof.GetPublicWitness()
			err := groth16.Verify(proof, verifyKey, publicWitness)
			if err != nil {
				flag = false
				fmt.Println("Task ", task.id, " Proof ", i, "Not Pass...", "Err: ", err)
			} else {
				fmt.Println("Task ", task.id, " Proof ", i, "Verify Success...")
			}
		}

	}
	if !flag {
		panic("Verify Failed!!!")
	}
}

// monitor 监控每个Task的进展
func (m *MParallelingMaster) monitor() {
	startTime := time.Now()
	for {
		if time.Since(startTime) >= time.Duration(1)*time.Minute {
			str := "Monitor Output\n"
			for _, task := range m.Tasks {
				str += strconv.Itoa(task.phase+1) + "/" + strconv.Itoa(len(m.slots)) + "\n"
			}
			fmt.Println(str)
			startTime = time.Now()
		}
	}
}

// coordinator 对某个buffer进行协调
func (m *MParallelingMaster) coordinator(id int, wg *sync.WaitGroup) {
	fmt.Println("Coordinator ", id, " Start...")
	finished := 0
	for {
		// 判断自己的buffer内是否有等待的task
		slot := m.slots[id]
		// 如果有需要处理的，则取出第一个元素
		if !slot.IsEmpty() {
			taskID := slot.Pop()
			task := m.Tasks[taskID]
			task.Process(id, slot.cs)
			// 处理完成后，将task放到下一个slot中
			m.Register(taskID, id+1)
			finished++
		}
		if finished == len(m.Tasks) {
			wg.Done()
			fmt.Println("Slot ", id, " Finished...")
			break
		}
	}
}
func (m *MParallelingMaster) Register(taskID int, slotID int) {
	if slotID >= len(m.slots) {
		return
	}
	slot := m.slots[slotID]
	slot.Push(taskID)
}
func NewMParallelingMaster() *MParallelingMaster {
	return &MParallelingMaster{
		Tasks:  make([]*Task, 0),
		slots:  make([]*Slot, 0),
		buffer: make([]*Buffer, 0),
	}
}
