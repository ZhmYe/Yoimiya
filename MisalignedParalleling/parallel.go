package MisalignedParalleling

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
)

// 这里写n个电路错位并行的逻辑
// 当上一个电路的上半结束后，下一个电路的上半可以填充；同理，当上一个电路的下半结束后，如果下一个电路的下半可以填充
// 这里的并行包括setup和prove
// todo setup能否进一步并行，setup的时间占了运行的大部分

type MParallelingMaster struct {
	Tasks []*Task // 所有任务
	slots []*Slot // 把每个cs看成是一个slot
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
	}
	m.Tasks = generator.generate(assignmentGenerator) // 生成指定数量的task,每个task包含属于自己的assignment即输入

}

func (m *MParallelingMaster) Start() {
	// todo 这里的逻辑
}
func NewMParallelingMaster() *MParallelingMaster {
	return &MParallelingMaster{
		Tasks: make([]*Task, 0),
		slots: make([]*Slot, 0),
	}
}
